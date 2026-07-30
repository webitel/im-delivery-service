package event

// [EVENT_KIND] Efficient numeric representation for transport.
type EventKind int16

//go:generate stringer -type=EventKind
const (
	Connected            EventKind = iota + 1 // [SYSTEM]
	DisconnectedEvent                         // [SYSTEM]
	MessageCreated                            // [BUSINESS]
	ThreadCreated                             // [BUSINESS]
	MessageRead                               // [BUSINESS]
	VariableSet                               // [BUSINESS]
	VariableFlush                             // [BUSINESS]
	MemberAdded                               // [BUSINESS]
	MemberLeft                                // [BUSINESS]
	InteractiveCallback                       // [BUSINESS]
	MessageEdited                             // [BUSINESS]
	MessageStatusChanged                      // [BUSINESS]
	MessageDeleted                            // [BUSINESS]
	Typing                                    // [BUSINESS] ephemeral, real-time only
)
