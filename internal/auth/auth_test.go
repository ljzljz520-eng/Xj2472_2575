package auth_test

import (
	"errors"
	"testing"

	"coldchain-alert/internal/auth"
	"coldchain-alert/internal/domain"
)

func TestRoleLoginAndPermissions(t *testing.T) {
	manager := auth.NewManager()
	session, err := manager.Login("quality", domain.RoleQuality)
	if err != nil || session.Token == "" {
		t.Fatalf("quality login failed: %#v %v", session, err)
	}
	if err := manager.Can(session.Token, func(role domain.Role) bool { return role.CanReviewAlerts() }); err != nil {
		t.Fatalf("quality review permission denied: %v", err)
	}
	if err := manager.Can(session.Token, func(role domain.Role) bool { return role.CanManageDevices() }); !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("quality manage permission should fail: %v", err)
	}
	if _, err := manager.Login("quality", domain.RoleWarehouse); !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("role mismatch should be unauthorized: %v", err)
	}
	if err := manager.Logout(session.Token); err != nil {
		t.Fatalf("logout failed: %v", err)
	}
}

func TestNormalizeRole(t *testing.T) {
	role, err := auth.NormalizeRole(" Warehouse ")
	if err != nil || role != domain.RoleWarehouse {
		t.Fatalf("role normalization failed: %s %v", role, err)
	}
	if _, err := auth.NormalizeRole("unknown"); !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("unknown role accepted: %v", err)
	}
}
