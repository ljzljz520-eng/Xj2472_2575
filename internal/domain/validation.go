package domain

import (
	"fmt"
	"strings"
)

func ValidateZone(zone Zone) error {
	if strings.TrimSpace(zone.ID) == "" || strings.TrimSpace(zone.Name) == "" {
		return fmt.Errorf("%w: zone id and name are required", ErrInvalidInput)
	}
	if zone.TargetMin >= zone.TargetMax {
		return fmt.Errorf("%w: target range is reversed", ErrInvalidInput)
	}
	if zone.HumidityMin >= zone.HumidityMax {
		return fmt.Errorf("%w: humidity range is reversed", ErrInvalidInput)
	}
	if zone.TargetMin < -80 || zone.TargetMax > 60 {
		return fmt.Errorf("%w: target temperature is outside sensor range", ErrInvalidInput)
	}
	return nil
}

func ValidateDevice(device Device) error {
	if strings.TrimSpace(device.ID) == "" || strings.TrimSpace(device.ZoneID) == "" {
		return fmt.Errorf("%w: device identity is required", ErrInvalidInput)
	}
	if strings.TrimSpace(device.Name) == "" || strings.TrimSpace(device.Model) == "" {
		return fmt.Errorf("%w: device name and model are required", ErrInvalidInput)
	}
	if device.BatteryPct < 0 || device.BatteryPct > 100 {
		return fmt.Errorf("%w: battery percentage must be between zero and one hundred", ErrInvalidInput)
	}
	return nil
}

func ValidateReading(reading TemperatureReading) error {
	if reading.ID == "" || reading.DeviceID == "" || reading.ZoneID == "" || reading.RecordedAt == "" {
		return fmt.Errorf("%w: reading identity and timestamp are required", ErrInvalidInput)
	}
	if reading.Temperature < -100 || reading.Temperature > 100 {
		return fmt.Errorf("%w: temperature is outside sensor range", ErrInvalidInput)
	}
	if reading.Humidity < 0 || reading.Humidity > 100 {
		return fmt.Errorf("%w: humidity must be between zero and one hundred", ErrInvalidInput)
	}
	if reading.DoorCount < 0 {
		return fmt.Errorf("%w: door count cannot be negative", ErrInvalidInput)
	}
	return nil
}

func ValidateAlert(alert TemperatureAlert) error {
	if alert.ID == "" || alert.ZoneID == "" || alert.DeviceID == "" || alert.ReadingID == "" {
		return fmt.Errorf("%w: alert references are required", ErrInvalidInput)
	}
	if alert.Severity.Weight() == 0 {
		return fmt.Errorf("%w: unknown alert severity", ErrInvalidInput)
	}
	if alert.Status == "" || alert.Reason == "" || alert.OpenedAt == "" {
		return fmt.Errorf("%w: alert status, reason, and timestamp are required", ErrInvalidInput)
	}
	return nil
}

func ValidateReview(review AlertReview) error {
	if review.ID == "" || review.AlertID == "" || strings.TrimSpace(review.Reviewer) == "" {
		return fmt.Errorf("%w: review identity is required", ErrInvalidInput)
	}
	if review.Decision != "approve" && review.Decision != "reject" {
		return fmt.Errorf("%w: decision must be approve or reject", ErrInvalidInput)
	}
	if review.ReviewedAt == "" {
		return fmt.Errorf("%w: review timestamp is required", ErrInvalidInput)
	}
	return nil
}

func ValidateDoorEvent(event DoorEvent) error {
	if event.ID == "" || event.ZoneID == "" || event.DeviceID == "" || event.RecordedAt == "" {
		return fmt.Errorf("%w: door event references are required", ErrInvalidInput)
	}
	if event.Duration < 0 {
		return fmt.Errorf("%w: door duration cannot be negative", ErrInvalidInput)
	}
	return nil
}

func ValidateRole(role Role) error {
	if !role.Valid() {
		return fmt.Errorf("%w: unsupported role %q", ErrInvalidInput, role)
	}
	return nil
}

func RequireNonEmpty(values ...string) error {
	for _, value := range values {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%w: empty value", ErrInvalidInput)
		}
	}
	return nil
}
