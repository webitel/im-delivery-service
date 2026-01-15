package model

import "github.com/google/uuid"

// EventKind defines the type of system or business event.
type EventKind int16

const (
	Connected EventKind = iota + 1
	MessageCreated
)

// EventPriority controls the backpressure strategy.
type EventPriority int32

const (
	PriorityLow    EventPriority = 10
	PriorityNormal EventPriority = 20
	PriorityHigh   EventPriority = 30
)

// Eventer represents the shared interface for all data flowing through the hub.
type Eventer interface {
	GetID() string
	GetTraceID() string
	GetKind() EventKind
	GetUserID() uuid.UUID
	GetPriority() EventPriority
	GetOccurredAt() int64
	GetPayload() any
	// [PERFORMANCE] GetCached returns pre-marshaled data (e.g., Protobuf message).
	GetCached() any
	// [PERFORMANCE] SetCached stores pre-marshaled data to avoid redundant computations.
	SetCached(any)
}
