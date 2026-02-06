package wsmarshaller

// [ENUM] EventPriority matches the Proto enumeration.
type EventPriority int32

const (
	PriorityUnspecified EventPriority = iota // 0
	PriorityHigh                             // 1
	PriorityNormal                           // 2
	PriorityLow                              // 3
)

// [CONSTANTS] Event type keys unified with Protobuf 'oneof' field names.
const (
	EventConnected    = "connected_event"
	EventDisconnected = "disconnected_event"
	EventMessage      = "message_event"
	EventAck          = "ack_event"
	EventError        = "error_event"
	EventPing         = "ping_event"
)

// [ENVELOPE] ServerEvent is the top-level WebSocket JSON container.
type ServerEvent struct {
	ID        string         `json:"id"`
	CreatedAt int64          `json:"created_at"`
	Priority  EventPriority  `json:"priority"`
	Payload   map[string]any `json:"payload"`
}
