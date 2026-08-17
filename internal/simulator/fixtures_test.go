package simulator_test

import (
	"path/filepath"
	"testing"

	"coldchain-alert/internal/service"
	"coldchain-alert/internal/simulator"
	"coldchain-alert/internal/store"
)

func TestSeedIsDeterministic(t *testing.T) {
	storage, err := store.Open(filepath.Join(t.TempDir(), "seed.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()
	devices := service.NewDeviceService(storage)
	alerts := service.NewAlertService(storage)
	events := service.NewEventService(storage)
	audits := service.NewAuditService(storage)
	workflow := service.NewWorkflowService(devices, alerts, events, audits)
	fixture, err := simulator.Seed(workflow)
	if err != nil {
		t.Fatal(err)
	}
	readings, err := storage.ListReadingsByZone(fixture.Zone.ID)
	if err != nil || len(readings) != 24 {
		t.Fatalf("seed reading count mismatch: %d %v", len(readings), err)
	}
	alertsList, err := storage.ListAlertsByZone(fixture.Zone.ID)
	if err != nil || len(alertsList) != 1 {
		t.Fatalf("seed alert count mismatch: %d %v", len(alertsList), err)
	}
	if alertsList[0].ID != "alert-seed-reading-16" {
		t.Fatalf("seed alert id mismatch: %#v", alertsList[0])
	}
}
