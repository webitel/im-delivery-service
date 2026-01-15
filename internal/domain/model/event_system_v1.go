package model

import (
	"time"

	"github.com/google/uuid"
)

// Interface guard
var _ InboundEventer = (*SystemEvent)(nil)

// ConnectedPayload represents the data sent to the client upon successful connection.
// Defining this as a struct fixes the "undefined" error and enables type-safe marshalling.
type ConnectedPayload struct {
	Ok            bool   `json:"ok"`
	ConnectionID  string `json:"connection_id"`
	ServerVersion string `json:"server_version"`
}

// SystemEvent is a static container for service-generated signals.
type SystemEvent struct {
	ID         string
	TraceID    string
	UserID     uuid.UUID
	Kind       InboundEventKind
	Priority   InboundEventPriority
	OccurredAt int64
	Payload    any
	cached     any
}

func (e *SystemEvent) GetID() string                     { return e.ID }
func (e *SystemEvent) GetTraceID() string                { return e.TraceID }
func (e *SystemEvent) GetKind() InboundEventKind         { return e.Kind }
func (e *SystemEvent) GetUserID() uuid.UUID              { return e.UserID }
func (e *SystemEvent) GetPriority() InboundEventPriority { return e.Priority }
func (e *SystemEvent) GetOccurredAt() int64              { return e.OccurredAt }
func (e *SystemEvent) GetPayload() any                   { return e.Payload }
func (e *SystemEvent) GetCached() any                    { return e.cached }
func (e *SystemEvent) SetCached(v any)                   { e.cached = v }

// NewConnectedEvent creates a connection signal using the explicit struct.
func NewConnectedEvent(userID uuid.UUID, connID string, version string) *SystemEvent {
	return &SystemEvent{
		ID:         uuid.NewString(),
		TraceID:    uuid.NewString(),
		UserID:     userID,
		Kind:       Connected,
		Priority:   PriorityNormal,
		OccurredAt: time.Now().UnixMilli(),
		// Use the struct instead of a map
		Payload: &ConnectedPayload{
			Ok:            true,
			ConnectionID:  connID,
			ServerVersion: version,
		},
	}
}
