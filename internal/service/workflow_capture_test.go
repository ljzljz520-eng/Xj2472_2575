package service_test

import (
	"path/filepath"
	"testing"

	"coldchain-alert/internal/domain"
	"coldchain-alert/internal/service"
	"coldchain-alert/internal/store"
)

func newWorkflow(t *testing.T) (*store.Store, *service.WorkflowService, *service.AlertService, *service.DashboardService, *service.EventService) {
	t.Helper()
	storage, err := store.Open(filepath.Join(t.TempDir(), "workflow.db"))
	if err != nil {
		t.Fatal(err)
	}
	devices := service.NewDeviceService(storage)
	alerts := service.NewAlertService(storage)
	events := service.NewEventService(storage)
	audits := service.NewAuditService(storage)
	return storage, service.NewWorkflowService(devices, alerts, events, audits), alerts, service.NewDashboardService(storage), events
}

func registerFixture(t *testing.T, workflow *service.WorkflowService) (domain.Zone, domain.Device) {
	t.Helper()
	zone := domain.Zone{ID: "zone-workflow", Name: "Workflow Room", TargetMin: -22, TargetMax: -18, HumidityMin: 35, HumidityMax: 65, Enabled: true}
	device := domain.Device{ID: "device-workflow", ZoneID: zone.ID, Name: "Workflow Sensor", Model: "W1", Online: true, BatteryPct: 90}
	if err := workflow.RegisterColdRoom(zone, device, "warehouse", "2026-01-01T00:00:00Z"); err != nil {
		t.Fatal(err)
	}
	return zone, device
}

func TestWorkflowCaptureCreatesTemperatureAlert(t *testing.T) {
	storage, workflow, alerts, _, _ := newWorkflow(t)
	defer storage.Close()
	zone, device := registerFixture(t, workflow)
	reading := domain.TemperatureReading{ID: "reading-critical", DeviceID: device.ID, ZoneID: zone.ID, Temperature: -8, Humidity: 52, RecordedAt: "2026-01-01T02:00:00Z"}
	alert, err := workflow.CaptureAndEvaluate(reading, "warehouse", reading.RecordedAt)
	if err != nil || alert == nil {
		t.Fatalf("critical reading did not create alert: %#v %v", alert, err)
	}
	if alert.Severity != domain.SeverityCritical || alert.Status != domain.AlertOpen {
		t.Fatalf("unexpected alert state: %#v", alert)
	}
	listed, err := alerts.ListAlerts(domain.AlertFilter{ZoneID: zone.ID})
	if err != nil || len(listed) != 1 {
		t.Fatalf("alert query mismatch: %#v %v", listed, err)
	}
}
