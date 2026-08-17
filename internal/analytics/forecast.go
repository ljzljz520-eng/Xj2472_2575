package analytics

import (
	"sort"

	"coldchain-alert/internal/domain"
)

type RiskSignal struct {
	Code     string  `json:"code"`
	Severity string  `json:"severity"`
	Score    float64 `json:"score"`
	Message  string  `json:"message"`
}

func BuildRiskSignals(zone domain.Zone, readings []domain.TemperatureReading, alerts []domain.TemperatureAlert, devices []domain.Device, doors []domain.DoorEvent) []RiskSignal {
	signals := make([]RiskSignal, 0, 5)
	if len(readings) > 0 {
		last := readings[len(readings)-1]
		severity, reason := domain.EvaluateSeverity(zone, last)
		if severity != "" {
			score := float64(severity.Weight() * 30)
			signals = append(signals, RiskSignal{Code: "temperature_excursion", Severity: string(severity), Score: score, Message: reason})
		}
	}
	open, critical := domain.CountOpenAlerts(alerts)
	if open > 0 {
		signals = append(signals, RiskSignal{Code: "open_alerts", Severity: "warning", Score: float64(open * 10), Message: "open alerts require an owner"})
	}
	if critical > 0 {
		signals = append(signals, RiskSignal{Code: "critical_alerts", Severity: "critical", Score: float64(critical * 25), Message: "critical alerts require quality review"})
	}
	online, total := CountOnline(devices)
	if total > 0 && online < total {
		signals = append(signals, RiskSignal{Code: "offline_devices", Severity: "warning", Score: float64((total - online) * 12), Message: "some sensors are offline"})
	}
	longDoors := 0
	for _, event := range doors {
		if event.Duration >= 60 {
			longDoors++
		}
	}
	if longDoors > 0 {
		signals = append(signals, RiskSignal{Code: "long_door_events", Severity: "warning", Score: float64(longDoors * 8), Message: "loading doors stayed open too long"})
	}
	sort.SliceStable(signals, func(i, j int) bool { return signals[i].Score > signals[j].Score })
	return signals
}

func RiskScore(signals []RiskSignal) float64 {
	score := 0.0
	for _, signal := range signals {
		score += signal.Score
	}
	if score > 100 {
		return 100
	}
	return score
}

func RiskBand(score float64) string {
	if score >= 70 {
		return "critical"
	}
	if score >= 35 {
		return "elevated"
	}
	if score > 0 {
		return "watch"
	}
	return "normal"
}

func Baseline(points []domain.SeriesPoint) float64 {
	if len(points) == 0 {
		return 0
	}
	return domain.Average(points)
}

func Deviation(points []domain.SeriesPoint, baseline float64) float64 {
	if len(points) == 0 {
		return 0
	}
	total := 0.0
	for _, point := range points {
		total += point.Value - baseline
	}
	return total / float64(len(points))
}
