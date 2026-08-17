package httpapi

import (
	"bytes"
	"net/http"
	"strings"

	"coldchain-alert/internal/domain"
	"coldchain-alert/internal/export"
)

type loginRequest struct {
	Username string      `json:"username"`
	Role     domain.Role `json:"role"`
}

func (a *API) handleLogin(writer http.ResponseWriter, request *http.Request) {
	if !methodAllowed(writer, request, http.MethodPost) {
		return
	}
	var input loginRequest
	if err := decodeJSON(request, &input); err != nil {
		writeError(writer, http.StatusBadRequest, err)
		return
	}
	session, err := a.auth.Login(input.Username, input.Role)
	if err != nil {
		writeError(writer, http.StatusUnauthorized, err)
		return
	}
	writeJSON(writer, http.StatusOK, session)
}

func (a *API) handleHealth(writer http.ResponseWriter, request *http.Request) {
	if !methodAllowed(writer, request, http.MethodGet) {
		return
	}
	writeJSON(writer, http.StatusOK, map[string]string{"status": "ok", "service": "coldchain-alert"})
}

func (a *API) handleZones(writer http.ResponseWriter, request *http.Request) {
	if !methodAllowed(writer, request, http.MethodGet) {
		return
	}
	zones, err := a.dashboard.AllZones()
	if err != nil {
		writeError(writer, http.StatusInternalServerError, err)
		return
	}
	writeJSON(writer, http.StatusOK, zones)
}

func (a *API) handleDashboard(writer http.ResponseWriter, request *http.Request) {
	if !methodAllowed(writer, request, http.MethodGet) {
		return
	}
	zoneID := request.URL.Query().Get("zone_id")
	if zoneID == "" {
		writeError(writer, http.StatusBadRequest, domain.ErrInvalidInput)
		return
	}
	summary, err := a.dashboard.Summary(zoneID, request.URL.Query().Get("since"))
	if err != nil {
		status := http.StatusInternalServerError
		if err == domain.ErrNotFound {
			status = http.StatusNotFound
		}
		writeError(writer, status, err)
		return
	}
	writeJSON(writer, http.StatusOK, summary)
}

func (a *API) handleAlerts(writer http.ResponseWriter, request *http.Request) {
	if !methodAllowed(writer, request, http.MethodGet) {
		return
	}
	filter := domain.AlertFilter{ZoneID: request.URL.Query().Get("zone_id"), Status: domain.AlertStatus(request.URL.Query().Get("status")), Severity: domain.AlertSeverity(request.URL.Query().Get("severity"))}
	alerts, err := a.alerts.ListAlerts(filter)
	if err != nil {
		writeError(writer, http.StatusInternalServerError, err)
		return
	}
	writeJSON(writer, http.StatusOK, alerts)
}

func (a *API) handleAlertPath(writer http.ResponseWriter, request *http.Request) {
	parts := strings.Split(strings.Trim(request.URL.Path, "/"), "/")
	if len(parts) < 3 || parts[2] == "" {
		writeError(writer, http.StatusNotFound, domain.ErrNotFound)
		return
	}
	alertID := parts[2]
	if len(parts) == 3 && request.Method == http.MethodGet {
		alert, err := a.alerts.ReadAlert(alertID)
		if err != nil {
			writeError(writer, http.StatusNotFound, err)
			return
		}
		writeJSON(writer, http.StatusOK, alert)
		return
	}
	if len(parts) == 4 && parts[3] == "severity" && request.Method == http.MethodGet {
		severity, err := a.alerts.GetAlertSeverity(alertID)
		if err != nil {
			writeError(writer, http.StatusNotFound, err)
			return
		}
		writeJSON(writer, http.StatusOK, map[string]string{"alert_id": alertID, "severity": string(severity)})
		return
	}
	if len(parts) == 4 && parts[3] == "ack" && request.Method == http.MethodPost {
		var input struct{ Actor, At string }
		if err := decodeJSON(request, &input); err != nil {
			writeError(writer, http.StatusBadRequest, err)
			return
		}
		alert, err := a.alerts.Acknowledge(alertID, input.Actor, input.At)
		if err != nil {
			writeError(writer, http.StatusConflict, err)
			return
		}
		writeJSON(writer, http.StatusOK, alert)
		return
	}
	if len(parts) == 4 && parts[3] == "review" && request.Method == http.MethodPost {
		var input struct{ Reviewer, Decision, Note, At string }
		if err := decodeJSON(request, &input); err != nil {
			writeError(writer, http.StatusBadRequest, err)
			return
		}
		review, err := a.alerts.Review(alertID, input.Reviewer, input.Decision, input.Note, input.At)
		if err != nil {
			writeError(writer, http.StatusConflict, err)
			return
		}
		writeJSON(writer, http.StatusOK, review)
		return
	}
	writeError(writer, http.StatusMethodNotAllowed, domain.ErrInvalidInput)
}

func (a *API) handleReading(writer http.ResponseWriter, request *http.Request) {
	if !methodAllowed(writer, request, http.MethodPost) {
		return
	}
	var reading domain.TemperatureReading
	if err := decodeJSON(request, &reading); err != nil {
		writeError(writer, http.StatusBadRequest, err)
		return
	}
	alert, err := a.workflow.CaptureAndEvaluate(reading, "api", reading.RecordedAt)
	if err != nil {
		writeError(writer, http.StatusBadRequest, err)
		return
	}
	writeJSON(writer, http.StatusAccepted, map[string]any{"reading": reading, "alert": alert})
}

func (a *API) handleDoorEvent(writer http.ResponseWriter, request *http.Request) {
	if !methodAllowed(writer, request, http.MethodPost) {
		return
	}
	var event domain.DoorEvent
	if err := decodeJSON(request, &event); err != nil {
		writeError(writer, http.StatusBadRequest, err)
		return
	}
	if err := a.workflow.RecordDoorAndAudit(event, "api", event.RecordedAt); err != nil {
		writeError(writer, http.StatusBadRequest, err)
		return
	}
	writeJSON(writer, http.StatusAccepted, event)
}

func (a *API) handleAlertExport(writer http.ResponseWriter, request *http.Request) {
	if !methodAllowed(writer, request, http.MethodGet) {
		return
	}
	alerts, err := a.alerts.ListAlerts(domain.AlertFilter{ZoneID: request.URL.Query().Get("zone_id")})
	if err != nil {
		writeError(writer, http.StatusInternalServerError, err)
		return
	}
	zoneMap := make(map[string]domain.Zone)
	rows := export.AlertRows(alerts, zoneMap)
	var buffer bytes.Buffer
	if err := export.WriteAlertCSV(&buffer, rows); err != nil {
		writeError(writer, http.StatusInternalServerError, err)
		return
	}
	writer.Header().Set("Content-Type", "text/csv")
	writer.WriteHeader(http.StatusOK)
	_, _ = writer.Write(buffer.Bytes())
}

func (a *API) handleComplianceReport(writer http.ResponseWriter, request *http.Request) {
	if !methodAllowed(writer, request, http.MethodGet) {
		return
	}
	zoneID := request.URL.Query().Get("zone_id")
	if zoneID == "" {
		writeError(writer, http.StatusBadRequest, domain.ErrInvalidInput)
		return
	}
	report, err := a.reports.Compliance(zoneID, request.URL.Query().Get("start"), request.URL.Query().Get("end"))
	if err != nil {
		writeError(writer, http.StatusNotFound, err)
		return
	}
	writeJSON(writer, http.StatusOK, report)
}
