package analytics

import (
	"sort"

	"coldchain-alert/internal/domain"
)

func BuildTemperatureSeries(readings []domain.TemperatureReading, limit int) []domain.SeriesPoint {
	points := make([]domain.SeriesPoint, 0, len(readings))
	ordered := append([]domain.TemperatureReading(nil), readings...)
	sort.SliceStable(ordered, func(i, j int) bool { return ordered[i].RecordedAt < ordered[j].RecordedAt })
	for _, reading := range ordered {
		points = append(points, domain.SeriesPoint{At: reading.RecordedAt, Value: reading.Temperature})
	}
	return trimPoints(points, limit)
}

func BuildHumiditySeries(readings []domain.TemperatureReading, limit int) []domain.SeriesPoint {
	points := make([]domain.SeriesPoint, 0, len(readings))
	ordered := append([]domain.TemperatureReading(nil), readings...)
	sort.SliceStable(ordered, func(i, j int) bool { return ordered[i].RecordedAt < ordered[j].RecordedAt })
	for _, reading := range ordered {
		points = append(points, domain.SeriesPoint{At: reading.RecordedAt, Value: reading.Humidity})
	}
	return trimPoints(points, limit)
}

func trimPoints(points []domain.SeriesPoint, limit int) []domain.SeriesPoint {
	if limit <= 0 || len(points) <= limit {
		return points
	}
	return append([]domain.SeriesPoint(nil), points[len(points)-limit:]...)
}

func CountOnline(devices []domain.Device) (int, int) {
	online := 0
	for _, device := range devices {
		if device.Online {
			online++
		}
	}
	return online, len(devices)
}

func CountDoorOpen(events []domain.DoorEvent) int {
	count := 0
	for _, event := range events {
		if event.Opened {
			count++
		}
	}
	return count
}

func AlertBreakdown(alerts []domain.TemperatureAlert) map[string]int {
	breakdown := map[string]int{
		string(domain.SeverityInfo):     0,
		string(domain.SeverityWarning):  0,
		string(domain.SeverityCritical): 0,
	}
	for _, alert := range alerts {
		breakdown[string(alert.Severity)]++
	}
	return breakdown
}

func Score(zone domain.Zone, alerts []domain.TemperatureAlert, devices []domain.Device) int {
	open, critical := domain.CountOpenAlerts(alerts)
	online, total := CountOnline(devices)
	return domain.HealthScore(open, critical, total-online, total)
}

func Trend(points []domain.SeriesPoint) string {
	if len(points) < 2 {
		return "flat"
	}
	first := points[0].Value
	last := points[len(points)-1].Value
	delta := last - first
	if delta > 0.5 {
		return "rising"
	}
	if delta < -0.5 {
		return "falling"
	}
	return "flat"
}
