package store_test

import (
	"errors"
	"path/filepath"
	"testing"

	"coldchain-alert/internal/domain"
	"coldchain-alert/internal/store"
)

func TestPersistenceSurvivesReopen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "coldchain.db")
	storage, err := store.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	zone := domain.Zone{ID: "zone-reopen", Name: "Reopen Room", TargetMin: -22, TargetMax: -18, HumidityMin: 35, HumidityMax: 65, Enabled: true}
	device := domain.Device{ID: "device-reopen", ZoneID: zone.ID, Name: "Reopen Sensor", Model: "R1", Online: true, BatteryPct: 88}
	reading := domain.TemperatureReading{ID: "reading-reopen", DeviceID: device.ID, ZoneID: zone.ID, Temperature: -10, Humidity: 50, RecordedAt: "2026-01-01T01:00:00Z"}
	alert := domain.TemperatureAlert{ID: "alert-reopen", ZoneID: zone.ID, DeviceID: device.ID, ReadingID: reading.ID, Severity: domain.SeverityWarning, Status: domain.AlertOpen, Reason: "reopen check", OpenedAt: reading.RecordedAt, UpdatedAt: reading.RecordedAt}
	if err := storage.SaveZone(zone); err != nil {
		t.Fatal(err)
	}
	if err := storage.SaveDevice(device); err != nil {
		t.Fatal(err)
	}
	if err := storage.SaveReading(reading); err != nil {
		t.Fatal(err)
	}
	if err := storage.SaveAlert(alert); err != nil {
		t.Fatal(err)
	}
	if err := storage.Close(); err != nil {
		t.Fatal(err)
	}
	reopened, err := store.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	if got, err := reopened.GetZone(zone.ID); err != nil || got.Name != zone.Name {
		t.Fatalf("zone did not survive reopen: %#v %v", got, err)
	}
	if got, err := reopened.GetDevice(device.ID); err != nil || got.Model != device.Model {
		t.Fatalf("device did not survive reopen: %#v %v", got, err)
	}
	if got, err := reopened.GetReading(reading.ID); err != nil || got.Temperature != reading.Temperature {
		t.Fatalf("reading did not survive reopen: %#v %v", got, err)
	}
	if got, err := reopened.GetAlert(alert.ID); err != nil || got == nil || got.Reason != alert.Reason {
		t.Fatalf("alert did not survive reopen: %#v %v", got, err)
	}
}

func TestMissingAlertReturnsNilRecord(t *testing.T) {
	storage, err := store.Open(filepath.Join(t.TempDir(), "missing.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()
	alert, err := storage.GetAlert("absent")
	if err != nil || alert != nil {
		t.Fatalf("missing alert contract changed: %#v %v", alert, err)
	}
	if !errors.Is(storage.DeleteAlert("absent"), domain.ErrNotFound) {
		t.Fatal("deleting missing alert should report not found")
	}
}
