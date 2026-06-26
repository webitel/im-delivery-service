package wsmarshaller

// [ENUM] EventPriority matches the Protobuf enumeration for delivery urgency.
//
//go:generate stringer -type=EventPriority -linecomment
type EventPriority int32

const (
	// [STRICT_TYPING] Explicit iota for proto-compatibility.
	PriorityUnspecified EventPriority = iota // unspecified
	PriorityHigh                             // high
	PriorityNormal                           // normal
	PriorityLow                              // low
)

// [ENUM] EventType defines the unique key used in the JSON payload map.
// This ensures that transport keys remain consistent across all sessions.
//
//go:generate stringer -type=EventType -linecomment
type EventType int32

const (
	// [TYPES] These comments define the exact string key in the JSON output.
	EventConnected           EventType = iota // connected_event
	EventDisconnected                         // disconnected_event
	EventMessage                              // message_event
	EventThreadCreated                        // thread_created_event
	EventAck                                  // ack_event
	EventError                                // error_event
	EventPing                                 // ping_event
	EventMemberAdded                          // member_added_event
	EventMemberLeft                           // member_left_event
	EventInteractiveCallback                  // interactive_callback
)

// [ENVELOPE] ServerEvent is the top-level WebSocket JSON container.
type ServerEvent struct {
	ID        string        `json:"id"`
	CreatedAt int64         `json:"created_at"`
	Priority  EventPriority `json:"priority"`
	// Payload keys are populated using EventType.String() for type-safety.
	Payload map[string]any `json:"payload"`
}
