package httpapi_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"coldchain-alert/internal/auth"
	"coldchain-alert/internal/domain"
	"coldchain-alert/internal/httpapi"
	"coldchain-alert/internal/service"
	"coldchain-alert/internal/store"
)

func TestHTTPLoginDashboardAndMissingRoute(t *testing.T) {
	storage, err := store.Open(filepath.Join(t.TempDir(), "http.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()
	devices := service.NewDeviceService(storage)
	alerts := service.NewAlertService(storage)
	events := service.NewEventService(storage)
	audits := service.NewAuditService(storage)
	reports := service.NewReportService(storage)
	workflow := service.NewWorkflowService(devices, alerts, events, audits)
	zone := domain.Zone{ID: "zone-http", Name: "HTTP Room", TargetMin: -22, TargetMax: -18, HumidityMin: 35, HumidityMax: 65, Enabled: true}
	device := domain.Device{ID: "device-http", ZoneID: zone.ID, Name: "HTTP Sensor", Model: "H1", Online: true, BatteryPct: 95}
	if err := workflow.RegisterColdRoom(zone, device, "bootstrap", "2026-01-01T00:00:00Z"); err != nil {
		t.Fatal(err)
	}
	if _, err := workflow.CaptureAndEvaluate(domain.TemperatureReading{ID: "reading-http", DeviceID: device.ID, ZoneID: zone.ID, Temperature: -20, Humidity: 50, RecordedAt: "2026-01-01T01:00:00Z"}, "bootstrap", "2026-01-01T01:00:00Z"); err != nil {
		t.Fatal(err)
	}
	api := httpapi.New(auth.NewManager(), devices, alerts, service.NewDashboardService(storage), events, workflow, audits, reports)
	handler := api.Handler()
	loginBody, _ := json.Marshal(map[string]string{"username": "visitor", "role": "visitor"})
	login := httptest.NewRecorder()
	handler.ServeHTTP(login, httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewReader(loginBody)))
	if login.Code != http.StatusOK {
		t.Fatalf("login failed: %d %s", login.Code, login.Body.String())
	}
	var session domain.Session
	if err := json.Unmarshal(login.Body.Bytes(), &session); err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodGet, "/api/dashboard?zone_id="+zone.ID, nil)
	request.Header.Set("Authorization", "Bearer "+session.Token)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("dashboard failed: %d %s", response.Code, response.Body.String())
	}
	missing := httptest.NewRequest(http.MethodGet, "/api/alerts/not-real", nil)
	missing.Header.Set("Authorization", "Bearer "+session.Token)
	missingResponse := httptest.NewRecorder()
	handler.ServeHTTP(missingResponse, missing)
	if missingResponse.Code != http.StatusNotFound {
		t.Fatalf("missing alert route should be 404: %d", missingResponse.Code)
	}
}
