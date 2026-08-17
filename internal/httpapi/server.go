package httpapi

import (
	"net/http"
	"strings"

	"coldchain-alert/internal/auth"
	"coldchain-alert/internal/domain"
	"coldchain-alert/internal/service"
)

type API struct {
	auth      *auth.Manager
	devices   *service.DeviceService
	alerts    *service.AlertService
	dashboard *service.DashboardService
	events    *service.EventService
	workflow  *service.WorkflowService
	audits    *service.AuditService
	reports   *service.ReportService
}

func New(authManager *auth.Manager, devices *service.DeviceService, alerts *service.AlertService, dashboard *service.DashboardService, events *service.EventService, workflow *service.WorkflowService, audits *service.AuditService, reports *service.ReportService) *API {
	return &API{auth: authManager, devices: devices, alerts: alerts, dashboard: dashboard, events: events, workflow: workflow, audits: audits, reports: reports}
}

func (a *API) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/login", a.handleLogin)
	mux.HandleFunc("/api/health", a.handleHealth)
	mux.HandleFunc("/api/zones", a.require(auth.ActionViewDashboard, a.handleZones))
	mux.HandleFunc("/api/dashboard", a.require(auth.ActionViewDashboard, a.handleDashboard))
	mux.HandleFunc("/api/alerts", a.require(auth.ActionReadAlert, a.handleAlerts))
	mux.HandleFunc("/api/alerts/", a.require(auth.ActionReadAlert, a.handleAlertPath))
	mux.HandleFunc("/api/readings", a.require(auth.ActionManageDevice, a.handleReading))
	mux.HandleFunc("/api/door-events", a.require(auth.ActionManageDevice, a.handleDoorEvent))
	mux.HandleFunc("/api/export/alerts", a.require(auth.ActionExport, a.handleAlertExport))
	mux.HandleFunc("/api/reports/compliance", a.require(auth.ActionExport, a.handleComplianceReport))
	return requestLogger(mux)
}

func (a *API) require(action auth.Action, next http.HandlerFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		session, err := a.session(request)
		if err != nil {
			writeError(writer, http.StatusUnauthorized, err)
			return
		}
		if err := auth.Require(session.Role, action); err != nil {
			writeError(writer, http.StatusForbidden, err)
			return
		}
		next(writer, request)
	}
}

func (a *API) session(request *http.Request) (domain.Session, error) {
	value := strings.TrimSpace(request.Header.Get("Authorization"))
	if !strings.HasPrefix(value, "Bearer ") {
		return domain.Session{}, authError{}
	}
	token := strings.TrimSpace(strings.TrimPrefix(value, "Bearer "))
	session, err := a.auth.Authenticate(token)
	if err != nil {
		return domain.Session{}, err
	}
	return session, nil
}

type authError struct{}

func (authError) Error() string { return "authorization header is required" }

func requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("X-Coldchain-Service", "temperature-alert")
		next.ServeHTTP(writer, request)
	})
}
