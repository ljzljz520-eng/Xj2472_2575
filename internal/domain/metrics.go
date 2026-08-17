package domain

import "fmt"

func EvaluateSeverity(zone Zone, reading TemperatureReading) (AlertSeverity, string) {
	temperatureOutside := reading.Temperature < zone.TargetMin || reading.Temperature > zone.TargetMax
	humidityOutside := reading.Humidity < zone.HumidityMin || reading.Humidity > zone.HumidityMax
	if !temperatureOutside && !humidityOutside {
		return "", ""
	}
	temperatureDistance := distanceOutside(reading.Temperature, zone.TargetMin, zone.TargetMax)
	humidityDistance := distanceOutside(reading.Humidity, zone.HumidityMin, zone.HumidityMax)
	if temperatureDistance >= 8 || humidityDistance >= 25 {
		return SeverityCritical, fmt.Sprintf("critical excursion: %.1fC and %.1f%% humidity", reading.Temperature, reading.Humidity)
	}
	if temperatureDistance >= 3 || humidityDistance >= 10 {
		return SeverityWarning, fmt.Sprintf("warning excursion: %.1fC and %.1f%% humidity", reading.Temperature, reading.Humidity)
	}
	return SeverityInfo, fmt.Sprintf("minor excursion: %.1fC and %.1f%% humidity", reading.Temperature, reading.Humidity)
}

func distanceOutside(value, minimum, maximum float64) float64 {
	if value < minimum {
		return minimum - value
	}
	if value > maximum {
		return value - maximum
	}
	return 0
}

func HealthScore(openAlerts, criticalAlerts, offlineDevices, totalDevices int) int {
	score := 100 - openAlerts*8 - criticalAlerts*15 - offlineDevices*10
	if totalDevices == 0 {
		return 0
	}
	if score < 0 {
		return 0
	}
	if score > 100 {
		return 100
	}
	return score
}

func IsReadingStable(points []SeriesPoint, minimum, maximum float64) bool {
	if len(points) == 0 {
		return false
	}
	for _, point := range points {
		if point.Value < minimum || point.Value > maximum {
			return false
		}
	}
	return true
}

func LatestPoint(points []SeriesPoint) (SeriesPoint, bool) {
	if len(points) == 0 {
		return SeriesPoint{}, false
	}
	latest := points[0]
	for _, point := range points[1:] {
		if point.At > latest.At {
			latest = point
		}
	}
	return latest, true
}

func Average(points []SeriesPoint) float64 {
	if len(points) == 0 {
		return 0
	}
	total := 0.0
	for _, point := range points {
		total += point.Value
	}
	return total / float64(len(points))
}
