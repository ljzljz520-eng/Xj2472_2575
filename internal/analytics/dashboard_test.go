package analytics_test

import (
	"testing"

	"coldchain-alert/internal/analytics"
	"coldchain-alert/internal/domain"
)

func TestDashboardSeriesAndAggregates(t *testing.T) {
	readings := []domain.TemperatureReading{{ZoneID: "z", Temperature: -20, Humidity: 45, RecordedAt: "2026-01-01T02:00:00Z"}, {ZoneID: "z", Temperature: -19, Humidity: 50, RecordedAt: "2026-01-01T01:00:00Z"}}
	series := analytics.BuildTemperatureSeries(readings, 24)
	if len(series) != 2 || series[0].Value != -19 {
		t.Fatalf("series ordering failed: %#v", series)
	}
	alerts := []domain.TemperatureAlert{{ZoneID: "z", Severity: domain.SeverityCritical, Status: domain.AlertOpen}}
	devices := []domain.Device{{ZoneID: "z", Online: true}, {ZoneID: "z", Online: false}}
	aggregates := analytics.AggregateZones([]domain.Zone{{ID: "z"}}, readings, alerts, nil)
	if len(aggregates) != 1 || aggregates[0].OpenAlerts != 1 || aggregates[0].Readings != 2 {
		t.Fatalf("aggregate mismatch: %#v", aggregates)
	}
	if analytics.Score(domain.Zone{ID: "z"}, alerts, devices) != 67 {
		t.Fatalf("unexpected dashboard score")
	}
	if analytics.Trend(series) != "falling" {
		t.Fatalf("unexpected trend %s", analytics.Trend(series))
	}
}

func TestRecentWindowAndBreakdown(t *testing.T) {
	readings := []domain.TemperatureReading{{ID: "old", RecordedAt: "2026-01-01T00:00:00Z"}, {ID: "new", RecordedAt: "2026-01-01T12:00:00Z"}}
	if recent := analytics.RecentReadings(readings, "2026-01-01T01:00:00Z", 0); len(recent) != 1 || recent[0].ID != "new" {
		t.Fatalf("window mismatch: %#v", recent)
	}
	breakdown := analytics.AlertBreakdown([]domain.TemperatureAlert{{Severity: domain.SeverityInfo}, {Severity: domain.SeverityCritical}})
	if breakdown["info"] != 1 || breakdown["critical"] != 1 {
		t.Fatalf("breakdown mismatch: %#v", breakdown)
	}
}
