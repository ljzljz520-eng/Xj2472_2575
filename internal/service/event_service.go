package service

import (
	"fmt"

	"coldchain-alert/internal/domain"
	"coldchain-alert/internal/store"
)

type EventService struct {
	storage *store.Store
}

func NewEventService(storage *store.Store) *EventService {
	return &EventService{storage: storage}
}

func (s *EventService) RecordDoorEvent(event domain.DoorEvent) error {
	if _, err := s.storage.GetZone(event.ZoneID); err != nil {
		return fmt.Errorf("door zone: %w", err)
	}
	if device, err := s.storage.GetDevice(event.DeviceID); err != nil {
		return fmt.Errorf("door device: %w", err)
	} else if device.ZoneID != event.ZoneID {
		return fmt.Errorf("%w: device belongs to another zone", domain.ErrInvalidInput)
	}
	return s.storage.SaveDoorEvent(event)
}

func (s *EventService) ListDoorEvents(zoneID string) ([]domain.DoorEvent, error) {
	if zoneID == "" {
		return s.storage.ListDoorEvents()
	}
	return s.storage.ListDoorEventsByZone(zoneID)
}

func (s *EventService) DoorSummary(zoneID string) (int, int, error) {
	events, err := s.ListDoorEvents(zoneID)
	if err != nil {
		return 0, 0, err
	}
	open := 0
	longOpen := 0
	for _, event := range events {
		if event.Opened {
			open++
		}
		if event.Duration >= 60 {
			longOpen++
		}
	}
	return open, longOpen, nil
}
