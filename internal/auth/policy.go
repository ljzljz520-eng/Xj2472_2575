package auth

import "coldchain-alert/internal/domain"

type Action string

const (
	ActionViewDashboard Action = "view_dashboard"
	ActionReadAlert     Action = "read_alert"
	ActionManageDevice  Action = "manage_device"
	ActionReviewAlert   Action = "review_alert"
	ActionExport        Action = "export"
)

func Allowed(role domain.Role, action Action) bool {
	switch action {
	case ActionViewDashboard, ActionReadAlert:
		return role.Valid()
	case ActionManageDevice:
		return role.CanManageDevices()
	case ActionReviewAlert:
		return role.CanReviewAlerts()
	case ActionExport:
		return role.CanExport()
	default:
		return false
	}
}

func Require(role domain.Role, action Action) error {
	if !Allowed(role, action) {
		return domain.ErrUnauthorized
	}
	return nil
}
