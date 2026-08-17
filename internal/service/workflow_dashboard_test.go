package service_test

import (
	"testing"

	"coldchain-alert/internal/domain"
)

func TestWorkflowDoorEventsAndDashboard(t *testing.T) {
	storage, workflow, _, dashboard, events := newWorkflow(t)
	defer storage.Close()
	zone, device := registerFixture(t, workflow)
	for index := 0; index < 6; index++ {
		reading := domain.TemperatureReading{ID: "reading-dashboard-" + string(rune('a'+index)), DeviceID: device.ID, ZoneID: zone.ID, Temperature: -20 + float64(index%2), Humidity: 45 + float64(index), RecordedAt: "2026-01-01T0" + string(rune('1'+index)) + ":00:00Z"}
		if _, err := workflow.CaptureAndEvaluate(reading, "warehouse", reading.RecordedAt); err != nil {
			t.Fatal(err)
		}
	}
	eventsInput := []domain.DoorEvent{{ID: "door-dashboard-1", ZoneID: zone.ID, DeviceID: device.ID, Opened: true, Duration: 80, RecordedAt: "2026-01-01T07:00:00Z"}, {ID: "door-dashboard-2", ZoneID: zone.ID, DeviceID: device.ID, Opened: false, Duration: 5, RecordedAt: "2026-01-01T08:00:00Z"}}
	for _, event := range eventsInput {
		if err := workflow.RecordDoorAndAudit(event, "warehouse", event.RecordedAt); err != nil {
			t.Fatal(err)
		}
	}
	open, longOpen, err := events.DoorSummary(zone.ID)
	if err != nil || open != 1 || longOpen != 1 {
		t.Fatalf("door summary mismatch: %d %d %v", open, longOpen, err)
	}
	summary, err := dashboard.Summary(zone.ID, "2026-01-01T00:00:00Z")
	if err != nil {
		t.Fatal(err)
	}
	if len(summary.TemperatureSeries) != 6 || summary.DoorOpenCount != 1 || summary.OnlineDevices != 1 || summary.HealthScore != 100 {
		t.Fatalf("dashboard mismatch: %#v", summary)
	}
}
