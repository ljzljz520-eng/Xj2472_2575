package analytics

import (
	"sort"

	"coldchain-alert/internal/domain"
)

func RecentReadings(readings []domain.TemperatureReading, boundary string, limit int) []domain.TemperatureReading {
	filtered := make([]domain.TemperatureReading, 0, len(readings))
	for _, reading := range readings {
		if reading.RecordedAt >= boundary {
			filtered = append(filtered, reading)
		}
	}
	sort.SliceStable(filtered, func(i, j int) bool { return filtered[i].RecordedAt < filtered[j].RecordedAt })
	if limit > 0 && len(filtered) > limit {
		filtered = filtered[len(filtered)-limit:]
	}
	return filtered
}

func RecentDoors(events []domain.DoorEvent, boundary string) []domain.DoorEvent {
	filtered := make([]domain.DoorEvent, 0, len(events))
	for _, event := range events {
		if event.RecordedAt >= boundary {
			filtered = append(filtered, event)
		}
	}
	sort.SliceStable(filtered, func(i, j int) bool { return filtered[i].RecordedAt < filtered[j].RecordedAt })
	return filtered
}
