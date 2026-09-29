package model

// ConnectedPayload represents the data sent to the client upon successful connection.
// Defining this as a struct fixes the "undefined" error and enables type-safe marshaling.
type ConnectedPayload struct {
	Ok            bool   `json:"ok"`
	ConnectionID  string `json:"connection_id"`
	ServerVersion string `json:"server_version"`
	// UpdatesCursor is the user's GetUpdates cursor now: equal to the saved one means nothing was missed.
	UpdatesCursor string `json:"updates_cursor,omitempty"`
}
