package store

import (
	"fmt"

	"coldchain-alert/internal/domain"
	bolt "go.etcd.io/bbolt"
)

func (s *Store) SaveAlert(alert domain.TemperatureAlert) error {
	if err := domain.ValidateAlert(alert); err != nil {
		return err
	}
	return s.withUpdate(func(tx *bolt.Tx) error {
		if tx.Bucket([]byte(BucketAlerts)).Get([]byte(alert.ID)) != nil {
			return fmt.Errorf("%w: alert %s", domain.ErrAlreadyExists, alert.ID)
		}
		return putJSON(tx, BucketAlerts, alert.ID, alert)
	})
}

func (s *Store) UpdateAlert(alert domain.TemperatureAlert) error {
	if err := domain.ValidateAlert(alert); err != nil {
		return err
	}
	return s.withUpdate(func(tx *bolt.Tx) error {
		if tx.Bucket([]byte(BucketAlerts)).Get([]byte(alert.ID)) == nil {
			return domain.ErrNotFound
		}
		return putJSON(tx, BucketAlerts, alert.ID, alert)
	})
}

func (s *Store) GetAlert(id string) (*domain.TemperatureAlert, error) {
	var alert domain.TemperatureAlert
	err := s.withView(func(tx *bolt.Tx) error { return getJSON(tx, BucketAlerts, id, &alert) })
	if err == domain.ErrNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &alert, nil
}

func (s *Store) ListAlerts() ([]domain.TemperatureAlert, error) {
	var alerts []domain.TemperatureAlert
	err := s.withView(func(tx *bolt.Tx) error {
		var err error
		alerts, err = listJSON[domain.TemperatureAlert](tx, BucketAlerts)
		return err
	})
	return alerts, err
}

func (s *Store) ListAlertsByZone(zoneID string) ([]domain.TemperatureAlert, error) {
	alerts, err := s.ListAlerts()
	if err != nil {
		return nil, err
	}
	filtered := make([]domain.TemperatureAlert, 0, len(alerts))
	for _, alert := range alerts {
		if alert.ZoneID == zoneID {
			filtered = append(filtered, alert)
		}
	}
	return filtered, nil
}

func (s *Store) DeleteAlert(id string) error {
	return s.Delete(BucketAlerts, id)
}
