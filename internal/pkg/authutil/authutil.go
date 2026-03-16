// Package authutil provides shared role-level utilities for permission checks.
package authutil

// RoleLevel returns a numeric privilege level for a room role.
// Higher number = higher privilege.
//
//	owner → 3, admin → 2, member → 1, unknown → 0
func RoleLevel(role string) int {
	switch role {
	case "owner":
		return 3
	case "admin":
		return 2
	case "member":
		return 1
	default:
		return 0
	}
}

// SystemRoleLevel returns a numeric privilege level for a system role.
//
//	owner → 3, admin → 2, user → 1
func SystemRoleLevel(role string) int {
	switch role {
	case "owner":
		return 3
	case "admin":
		return 2
	default:
		return 1
	}
}
