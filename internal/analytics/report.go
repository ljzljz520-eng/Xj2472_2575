package analytics

import (
	"fmt"
	"sort"

	"coldchain-alert/internal/domain"
)

type ComplianceReport struct {
	ZoneID             string   `json:"zone_id"`
	ZoneName           string   `json:"zone_name"`
	WindowStart        string   `json:"window_start"`
	WindowEnd          string   `json:"window_end"`
	ReadingCount       int      `json:"reading_count"`
	StableReadings     int      `json:"stable_readings"`
	ExcursionReadings  int      `json:"excursion_readings"`
	OpenAlerts         int      `json:"open_alerts"`
	ResolvedAlerts     int      `json:"resolved_alerts"`
	CriticalAlerts     int      `json:"critical_alerts"`
	DoorOpenCount      int      `json:"door_open_count"`
	LongDoorEvents     int      `json:"long_door_events"`
	OnlineDevices      int      `json:"online_devices"`
	TotalDevices       int      `json:"total_devices"`
	AverageTemperature float64  `json:"average_temperature"`
	MinimumTemperature float64  `json:"minimum_temperature"`
	MaximumTemperature float64  `json:"maximum_temperature"`
	AverageHumidity    float64  `json:"average_humidity"`
	CompliancePercent  float64  `json:"compliance_percent"`
	Status             string   `json:"status"`
	Recommendations    []string `json:"recommendations"`
}

type ReadingRange struct {
	Minimum float64 `json:"minimum"`
	Maximum float64 `json:"maximum"`
	Average float64 `json:"average"`
}

func BuildComplianceReport(zone domain.Zone, readings []domain.TemperatureReading, alerts []domain.TemperatureAlert, doors []domain.DoorEvent, devices []domain.Device, start, end string) ComplianceReport {
	report := ComplianceReport{ZoneID: zone.ID, ZoneName: zone.Name, WindowStart: start, WindowEnd: end, TotalDevices: len(devices)}
	report.MinimumTemperature, report.MaximumTemperature = 0, 0
	totalTemperature := 0.0
	totalHumidity := 0.0
	for index, reading := range readings {
		if start != "" && reading.RecordedAt < start {
			continue
		}
		if end != "" && reading.RecordedAt > end {
			continue
		}
		report.ReadingCount++
		if index == 0 || reading.Temperature < report.MinimumTemperature {
			report.MinimumTemperature = reading.Temperature
		}
		if index == 0 || reading.Temperature > report.MaximumTemperature {
			report.MaximumTemperature = reading.Temperature
		}
		totalTemperature += reading.Temperature
		totalHumidity += reading.Humidity
		if reading.Temperature >= zone.TargetMin && reading.Temperature <= zone.TargetMax && reading.Humidity >= zone.HumidityMin && reading.Humidity <= zone.HumidityMax {
			report.StableReadings++
		} else {
			report.ExcursionReadings++
		}
	}
	if report.ReadingCount > 0 {
		report.AverageTemperature = totalTemperature / float64(report.ReadingCount)
		report.AverageHumidity = totalHumidity / float64(report.ReadingCount)
	}
	for _, alert := range alerts {
		if start != "" && alert.OpenedAt < start {
			continue
		}
		if end != "" && alert.OpenedAt > end {
			continue
		}
		if domain.IsOpen(alert.Status) {
			report.OpenAlerts++
		} else if domain.StatusTerminal(alert.Status) {
			report.ResolvedAlerts++
		}
		if alert.Severity == domain.SeverityCritical {
			report.CriticalAlerts++
		}
	}
	for _, event := range doors {
		if start != "" && event.RecordedAt < start {
			continue
		}
		if end != "" && event.RecordedAt > end {
			continue
		}
		if event.Opened {
			report.DoorOpenCount++
		}
		if event.Duration >= 60 {
			report.LongDoorEvents++
		}
	}
	for _, device := range devices {
		if device.Online {
			report.OnlineDevices++
		}
	}
	if report.ReadingCount > 0 {
		report.CompliancePercent = float64(report.StableReadings) * 100 / float64(report.ReadingCount)
	}
	report.Status = reportStatus(report)
	report.Recommendations = Recommendations(report)
	return report
}

func reportStatus(report ComplianceReport) string {
	if report.ReadingCount == 0 {
		return "no_data"
	}
	if report.CriticalAlerts > 0 || report.CompliancePercent < 80 {
		return "attention"
	}
	if report.OpenAlerts > 0 || report.CompliancePercent < 95 {
		return "watch"
	}
	return "compliant"
}

func Recommendations(report ComplianceReport) []string {
	recommendations := make([]string, 0, 4)
	if report.CriticalAlerts > 0 {
		recommendations = append(recommendations, "quality inspector should review critical excursions")
	}
	if report.OpenAlerts > 0 {
		recommendations = append(recommendations, "assign open alerts to a warehouse operator")
	}
	if report.LongDoorEvents > 0 {
		recommendations = append(recommendations, "inspect door seals and loading procedures")
	}
	if report.OnlineDevices < report.TotalDevices {
		recommendations = append(recommendations, "restore offline sensors before the next shift")
	}
	if len(recommendations) == 0 {
		recommendations = append(recommendations, "continue the configured monitoring cadence")
	}
	return recommendations
}

func TemperatureRange(readings []domain.TemperatureReading) ReadingRange {
	if len(readings) == 0 {
		return ReadingRange{}
	}
	rangeValue := ReadingRange{Minimum: readings[0].Temperature, Maximum: readings[0].Temperature}
	for _, reading := range readings {
		if reading.Temperature < rangeValue.Minimum {
			rangeValue.Minimum = reading.Temperature
		}
		if reading.Temperature > rangeValue.Maximum {
			rangeValue.Maximum = reading.Temperature
		}
		rangeValue.Average += reading.Temperature
	}
	rangeValue.Average /= float64(len(readings))
	return rangeValue
}

func HumidityRange(readings []domain.TemperatureReading) ReadingRange {
	if len(readings) == 0 {
		return ReadingRange{}
	}
	rangeValue := ReadingRange{Minimum: readings[0].Humidity, Maximum: readings[0].Humidity}
	for _, reading := range readings {
		if reading.Humidity < rangeValue.Minimum {
			rangeValue.Minimum = reading.Humidity
		}
		if reading.Humidity > rangeValue.Maximum {
			rangeValue.Maximum = reading.Humidity
		}
		rangeValue.Average += reading.Humidity
	}
	rangeValue.Average /= float64(len(readings))
	return rangeValue
}

func Explain(report ComplianceReport) string {
	return fmt.Sprintf("%s: %.1f%% compliant, %d open alerts, %d/%d sensors online", report.ZoneName, report.CompliancePercent, report.OpenAlerts, report.OnlineDevices, report.TotalDevices)
}

func SortReports(reports []ComplianceReport) {
	sort.SliceStable(reports, func(i, j int) bool {
		if reports[i].Status != reports[j].Status {
			return reports[i].Status < reports[j].Status
		}
		return reports[i].ZoneID < reports[j].ZoneID
	})
}
