package model

// AuthContact represents the authenticated user session details.
type AuthContact struct {
	DC        int64  `json:"dc"`         // Domain ID
	ContactID string `json:"contact_id"` // Unique User/Contact ID (UUID string)
	Sub       string `json:"sub"`        // Subject (usually same as ContactID or username)
	Iss       string `json:"iss"`        // Issuer
	Name      string `json:"name"`       // Display name

	// [PUSH_INTEGRATION]
	// Devices contains all registered push tokens for this user.
	// This list is synchronized with Redis Presence during WS handshake.
	Devices []Device `json:"devices,omitempty"`
}
