package domain_test

import (
	"testing"

	"coldchain-alert/internal/domain"
)

func TestEvaluateSeverityAndHealth(t *testing.T) {
	zone := domain.Zone{TargetMin: -22, TargetMax: -18, HumidityMin: 35, HumidityMax: 65}
	reading := domain.TemperatureReading{Temperature: -20, Humidity: 50}
	severity, reason := domain.EvaluateSeverity(zone, reading)
	if severity != "" || reason != "" {
		t.Fatalf("stable reading generated alert: %s %s", severity, reason)
	}
	reading.Temperature = -14
	severity, reason = domain.EvaluateSeverity(zone, reading)
	if severity != domain.SeverityWarning || reason == "" {
		t.Fatalf("warning reading was not classified: %s %s", severity, reason)
	}
	if score := domain.HealthScore(1, 1, 1, 2); score != 67 {
		t.Fatalf("unexpected health score %d", score)
	}
	if score := domain.HealthScore(0, 0, 0, 0); score != 0 {
		t.Fatalf("empty fleet should have zero score, got %d", score)
	}
}

func TestFilterAlertsAndSeriesHelpers(t *testing.T) {
	alerts := []domain.TemperatureAlert{{ID: "a", ZoneID: "z", Severity: domain.SeverityInfo, Status: domain.AlertOpen, OpenedAt: "1"}, {ID: "b", ZoneID: "z", Severity: domain.SeverityCritical, Status: domain.AlertResolved, OpenedAt: "2"}}
	filtered := domain.FilterAlerts(alerts, domain.AlertFilter{ZoneID: "z", Status: domain.AlertOpen})
	if len(filtered) != 1 || filtered[0].ID != "a" {
		t.Fatalf("unexpected alert filter: %#v", filtered)
	}
	points := []domain.SeriesPoint{{At: "1", Value: 1}, {At: "2", Value: 2}}
	if latest, ok := domain.LatestPoint(points); !ok || latest.At != "2" {
		t.Fatalf("latest point mismatch: %#v %v", latest, ok)
	}
	if domain.Average(points) != 1.5 || domain.IsReadingStable(points, 0, 2) != true {
		t.Fatal("series helper mismatch")
	}
}
