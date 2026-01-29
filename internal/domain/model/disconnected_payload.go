package model

// DisconnectedPayload represents the notification sent before the server closes the stream.
type DisconnectedPayload struct {
	Reason string `json:"reason"`
	Code   string `json:"code,omitempty"` // Optional: "SHUTDOWN", "EVICTED", "TIMEOUT"
}
