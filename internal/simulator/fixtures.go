package simulator

import (
	"fmt"

	"coldchain-alert/internal/domain"
	"coldchain-alert/internal/service"
)

type Fixture struct {
	Zone   domain.Zone
	Device domain.Device
}

func DefaultFixture() Fixture {
	return Fixture{
		Zone:   domain.Zone{ID: "zone-north", Name: "North Frozen Storage", TargetMin: -22, TargetMax: -18, HumidityMin: 35, HumidityMax: 65, Enabled: true},
		Device: domain.Device{ID: "sensor-north-01", ZoneID: "zone-north", Name: "North Sensor 01", Model: "TC-900", Online: true, LastSeen: "2026-01-24T12:00:00Z", BatteryPct: 92},
	}
}

func Seed(workflow *service.WorkflowService) (Fixture, error) {
	fixture := DefaultFixture()
	if err := workflow.RegisterColdRoom(fixture.Zone, fixture.Device, "bootstrap", "2026-01-24T00:00:00Z"); err != nil {
		return Fixture{}, err
	}
	for hour := 0; hour < 24; hour++ {
		at := fmt.Sprintf("2026-01-24T%02d:00:00Z", hour)
		temperature := -20.0
		humidity := 48.0 + float64(hour%4)
		if hour == 16 {
			temperature = -11.5
		}
		reading := domain.TemperatureReading{ID: fmt.Sprintf("seed-reading-%02d", hour), DeviceID: fixture.Device.ID, ZoneID: fixture.Zone.ID, Temperature: temperature, Humidity: humidity, DoorCount: hour / 6, RecordedAt: at}
		if _, err := workflow.CaptureAndEvaluate(reading, "simulator", at); err != nil {
			return Fixture{}, err
		}
	}
	for index, opened := range []bool{true, true, false, true} {
		event := domain.DoorEvent{ID: fmt.Sprintf("seed-door-%02d", index), ZoneID: fixture.Zone.ID, DeviceID: fixture.Device.ID, Opened: opened, Duration: 20 + index*25, RecordedAt: fmt.Sprintf("2026-01-24T%02d:30:00Z", index+8)}
		if err := workflow.RecordDoorAndAudit(event, "simulator", event.RecordedAt); err != nil {
			return Fixture{}, err
		}
	}
	return fixture, nil
}

func ReadingFor(zone, device, id, at string, temperature, humidity float64) domain.TemperatureReading {
	return domain.TemperatureReading{ID: id, ZoneID: zone, DeviceID: device, Temperature: temperature, Humidity: humidity, RecordedAt: at}
}

func NormalReading(zone, device, id, at string) domain.TemperatureReading {
	return ReadingFor(zone, device, id, at, -20, 50)
}

func CriticalReading(zone, device, id, at string) domain.TemperatureReading {
	return ReadingFor(zone, device, id, at, -8, 50)
}
