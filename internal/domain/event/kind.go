package event

// [EVENT_KIND] Efficient numeric representation for transport.
type EventKind int16

//go:generate stringer -type=EventKind
const (
	Connected         EventKind = iota + 1 // [SYSTEM]
	DisconnectedEvent                      // [SYSTEM]
	MessageCreated                         // [BUSINESS]
	ThreadCreated                          // [BUSINESS]
	MessageRead                            // [BUSINESS]
)
