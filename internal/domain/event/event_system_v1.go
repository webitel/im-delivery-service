package event

import (
	"time"

	"github.com/google/uuid"
)

// [GUARD] Ensure compliance with the Eventer interface.
var _ Eventer = (*SystemEvent)(nil)

// SystemEvent is a generic envelope for internal signals and domain notifications.
type SystemEvent struct {
	id         string
	traceID    string
	userID     uuid.UUID
	kind       EventKind
	priority   EventPriority
	occurredAt int64
	payload    any
	cached     any // Atomic/Sync.Pool optimization for transport-specific serialization
}

// [INTERFACE_IMPLEMENTATION]
func (e *SystemEvent) GetID() string              { return e.id }
func (e *SystemEvent) GetTraceID() string         { return e.traceID }
func (e *SystemEvent) GetKind() EventKind         { return e.kind }
func (e *SystemEvent) GetUserID() uuid.UUID       { return e.userID }
func (e *SystemEvent) GetPriority() EventPriority { return e.priority }
func (e *SystemEvent) GetOccurredAt() int64       { return e.occurredAt }
func (e *SystemEvent) GetPayload() any            { return e.payload }
func (e *SystemEvent) GetCached() any             { return e.cached }
func (e *SystemEvent) SetCached(v any)            { e.cached = v }

// GetRoutingKey is used for message broker exchange logic.
func (e *SystemEvent) GetRoutingKey() string {
	return ""
}

// NewSystemEvent is a universal factory for creating any signal.
func NewSystemEvent(userID uuid.UUID, kind EventKind, priority EventPriority, payload any) *SystemEvent {
	return &SystemEvent{
		id:         uuid.NewString(),
		traceID:    uuid.NewString(),
		userID:     userID,
		kind:       kind,
		priority:   priority,
		occurredAt: time.Now().UnixMilli(),
		payload:    payload,
	}
}
