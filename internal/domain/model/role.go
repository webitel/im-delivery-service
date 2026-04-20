package model

// ThreadRole defines the participation level in a chat thread.
type ThreadRole int32

const (
	RoleUnspecified ThreadRole = iota // 0
	RoleMember                        // 1
	RoleAdmin                         // 2
	RoleOwner                         // 3
	RoleSupervisor                    // 4
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

// ParseRole safely converts an integer (e.g. from Protobuf/AMQP) to ThreadRole.
func ParseRole(v int32) ThreadRole {
	if v < 0 || v > 4 {
		return RoleUnspecified
	}
	return ThreadRole(v)
}
