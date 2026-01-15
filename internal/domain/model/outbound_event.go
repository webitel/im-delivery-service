package model

import (
	"time"

	"github.com/google/uuid"
)

// OutboundEventer defines the contract for events that are being published
// from this service to the outside world (e.g., Message Delivery Receipts).
type OutboundEventer interface {
	GetRoutingKey() string
	GetExchange() string
	ToJSON() ([]byte, error)
}

// OutboundEvent is a concrete implementation for publishing.
type OutboundEvent struct {
	ID        string           `json:"id"`
	Source    string           `json:"source"` // e.g., "im-delivery-service"
	UserID    uuid.UUID        `json:"user_id"`
	Kind      InboundEventKind `json:"kind"`
	Payload   any              `json:"payload"`
	Timestamp int64            `json:"timestamp"`
}

// NewOutboundEvent creates a fresh event ready for publishing.
func NewOutboundEvent(userID uuid.UUID, kind InboundEventKind, payload any) *OutboundEvent {
	return &OutboundEvent{
		ID:        uuid.NewString(),
		Source:    "im-delivery-service",
		UserID:    userID,
		Kind:      kind,
		Payload:   payload,
		Timestamp: time.Now().UnixMilli(),
	}
}
