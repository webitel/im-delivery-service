package event

import "github.com/google/uuid"

type EventKind int16

//go:generate stringer -type=EventKind
const (
	Connected      EventKind = iota + 1 // [SYSTEM]
	MessageCreated                      // [BUSINESS]
)

type EventPriority int32

const (
	PriorityLow    EventPriority = 10
	PriorityNormal EventPriority = 20
	PriorityHigh   EventPriority = 30
)

// Eventer defines the contract for all data packets flowing through the Hub.
type Eventer interface {
	GetID() string
	GetKind() EventKind
	GetUserID() uuid.UUID
	GetPriority() EventPriority
	GetOccurredAt() int64
	GetPayload() any
	GetCached() any
	SetCached(any)
}

// Exportable defines an event that should be re-published to the message bus.
type Exportable interface {
	// We return the key only if the event is ready to be exported.
	// If it returns an empty string, the binder will skip publishing.
	GetRoutingKey() string
}
