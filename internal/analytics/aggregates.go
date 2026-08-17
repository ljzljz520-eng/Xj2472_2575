package analytics

import (
	"sort"

	"coldchain-alert/internal/domain"
)

type ZoneAggregate struct {
	ZoneID       string  `json:"zone_id"`
	Readings     int     `json:"readings"`
	Alerts       int     `json:"alerts"`
	OpenAlerts   int     `json:"open_alerts"`
	AverageTemp  float64 `json:"average_temperature"`
	AverageHumid float64 `json:"average_humidity"`
	DoorEvents   int     `json:"door_events"`
}

func AggregateZones(zones []domain.Zone, readings []domain.TemperatureReading, alerts []domain.TemperatureAlert, doors []domain.DoorEvent) []ZoneAggregate {
	aggregates := make(map[string]*ZoneAggregate, len(zones))
	for _, zone := range zones {
		aggregates[zone.ID] = &ZoneAggregate{ZoneID: zone.ID}
	}
	for _, reading := range readings {
		aggregate, ok := aggregates[reading.ZoneID]
		if !ok {
			continue
		}
		aggregate.Readings++
		aggregate.AverageTemp += reading.Temperature
		aggregate.AverageHumid += reading.Humidity
	}
	for _, alert := range alerts {
		aggregate, ok := aggregates[alert.ZoneID]
		if !ok {
			continue
		}
		aggregate.Alerts++
		if domain.IsOpen(alert.Status) {
			aggregate.OpenAlerts++
		}
	}
	for _, door := range doors {
		if aggregate, ok := aggregates[door.ZoneID]; ok {
			aggregate.DoorEvents++
		}
	}
	result := make([]ZoneAggregate, 0, len(aggregates))
	for _, aggregate := range aggregates {
		if aggregate.Readings > 0 {
			aggregate.AverageTemp /= float64(aggregate.Readings)
			aggregate.AverageHumid /= float64(aggregate.Readings)
		}
		result = append(result, *aggregate)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ZoneID < result[j].ZoneID })
	return result
}

func CriticalZones(aggregates []ZoneAggregate) []string {
	ids := make([]string, 0)
	for _, aggregate := range aggregates {
		if aggregate.OpenAlerts > 0 {
			ids = append(ids, aggregate.ZoneID)
		}
	}
	sort.Strings(ids)
	return ids
}

func LatestReadings(readings []domain.TemperatureReading) map[string]domain.TemperatureReading {
	latest := make(map[string]domain.TemperatureReading)
	for _, reading := range readings {
		current, ok := latest[reading.ZoneID]
		if !ok || reading.RecordedAt > current.RecordedAt {
			latest[reading.ZoneID] = reading
		}
	}
	return latest
}
