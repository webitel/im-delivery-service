package event

type EventKind int16

//go:generate stringer -type=EventKind
const (
	Connected      EventKind = iota + 1 // [SYSTEM]
	Disconnected                        // [SYSTEM]
	MessageCreated                      // [BUSINESS]
	ThreadCreated                       // [BUSINESS]
	MessageRead                         // [BUSINESS]
)
