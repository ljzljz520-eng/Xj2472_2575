package service

import (
	"fmt"

	"coldchain-alert/internal/analytics"
	"coldchain-alert/internal/domain"
	"coldchain-alert/internal/store"
)

type ReportService struct {
	storage *store.Store
}

func NewReportService(storage *store.Store) *ReportService {
	return &ReportService{storage: storage}
}

func (s *ReportService) Compliance(zoneID, start, end string) (analytics.ComplianceReport, error) {
	zone, err := s.storage.GetZone(zoneID)
	if err != nil {
		return analytics.ComplianceReport{}, err
	}
	readings, err := s.storage.ListReadingsByZone(zoneID)
	if err != nil {
		return analytics.ComplianceReport{}, fmt.Errorf("report readings: %w", err)
	}
	alerts, err := s.storage.ListAlertsByZone(zoneID)
	if err != nil {
		return analytics.ComplianceReport{}, fmt.Errorf("report alerts: %w", err)
	}
	doors, err := s.storage.ListDoorEventsByZone(zoneID)
	if err != nil {
		return analytics.ComplianceReport{}, fmt.Errorf("report doors: %w", err)
	}
	devices, err := s.storage.ListDevicesByZone(zoneID)
	if err != nil {
		return analytics.ComplianceReport{}, fmt.Errorf("report devices: %w", err)
	}
	return analytics.BuildComplianceReport(*zone, readings, alerts, doors, devices, start, end), nil
}

func (s *ReportService) Risk(zoneID string) ([]analytics.RiskSignal, error) {
	zone, err := s.storage.GetZone(zoneID)
	if err != nil {
		return nil, err
	}
	readings, err := s.storage.ListReadingsByZone(zoneID)
	if err != nil {
		return nil, err
	}
	alerts, err := s.storage.ListAlertsByZone(zoneID)
	if err != nil {
		return nil, err
	}
	devices, err := s.storage.ListDevicesByZone(zoneID)
	if err != nil {
		return nil, err
	}
	doors, err := s.storage.ListDoorEventsByZone(zoneID)
	if err != nil {
		return nil, err
	}
	return analytics.BuildRiskSignals(*zone, readings, alerts, devices, doors), nil
}

func (s *ReportService) ZoneHealth(zoneID string) (string, error) {
	report, err := s.Compliance(zoneID, "", "")
	if err != nil {
		return "", err
	}
	if report.Status == "no_data" {
		return "unknown", nil
	}
	if report.Status == "attention" {
		return "investigate", nil
	}
	if report.Status == "watch" {
		return "monitor", nil
	}
	return "ready", nil
}

func BuildExportRows(alerts []domain.TemperatureAlert, zones []domain.Zone) []domain.ExportRow {
	zoneMap := make(map[string]domain.Zone, len(zones))
	for _, zone := range zones {
		zoneMap[zone.ID] = zone
	}
	rows := make([]domain.ExportRow, 0, len(alerts))
	for _, alert := range alerts {
		name := alert.ZoneID
		if zone, ok := zoneMap[alert.ZoneID]; ok {
			name = zone.Name
		}
		rows = append(rows, domain.ExportRow{AlertID: alert.ID, Zone: name, Severity: string(alert.Severity), Status: string(alert.Status), Reason: alert.Reason, OpenedAt: alert.OpenedAt, UpdatedAt: alert.UpdatedAt})
	}
	return rows
}
