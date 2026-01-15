package model

import "github.com/google/uuid"

type InboundEventKind int16

const (
	Connected      InboundEventKind = iota + 1 // [SYSTEM]
	MessageCreated                             // [BUSINESS]
)

type InboundEventPriority int32

const (
	PriorityLow    InboundEventPriority = 10
	PriorityNormal InboundEventPriority = 20
	PriorityHigh   InboundEventPriority = 30
)

// InboundEventer defines the contract for all data packets flowing through the Hub.
type InboundEventer interface {
	GetID() string
	GetTraceID() string
	GetKind() InboundEventKind
	GetUserID() uuid.UUID
	GetPriority() InboundEventPriority
	GetOccurredAt() int64
	GetPayload() any
	GetCached() any
	SetCached(any)
}
