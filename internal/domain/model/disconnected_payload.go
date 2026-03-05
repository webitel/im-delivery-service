package model

// [DISCONNECTED_PAYLOAD] NOTIFICATION SENT BEFORE THE SERVER CLOSES THE STREAM
// It helps the client to decide whether to reconnect immediately or wait.
type DisconnectedPayload struct {
	// [REASON] A human-readable message describing why the session ended.
	Reason string `json:"reason"`

	// [CODE] Standard HTTP Status Code or internal service code.
	// 401: Unauthorized, 408: Request Timeout, 429: Too Many Requests, 503: Service Unavailable.
	Code int `json:"code,omitempty"`

	// [STATUS] A machine-readable string identifier for programmatic handling.
	// Examples: "SHUTDOWN", "EVICTED", "TIMEOUT", "AUTH_EXPIRED".
	Status string `json:"status,omitempty"`
}

// [CONSTANTS] COMMON DISCONNECTION REASONS
const (
	StatusShutdown = "SHUTDOWN"
	StatusEvicted  = "EVICTED"
	StatusTimeout  = "TIMEOUT"
	UNAUTHORIZED   = "UNAUTHORIZED"
)
