package models

// Роли пользователей.
const (
	RoleAkimatAdmin     = "AKIMAT_ADMIN"
	RoleTooAdmin        = "TOO_ADMIN"
	RoleContractorAdmin = "CONTRACTOR_ADMIN"
	RoleDriver          = "DRIVER"
)

// Типы организаций.
const (
	OrgTypeAkimat     = "AKIMAT"
	OrgTypeToo        = "TOO"
	OrgTypeContractor = "CONTRACTOR"
)

// IsAdmin проверяет, относится ли роль к административным.
func IsAdmin(role string) bool {
	switch role {
	case RoleAkimatAdmin, RoleTooAdmin, RoleContractorAdmin:
		return true
	default:
		return false
	}
}

// CanCreateOrganization определяет, может ли роль создавать организации заданного типа.
func CanCreateOrganization(role, orgType string) bool {
	switch role {
	case RoleAkimatAdmin:
		return orgType == OrgTypeToo
	case RoleTooAdmin:
		return orgType == OrgTypeContractor
	default:
		return false
	}
}

// IsAkimatAdmin проверяет, является ли роль администратором акимата.
func IsAkimatAdmin(role string) bool {
	return role == RoleAkimatAdmin
}

// IsTooAdmin проверяет, является ли роль администратором ТОО.
func IsTooAdmin(role string) bool {
	return role == RoleTooAdmin
}

// IsContractorAdmin проверяет, является ли роль администратором подрядчика.
func IsContractorAdmin(role string) bool {
	return role == RoleContractorAdmin
}

// IsDriver проверяет, является ли роль водителем.
func IsDriver(role string) bool {
	return role == RoleDriver
}
