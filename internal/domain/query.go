package domain

import "sort"

func SortAlerts(alerts []TemperatureAlert) {
	sort.SliceStable(alerts, func(i, j int) bool {
		if alerts[i].Severity.Weight() != alerts[j].Severity.Weight() {
			return alerts[i].Severity.Weight() > alerts[j].Severity.Weight()
		}
		return alerts[i].OpenedAt > alerts[j].OpenedAt
	})
}

func FilterAlerts(alerts []TemperatureAlert, filter AlertFilter) []TemperatureAlert {
	filtered := make([]TemperatureAlert, 0, len(alerts))
	for _, alert := range alerts {
		if filter.ZoneID != "" && alert.ZoneID != filter.ZoneID {
			continue
		}
		if filter.Status != "" && alert.Status != filter.Status {
			continue
		}
		if filter.Severity != "" && alert.Severity != filter.Severity {
			continue
		}
		filtered = append(filtered, alert)
	}
	SortAlerts(filtered)
	return filtered
}

func CountOpenAlerts(alerts []TemperatureAlert) (int, int) {
	openCount := 0
	criticalCount := 0
	for _, alert := range alerts {
		if IsOpen(alert.Status) {
			openCount++
			if alert.Severity == SeverityCritical {
				criticalCount++
			}
		}
	}
	return openCount, criticalCount
}
