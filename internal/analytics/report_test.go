package analytics_test

import (
	"testing"

	"coldchain-alert/internal/analytics"
	"coldchain-alert/internal/domain"
)

func TestComplianceReportAndRiskSignals(t *testing.T) {
	zone := domain.Zone{ID: "z", Name: "Room", TargetMin: -22, TargetMax: -18, HumidityMin: 35, HumidityMax: 65}
	readings := []domain.TemperatureReading{{Temperature: -20, Humidity: 50, RecordedAt: "2026-01-01T01:00:00Z"}, {Temperature: -10, Humidity: 50, RecordedAt: "2026-01-01T02:00:00Z"}}
	alerts := []domain.TemperatureAlert{{Severity: domain.SeverityCritical, Status: domain.AlertOpen, OpenedAt: "2026-01-01T02:00:00Z"}}
	doors := []domain.DoorEvent{{Opened: true, Duration: 90, RecordedAt: "2026-01-01T02:00:00Z"}}
	devices := []domain.Device{{Online: true}, {Online: false}}
	report := analytics.BuildComplianceReport(zone, readings, alerts, doors, devices, "2026-01-01T00:00:00Z", "2026-01-01T03:00:00Z")
	if report.ReadingCount != 2 || report.StableReadings != 1 || report.OpenAlerts != 1 || report.LongDoorEvents != 1 || report.Status != "attention" {
		t.Fatalf("report mismatch: %#v", report)
	}
	if len(report.Recommendations) != 4 {
		t.Fatalf("recommendations mismatch: %#v", report.Recommendations)
	}
	signals := analytics.BuildRiskSignals(zone, readings, alerts, devices, doors)
	if analytics.RiskBand(analytics.RiskScore(signals)) != "critical" || len(signals) < 4 {
		t.Fatalf("risk signals mismatch: %#v", signals)
	}
	if analytics.Explain(report) == "" || analytics.TemperatureRange(readings).Average != -15 {
		t.Fatal("report helpers returned empty values")
	}
}
