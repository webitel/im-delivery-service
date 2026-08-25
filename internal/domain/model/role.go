package model

// ThreadRole defines the participation level in a chat thread.
type ThreadRole int32

// WARNING: due to unmatched logic between proto
// and thread-service implementation of roles
// we need to swap RoleOwner<->RoleSupervisor
const (
	RoleUnspecified ThreadRole = iota
	RoleMember
	RoleAdmin
	RoleSupervisor
	RoleOwner
)

// String returns the human-readable representation of the role.
// Useful for JSON output in WebSockets.
func (r ThreadRole) String() string {
	switch r {
	case RoleMember:
		return "ROLE_MEMBER"
	case RoleAdmin:
		return "ROLE_ADMIN"
	case RoleOwner:
		return "ROLE_OWNER"
	case RoleSupervisor:
		return "ROLE_SUPERVISOR"
	default:
		return "ROLE_UNSPECIFIED"
	}
}

// ParseRoleName converts a ThreadRole name (as carried by AMQP events) back to
// the enum. An unknown name degrades to RoleUnspecified.
func ParseRoleName(v string) ThreadRole {
	switch v {
	case "ROLE_MEMBER":
		return RoleMember
	case "ROLE_ADMIN":
		return RoleAdmin
	case "ROLE_SUPERVISOR":
		return RoleSupervisor
	case "ROLE_OWNER":
		return RoleOwner
	default:
		return RoleUnspecified
	}
}

// ParseRole safely converts an integer (e.g. from Protobuf/AMQP) to ThreadRole.
func ParseRole(v int32) ThreadRole {
	if v < 0 || v > 4 {
		return RoleUnspecified
	}

	return ThreadRole(v)
}
