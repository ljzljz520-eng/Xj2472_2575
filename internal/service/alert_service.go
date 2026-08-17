package service

import (
	"fmt"

	"coldchain-alert/internal/domain"
	"coldchain-alert/internal/store"
)

type AlertService struct {
	storage *store.Store
}

func NewAlertService(storage *store.Store) *AlertService {
	return &AlertService{storage: storage}
}

func (s *AlertService) IngestReading(reading domain.TemperatureReading) (*domain.TemperatureAlert, error) {
	if err := s.storage.SaveReading(reading); err != nil {
		return nil, err
	}
	zone, err := s.storage.GetZone(reading.ZoneID)
	if err != nil {
		return nil, err
	}
	device, err := s.storage.GetDevice(reading.DeviceID)
	if err != nil {
		return nil, err
	}
	severity, reason := domain.EvaluateSeverity(*zone, reading)
	if severity == "" {
		return nil, nil
	}
	alert := domain.TemperatureAlert{
		ID:          "alert-" + reading.ID,
		ZoneID:      zone.ID,
		DeviceID:    device.ID,
		ReadingID:   reading.ID,
		Severity:    severity,
		Status:      domain.AlertOpen,
		Reason:      reason,
		Temperature: reading.Temperature,
		Humidity:    reading.Humidity,
		OpenedAt:    reading.RecordedAt,
		UpdatedAt:   reading.RecordedAt,
	}
	if err := s.storage.SaveAlert(alert); err != nil {
		return nil, err
	}
	return &alert, nil
}

func (s *AlertService) ReadAlert(alertID string) (*domain.TemperatureAlert, error) {
	alert, err := s.storage.GetAlert(alertID)
	if err != nil {
		return nil, err
	}
	if alert == nil {
		return nil, domain.ErrNotFound
	}
	return alert, nil
}

func (s *AlertService) GetAlertSeverity(alertID string) (domain.AlertSeverity, error) {
	alert, err := s.ReadAlert(alertID)
	if err != nil {
		return "", err
	}
	return alert.Severity, nil
}

func (s *AlertService) ListAlerts(filter domain.AlertFilter) ([]domain.TemperatureAlert, error) {
	alerts, err := s.storage.ListAlerts()
	if err != nil {
		return nil, err
	}
	return domain.FilterAlerts(alerts, filter), nil
}

func (s *AlertService) Acknowledge(alertID, actor, at string) (*domain.TemperatureAlert, error) {
	alert, err := s.ReadAlert(alertID)
	if err != nil {
		return nil, err
	}
	if alert.Status != domain.AlertOpen {
		return nil, fmt.Errorf("%w: acknowledge from %s", domain.ErrInvalidState, alert.Status)
	}
	alert.Status = domain.AlertAcknowledged
	alert.AssignedTo = actor
	alert.UpdatedAt = at
	if err := s.storage.UpdateAlert(*alert); err != nil {
		return nil, err
	}
	return alert, nil
}

func (s *AlertService) Review(alertID, reviewer, decision, note, at string) (*domain.AlertReview, error) {
	alert, err := s.ReadAlert(alertID)
	if err != nil {
		return nil, err
	}
	if alert.Status != domain.AlertAcknowledged && alert.Status != domain.AlertUnderReview {
		return nil, fmt.Errorf("%w: review from %s", domain.ErrInvalidState, alert.Status)
	}
	alert.Status = domain.AlertUnderReview
	alert.UpdatedAt = at
	if err := s.storage.UpdateAlert(*alert); err != nil {
		return nil, err
	}
	review := domain.AlertReview{ID: "review-" + alertID + "-" + reviewer, AlertID: alertID, Reviewer: reviewer, Decision: decision, Note: note, ReviewedAt: at}
	if err := s.storage.SaveReview(review); err != nil {
		return nil, err
	}
	if decision == "approve" {
		alert.Status = domain.AlertResolved
	} else {
		alert.Status = domain.AlertAcknowledged
	}
	if err := s.storage.UpdateAlert(*alert); err != nil {
		return nil, err
	}
	return &review, nil
}

func (s *AlertService) Reviews(alertID string) ([]domain.AlertReview, error) {
	return s.storage.ListReviewsByAlert(alertID)
}
