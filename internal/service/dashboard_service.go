package service

import (
	"fmt"

	"coldchain-alert/internal/analytics"
	"coldchain-alert/internal/domain"
	"coldchain-alert/internal/store"
)

type DashboardService struct {
	storage *store.Store
}

func NewDashboardService(storage *store.Store) *DashboardService {
	return &DashboardService{storage: storage}
}

func (s *DashboardService) Summary(zoneID string, boundary string) (domain.DashboardSummary, error) {
	zone, err := s.storage.GetZone(zoneID)
	if err != nil {
		return domain.DashboardSummary{}, err
	}
	readings, err := s.storage.ListReadingsByZone(zoneID)
	if err != nil {
		return domain.DashboardSummary{}, fmt.Errorf("dashboard readings: %w", err)
	}
	if boundary != "" {
		readings = analytics.RecentReadings(readings, boundary, 24)
	}
	alerts, err := s.storage.ListAlertsByZone(zoneID)
	if err != nil {
		return domain.DashboardSummary{}, fmt.Errorf("dashboard alerts: %w", err)
	}
	devices, err := s.storage.ListDevicesByZone(zoneID)
	if err != nil {
		return domain.DashboardSummary{}, fmt.Errorf("dashboard devices: %w", err)
	}
	doors, err := s.storage.ListDoorEventsByZone(zoneID)
	if err != nil {
		return domain.DashboardSummary{}, fmt.Errorf("dashboard doors: %w", err)
	}
	openAlerts, criticalAlerts := domain.CountOpenAlerts(alerts)
	online, total := analytics.CountOnline(devices)
	var latest *domain.TemperatureReading
	if len(readings) > 0 {
		candidate := readings[len(readings)-1]
		latest = &candidate
	}
	return domain.DashboardSummary{
		Zone:              *zone,
		LatestReading:     latest,
		TemperatureSeries: analytics.BuildTemperatureSeries(readings, 24),
		HumiditySeries:    analytics.BuildHumiditySeries(readings, 24),
		OpenAlertCount:    openAlerts,
		CriticalCount:     criticalAlerts,
		DoorOpenCount:     analytics.CountDoorOpen(doors),
		OnlineDevices:     online,
		TotalDevices:      total,
		HealthScore:       domain.HealthScore(openAlerts, criticalAlerts, total-online, total),
	}, nil
}

func (s *DashboardService) AllZones() ([]analytics.ZoneAggregate, error) {
	zones, err := s.storage.ListZones()
	if err != nil {
		return nil, err
	}
	readings, err := s.storage.ListReadings()
	if err != nil {
		return nil, err
	}
	alerts, err := s.storage.ListAlerts()
	if err != nil {
		return nil, err
	}
	doors, err := s.storage.ListDoorEvents()
	if err != nil {
		return nil, err
	}
	return analytics.AggregateZones(zones, readings, alerts, doors), nil
}
