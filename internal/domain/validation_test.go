package domain_test

import (
	"errors"
	"testing"

	"coldchain-alert/internal/domain"
)

func TestValidateZoneAndDevice(t *testing.T) {
	zone := domain.Zone{ID: "zone-a", Name: "A", TargetMin: -22, TargetMax: -18, HumidityMin: 35, HumidityMax: 65, Enabled: true}
	if err := domain.ValidateZone(zone); err != nil {
		t.Fatalf("valid zone rejected: %v", err)
	}
	zone.TargetMax = -30
	if !errors.Is(domain.ValidateZone(zone), domain.ErrInvalidInput) {
		t.Fatal("reversed target range accepted")
	}
	device := domain.Device{ID: "device-a", ZoneID: "zone-a", Name: "Sensor", Model: "T1", BatteryPct: 101}
	if !errors.Is(domain.ValidateDevice(device), domain.ErrInvalidInput) {
		t.Fatal("invalid battery accepted")
	}
}

func TestValidateReadingAlertReviewAndDoor(t *testing.T) {
	reading := domain.TemperatureReading{ID: "r1", DeviceID: "d1", ZoneID: "z1", Temperature: -20, Humidity: 50, RecordedAt: "2026-01-01T00:00:00Z"}
	if err := domain.ValidateReading(reading); err != nil {
		t.Fatalf("reading rejected: %v", err)
	}
	alert := domain.TemperatureAlert{ID: "a1", ZoneID: "z1", DeviceID: "d1", ReadingID: "r1", Severity: domain.SeverityWarning, Status: domain.AlertOpen, Reason: "excursion", OpenedAt: reading.RecordedAt}
	if err := domain.ValidateAlert(alert); err != nil {
		t.Fatalf("alert rejected: %v", err)
	}
	review := domain.AlertReview{ID: "review-a", AlertID: "a1", Reviewer: "quality", Decision: "approve", ReviewedAt: reading.RecordedAt}
	if err := domain.ValidateReview(review); err != nil {
		t.Fatalf("review rejected: %v", err)
	}
	door := domain.DoorEvent{ID: "door-a", ZoneID: "z1", DeviceID: "d1", Opened: true, Duration: 15, RecordedAt: reading.RecordedAt}
	if err := domain.ValidateDoorEvent(door); err != nil {
		t.Fatalf("door rejected: %v", err)
	}
}
