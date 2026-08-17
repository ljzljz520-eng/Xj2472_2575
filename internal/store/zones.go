package store

import (
	"fmt"

	"coldchain-alert/internal/domain"
	"go.etcd.io/bbolt"
)

func (s *Store) SaveZone(zone domain.Zone) error {
	if err := domain.ValidateZone(zone); err != nil {
		return err
	}
	return s.withUpdate(func(tx *bbolt.Tx) error {
		if tx.Bucket([]byte(BucketZones)).Get([]byte(zone.ID)) != nil {
			return fmt.Errorf("%w: zone %s", domain.ErrAlreadyExists, zone.ID)
		}
		return putJSON(tx, BucketZones, zone.ID, zone)
	})
}

func (s *Store) UpdateZone(zone domain.Zone) error {
	if err := domain.ValidateZone(zone); err != nil {
		return err
	}
	return s.withUpdate(func(tx *bbolt.Tx) error {
		if tx.Bucket([]byte(BucketZones)).Get([]byte(zone.ID)) == nil {
			return domain.ErrNotFound
		}
		return putJSON(tx, BucketZones, zone.ID, zone)
	})
}

func (s *Store) GetZone(id string) (*domain.Zone, error) {
	var zone domain.Zone
	err := s.withView(func(tx *bbolt.Tx) error { return getJSON(tx, BucketZones, id, &zone) })
	if err != nil {
		return nil, err
	}
	return &zone, nil
}

func (s *Store) ListZones() ([]domain.Zone, error) {
	var zones []domain.Zone
	err := s.withView(func(tx *bbolt.Tx) error {
		var err error
		zones, err = listJSON[domain.Zone](tx, BucketZones)
		return err
	})
	return zones, err
}

func (s *Store) EnabledZones() ([]domain.Zone, error) {
	zones, err := s.ListZones()
	if err != nil {
		return nil, err
	}
	enabled := make([]domain.Zone, 0, len(zones))
	for _, zone := range zones {
		if zone.Enabled {
			enabled = append(enabled, zone)
		}
	}
	return enabled, nil
}
