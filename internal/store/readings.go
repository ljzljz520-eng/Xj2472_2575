package store

import (
	"fmt"

	"coldchain-alert/internal/domain"
	bolt "go.etcd.io/bbolt"
)

func (s *Store) SaveReading(reading domain.TemperatureReading) error {
	if err := domain.ValidateReading(reading); err != nil {
		return err
	}
	if _, err := s.GetDevice(reading.DeviceID); err != nil {
		return fmt.Errorf("reading device: %w", err)
	}
	return s.withUpdate(func(tx *bolt.Tx) error {
		if tx.Bucket([]byte(BucketReadings)).Get([]byte(reading.ID)) != nil {
			return fmt.Errorf("%w: reading %s", domain.ErrAlreadyExists, reading.ID)
		}
		return putJSON(tx, BucketReadings, reading.ID, reading)
	})
}

func (s *Store) GetReading(id string) (*domain.TemperatureReading, error) {
	var reading domain.TemperatureReading
	err := s.withView(func(tx *bolt.Tx) error { return getJSON(tx, BucketReadings, id, &reading) })
	if err != nil {
		return nil, err
	}
	return &reading, nil
}

func (s *Store) ListReadings() ([]domain.TemperatureReading, error) {
	var readings []domain.TemperatureReading
	err := s.withView(func(tx *bolt.Tx) error {
		var err error
		readings, err = listJSON[domain.TemperatureReading](tx, BucketReadings)
		return err
	})
	return readings, err
}

func (s *Store) ListReadingsByZone(zoneID string) ([]domain.TemperatureReading, error) {
	readings, err := s.ListReadings()
	if err != nil {
		return nil, err
	}
	filtered := make([]domain.TemperatureReading, 0, len(readings))
	for _, reading := range readings {
		if reading.ZoneID == zoneID {
			filtered = append(filtered, reading)
		}
	}
	return filtered, nil
}

func (s *Store) LatestReading(zoneID string) (*domain.TemperatureReading, error) {
	readings, err := s.ListReadingsByZone(zoneID)
	if err != nil {
		return nil, err
	}
	if len(readings) == 0 {
		return nil, domain.ErrNotFound
	}
	latest := readings[0]
	for _, reading := range readings[1:] {
		if reading.RecordedAt > latest.RecordedAt {
			latest = reading
		}
	}
	return &latest, nil
}
