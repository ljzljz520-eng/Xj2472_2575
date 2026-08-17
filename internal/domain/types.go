package domain

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrNotFound      = errors.New("record not found")
	ErrUnauthorized  = errors.New("role is not authorized")
	ErrInvalidInput  = errors.New("invalid input")
	ErrAlreadyExists = errors.New("record already exists")
	ErrInvalidState  = errors.New("invalid state transition")
	ErrStorageClosed = errors.New("storage is closed")
)

type Role string

const (
	RoleWarehouse Role = "warehouse"
	RoleQuality   Role = "quality"
	RoleVisitor   Role = "visitor"
)

func (r Role) Valid() bool {
	switch r {
	case RoleWarehouse, RoleQuality, RoleVisitor:
		return true
	default:
		return false
	}
}

func (r Role) CanManageDevices() bool { return r == RoleWarehouse }
func (r Role) CanReviewAlerts() bool  { return r == RoleQuality }
func (r Role) CanExport() bool        { return r == RoleWarehouse || r == RoleQuality }

type Session struct {
	Token    string `json:"token"`
	Username string `json:"username"`
	Role     Role   `json:"role"`
}

type Zone struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	TargetMin   float64 `json:"target_min"`
	TargetMax   float64 `json:"target_max"`
	HumidityMin float64 `json:"humidity_min"`
	HumidityMax float64 `json:"humidity_max"`
	Enabled     bool    `json:"enabled"`
}

type Device struct {
	ID         string `json:"id"`
	ZoneID     string `json:"zone_id"`
	Name       string `json:"name"`
	Model      string `json:"model"`
	Online     bool   `json:"online"`
	LastSeen   string `json:"last_seen"`
	BatteryPct int    `json:"battery_pct"`
}

type TemperatureReading struct {
	ID          string  `json:"id"`
	DeviceID    string  `json:"device_id"`
	ZoneID      string  `json:"zone_id"`
	Temperature float64 `json:"temperature"`
	Humidity    float64 `json:"humidity"`
	DoorCount   int     `json:"door_count"`
	RecordedAt  string  `json:"recorded_at"`
}

type AlertSeverity string

const (
	SeverityInfo     AlertSeverity = "info"
	SeverityWarning  AlertSeverity = "warning"
	SeverityCritical AlertSeverity = "critical"
)

func (s AlertSeverity) Weight() int {
	switch s {
	case SeverityCritical:
		return 3
	case SeverityWarning:
		return 2
	case SeverityInfo:
		return 1
	default:
		return 0
	}
}

type AlertStatus string

const (
	AlertOpen         AlertStatus = "open"
	AlertAcknowledged AlertStatus = "acknowledged"
	AlertUnderReview  AlertStatus = "under_review"
	AlertResolved     AlertStatus = "resolved"
	AlertDismissed    AlertStatus = "dismissed"
)

type TemperatureAlert struct {
	ID          string        `json:"id"`
	ZoneID      string        `json:"zone_id"`
	DeviceID    string        `json:"device_id"`
	ReadingID   string        `json:"reading_id"`
	Severity    AlertSeverity `json:"severity"`
	Status      AlertStatus   `json:"status"`
	Reason      string        `json:"reason"`
	Temperature float64       `json:"temperature"`
	Humidity    float64       `json:"humidity"`
	OpenedAt    string        `json:"opened_at"`
	UpdatedAt   string        `json:"updated_at"`
	AssignedTo  string        `json:"assigned_to"`
}

type AlertReview struct {
	ID         string `json:"id"`
	AlertID    string `json:"alert_id"`
	Reviewer   string `json:"reviewer"`
	Decision   string `json:"decision"`
	Note       string `json:"note"`
	ReviewedAt string `json:"reviewed_at"`
}

type DoorEvent struct {
	ID         string `json:"id"`
	ZoneID     string `json:"zone_id"`
	DeviceID   string `json:"device_id"`
	Opened     bool   `json:"opened"`
	Duration   int    `json:"duration_seconds"`
	RecordedAt string `json:"recorded_at"`
}

type AuditEntry struct {
	ID        string `json:"id"`
	Actor     string `json:"actor"`
	Action    string `json:"action"`
	Entity    string `json:"entity"`
	EntityID  string `json:"entity_id"`
	Detail    string `json:"detail"`
	CreatedAt string `json:"created_at"`
}

type DashboardSummary struct {
	Zone              Zone                `json:"zone"`
	LatestReading     *TemperatureReading `json:"latest_reading"`
	TemperatureSeries []SeriesPoint       `json:"temperature_series"`
	HumiditySeries    []SeriesPoint       `json:"humidity_series"`
	OpenAlertCount    int                 `json:"open_alert_count"`
	CriticalCount     int                 `json:"critical_count"`
	DoorOpenCount     int                 `json:"door_open_count"`
	OnlineDevices     int                 `json:"online_devices"`
	TotalDevices      int                 `json:"total_devices"`
	HealthScore       int                 `json:"health_score"`
}

type SeriesPoint struct {
	At    string  `json:"at"`
	Value float64 `json:"value"`
}

type AlertFilter struct {
	ZoneID   string
	Status   AlertStatus
	Severity AlertSeverity
}

type ExportRow struct {
	AlertID   string
	Zone      string
	Severity  string
	Status    string
	Reason    string
	OpenedAt  string
	UpdatedAt string
}

func NewID(prefix string, number int) string {
	return fmt.Sprintf("%s-%04d", strings.ToLower(strings.TrimSpace(prefix)), number)
}

func StatusTerminal(status AlertStatus) bool {
	return status == AlertResolved || status == AlertDismissed
}

func IsOpen(status AlertStatus) bool {
	return status == AlertOpen || status == AlertAcknowledged || status == AlertUnderReview
}
