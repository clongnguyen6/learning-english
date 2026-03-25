package policies

import "strings"

const (
	RoleLearner = "learner"
	RoleAdmin   = "admin"
)

func NormalizeRole(role string) string {
	return strings.ToLower(strings.TrimSpace(role))
}

func IsSupportedRole(role string) bool {
	switch NormalizeRole(role) {
	case RoleLearner, RoleAdmin:
		return true
	default:
		return false
	}
}

func HasAnyRole(role string, allowedRoles ...string) bool {
	normalizedRole := NormalizeRole(role)
	if !IsSupportedRole(normalizedRole) {
		return false
	}

	for _, allowedRole := range allowedRoles {
		if normalizedRole == NormalizeRole(allowedRole) {
			return true
		}
	}

	return false
}
