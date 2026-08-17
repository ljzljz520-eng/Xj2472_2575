package store

import (
	"fmt"

	"coldchain-alert/internal/domain"
	bolt "go.etcd.io/bbolt"
)

func (s *Store) SaveDevice(device domain.Device) error {
	if err := domain.ValidateDevice(device); err != nil {
		return err
	}
	if _, err := s.GetZone(device.ZoneID); err != nil {
		return fmt.Errorf("device zone: %w", err)
	}
	return s.withUpdate(func(tx *bolt.Tx) error {
		if tx.Bucket([]byte(BucketDevices)).Get([]byte(device.ID)) != nil {
			return fmt.Errorf("%w: device %s", domain.ErrAlreadyExists, device.ID)
		}
		return putJSON(tx, BucketDevices, device.ID, device)
	})
}

func (s *Store) UpdateDevice(device domain.Device) error {
	if err := domain.ValidateDevice(device); err != nil {
		return err
	}
	return s.withUpdate(func(tx *bolt.Tx) error {
		if tx.Bucket([]byte(BucketDevices)).Get([]byte(device.ID)) == nil {
			return domain.ErrNotFound
		}
		return putJSON(tx, BucketDevices, device.ID, device)
	})
}

func (s *Store) GetDevice(id string) (*domain.Device, error) {
	var device domain.Device
	err := s.withView(func(tx *bolt.Tx) error { return getJSON(tx, BucketDevices, id, &device) })
	if err != nil {
		return nil, err
	}
	return &device, nil
}

func (s *Store) ListDevices() ([]domain.Device, error) {
	var devices []domain.Device
	err := s.withView(func(tx *bolt.Tx) error {
		var err error
		devices, err = listJSON[domain.Device](tx, BucketDevices)
		return err
	})
	return devices, err
}

func (s *Store) ListDevicesByZone(zoneID string) ([]domain.Device, error) {
	devices, err := s.ListDevices()
	if err != nil {
		return nil, err
	}
	filtered := make([]domain.Device, 0, len(devices))
	for _, device := range devices {
		if device.ZoneID == zoneID {
			filtered = append(filtered, device)
		}
	}
	return filtered, nil
}
