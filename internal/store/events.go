package store

import (
	"fmt"

	"coldchain-alert/internal/domain"
	bolt "go.etcd.io/bbolt"
)

func (s *Store) SaveDoorEvent(event domain.DoorEvent) error {
	if err := domain.ValidateDoorEvent(event); err != nil {
		return err
	}
	return s.withUpdate(func(tx *bolt.Tx) error {
		if tx.Bucket([]byte(BucketDoors)).Get([]byte(event.ID)) != nil {
			return fmt.Errorf("%w: door event %s", domain.ErrAlreadyExists, event.ID)
		}
		return putJSON(tx, BucketDoors, event.ID, event)
	})
}

func (s *Store) ListDoorEvents() ([]domain.DoorEvent, error) {
	var events []domain.DoorEvent
	err := s.withView(func(tx *bolt.Tx) error {
		var err error
		events, err = listJSON[domain.DoorEvent](tx, BucketDoors)
		return err
	})
	return events, err
}

func (s *Store) ListDoorEventsByZone(zoneID string) ([]domain.DoorEvent, error) {
	events, err := s.ListDoorEvents()
	if err != nil {
		return nil, err
	}
	filtered := make([]domain.DoorEvent, 0, len(events))
	for _, event := range events {
		if event.ZoneID == zoneID {
			filtered = append(filtered, event)
		}
	}
	return filtered, nil
}

func (s *Store) DoorOpenCount(zoneID string) (int, error) {
	events, err := s.ListDoorEventsByZone(zoneID)
	if err != nil {
		return 0, err
	}
	count := 0
	for _, event := range events {
		if event.Opened {
			count++
		}
	}
	return count, nil
}
