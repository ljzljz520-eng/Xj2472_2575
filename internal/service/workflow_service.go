package service

import (
	"fmt"

	"coldchain-alert/internal/domain"
)

type WorkflowService struct {
	devices *DeviceService
	alerts  *AlertService
	events  *EventService
	audits  *AuditService
}

func NewWorkflowService(devices *DeviceService, alerts *AlertService, events *EventService, audits *AuditService) *WorkflowService {
	return &WorkflowService{devices: devices, alerts: alerts, events: events, audits: audits}
}

func (s *WorkflowService) RegisterColdRoom(zone domain.Zone, device domain.Device, actor, at string) error {
	if err := s.devices.RegisterZone(zone); err != nil {
		return err
	}
	if err := s.devices.RegisterDevice(device); err != nil {
		return err
	}
	_, err := s.audits.Record(actor, "register", "zone", zone.ID, "cold room and sensor registered", at)
	return err
}

func (s *WorkflowService) CaptureAndEvaluate(reading domain.TemperatureReading, actor, at string) (*domain.TemperatureAlert, error) {
	alert, err := s.alerts.IngestReading(reading)
	if err != nil {
		return nil, err
	}
	if alert == nil {
		_, auditErr := s.audits.Record(actor, "capture", "reading", reading.ID, "reading within configured thresholds", at)
		if auditErr != nil {
			return nil, auditErr
		}
		return nil, nil
	}
	if _, err := s.audits.Record(actor, "open", "temperature_alert", alert.ID, alert.Reason, at); err != nil {
		return nil, err
	}
	return alert, nil
}

func (s *WorkflowService) AcknowledgeAndReview(alertID, warehouseUser, qualityUser, decision, note, at string) (*domain.AlertReview, error) {
	if _, err := s.alerts.Acknowledge(alertID, warehouseUser, at); err != nil {
		return nil, err
	}
	if _, err := s.audits.Record(warehouseUser, "acknowledge", "temperature_alert", alertID, "warehouse accepted alert", at); err != nil {
		return nil, err
	}
	review, err := s.alerts.Review(alertID, qualityUser, decision, note, at)
	if err != nil {
		return nil, err
	}
	if _, err := s.audits.Record(qualityUser, "review", "temperature_alert", alertID, decision, at); err != nil {
		return nil, err
	}
	return review, nil
}

func (s *WorkflowService) RecordDoorAndAudit(event domain.DoorEvent, actor, at string) error {
	if err := s.events.RecordDoorEvent(event); err != nil {
		return err
	}
	_, err := s.audits.Record(actor, "door_event", "door_event", event.ID, fmt.Sprintf("opened=%t duration=%d", event.Opened, event.Duration), at)
	return err
}
