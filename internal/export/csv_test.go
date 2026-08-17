package export_test

import (
	"bytes"
	"strings"
	"testing"

	"coldchain-alert/internal/domain"
	"coldchain-alert/internal/export"
)

func TestWriteAlertCSV(t *testing.T) {
	alerts := []domain.TemperatureAlert{{ID: "a1", ZoneID: "z1", Severity: domain.SeverityCritical, Status: domain.AlertOpen, Reason: "door, open", OpenedAt: "t1", UpdatedAt: "t1"}}
	rows := export.AlertRows(alerts, map[string]domain.Zone{"z1": {ID: "z1", Name: "North"}})
	var buffer bytes.Buffer
	if err := export.WriteAlertCSV(&buffer, rows); err != nil {
		t.Fatalf("csv write failed: %v", err)
	}
	text := buffer.String()
	if !strings.Contains(text, "alert_id") || !strings.Contains(text, "\"door, open\"") {
		t.Fatalf("csv escaping missing: %s", text)
	}
}

func TestBuildPackageJSON(t *testing.T) {
	summary := domain.DashboardSummary{Zone: domain.Zone{ID: "z1"}}
	payload := export.BuildPackage(summary, []domain.ExportRow{{AlertID: "a1"}})
	var buffer bytes.Buffer
	if err := export.WriteJSON(&buffer, payload); err != nil || !strings.Contains(buffer.String(), "generated_for") {
		t.Fatalf("json export failed: %v %s", err, buffer.String())
	}
}
