package service_test

import (
	"testing"

	"coldchain-alert/internal/domain"
)

func TestWorkflowAlertReviewAndResolution(t *testing.T) {
	storage, workflow, alerts, _, _ := newWorkflow(t)
	defer storage.Close()
	zone, device := registerFixture(t, workflow)
	reading := domain.TemperatureReading{ID: "reading-review", DeviceID: device.ID, ZoneID: zone.ID, Temperature: -8, Humidity: 52, RecordedAt: "2026-01-01T03:00:00Z"}
	alert, err := workflow.CaptureAndEvaluate(reading, "warehouse", reading.RecordedAt)
	if err != nil || alert == nil {
		t.Fatalf("alert creation failed: %#v %v", alert, err)
	}
	review, err := workflow.AcknowledgeAndReview(alert.ID, "warehouse", "quality", "approve", "verified chamber door", "2026-01-01T04:00:00Z")
	if err != nil || review == nil {
		t.Fatalf("review failed: %#v %v", review, err)
	}
	if review.Decision != "approve" {
		t.Fatalf("wrong review decision: %#v", review)
	}
	updated, err := alerts.ReadAlert(alert.ID)
	if err != nil || updated.Status != domain.AlertResolved {
		t.Fatalf("alert was not resolved: %#v %v", updated, err)
	}
	reviews, err := alerts.Reviews(alert.ID)
	if err != nil || len(reviews) != 1 {
		t.Fatalf("review history mismatch: %#v %v", reviews, err)
	}
}
