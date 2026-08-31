package models

type Role string

const (
	RoleSuperAdmin Role = "super_admin"
	RoleAdmin      Role = "admin"
	RoleSales      Role = "sales"
	RoleFinance    Role = "finance"
)

func IsValidRole(role string) bool {
	switch Role(role) {
	case RoleSuperAdmin, RoleAdmin, RoleSales, RoleFinance:
		return true
	default:
		return false
	}
}

func IsTenantRole(role string) bool {
	switch Role(role) {
	case RoleAdmin, RoleSales, RoleFinance:
		return true
	default:
		return false
	}
}

func IsStaffRole(role string) bool {
	switch Role(role) {
	case RoleSales, RoleFinance:
		return true
	default:
		return false
	}
}
