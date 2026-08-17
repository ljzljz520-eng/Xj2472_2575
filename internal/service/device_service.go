package service

import (
	"fmt"

	"coldchain-alert/internal/domain"
	"coldchain-alert/internal/store"
)

type DeviceService struct {
	storage *store.Store
}

func NewDeviceService(storage *store.Store) *DeviceService {
	return &DeviceService{storage: storage}
}

func (s *DeviceService) RegisterZone(zone domain.Zone) error {
	if !zone.Enabled {
		zone.Enabled = true
	}
	return s.storage.SaveZone(zone)
}

func (s *DeviceService) ChangeZone(zone domain.Zone) error {
	return s.storage.UpdateZone(zone)
}

func (s *DeviceService) RegisterDevice(device domain.Device) error {
	return s.storage.SaveDevice(device)
}

func (s *DeviceService) SetDeviceOnline(deviceID string, online bool, seenAt string) (*domain.Device, error) {
	device, err := s.storage.GetDevice(deviceID)
	if err != nil {
		return nil, err
	}
	device.Online = online
	if seenAt != "" {
		device.LastSeen = seenAt
	}
	if err := s.storage.UpdateDevice(*device); err != nil {
		return nil, err
	}
	return device, nil
}

func (s *DeviceService) DeviceHealth(zoneID string) (int, int, error) {
	devices, err := s.storage.ListDevicesByZone(zoneID)
	if err != nil {
		return 0, 0, fmt.Errorf("list zone devices: %w", err)
	}
	online := 0
	for _, device := range devices {
		if device.Online {
			online++
		}
	}
	return online, len(devices), nil
}

func (s *DeviceService) ListDevices(zoneID string) ([]domain.Device, error) {
	if zoneID == "" {
		return s.storage.ListDevices()
	}
	return s.storage.ListDevicesByZone(zoneID)
}
