package model

import "github.com/google/uuid"

type EventKind int16

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
	GetRoutingKey() string
	GetKind() EventKind
	GetUserID() uuid.UUID
	GetPriority() EventPriority
	GetOccurredAt() int64
	GetPayload() any
	GetCached() any
	SetCached(any)
}
