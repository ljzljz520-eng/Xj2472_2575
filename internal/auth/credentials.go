package auth

import (
	"strings"

	"coldchain-alert/internal/domain"
)

func NormalizeRole(value string) (domain.Role, error) {
	role := domain.Role(strings.ToLower(strings.TrimSpace(value)))
	if err := domain.ValidateRole(role); err != nil {
		return "", err
	}
	return role, nil
}

func UserForRole(role domain.Role) User {
	switch role {
	case domain.RoleWarehouse:
		return User{Username: "warehouse", Display: "Warehouse Operator", Role: role}
	case domain.RoleQuality:
		return User{Username: "quality", Display: "Quality Inspector", Role: role}
	default:
		return User{Username: "visitor", Display: "Read Only Visitor", Role: role}
	}
}
