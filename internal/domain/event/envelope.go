package event

import (
	"sync/atomic"

	"github.com/google/uuid"
)

// [ENVELOPE] Generic event container. Exported for JSON persistence.
type Envelope[T any] struct {
	ID         uuid.UUID         `json:"id"`
	Payload    T                 `json:"payload"`
	UserID     uuid.UUID         `json:"user_id"`
	DomainID   int64             `json:"domain_id"`
	Kind       EventKind         `json:"kind"`
	Priority   EventPriority     `json:"priority"`
	Metadata   map[string]string `json:"metadata,omitempty"`
	CanPush    bool              `json:"can_push"`
	Echo       bool              `json:"echo"`
	OccurredAt int64             `json:"occurred_at"`
	TraceID    string            `json:"trace_id"`

	// [INTERNAL_CACHE] Temporary storage for transport-level data.
	cached atomic.Value `json:"-"`
}

func (e *Envelope[T]) GetID() string                  { return e.ID.String() }
func (e *Envelope[T]) GetKind() EventKind             { return e.Kind }
func (e *Envelope[T]) GetKindName() string            { return e.Kind.String() }
func (e *Envelope[T]) GetUserID() uuid.UUID           { return e.UserID }
func (e *Envelope[T]) GetPriority() EventPriority     { return e.Priority }
func (e *Envelope[T]) GetPayload() any                { return e.Payload }
func (e *Envelope[T]) GetMetadata() map[string]string { return e.Metadata }
func (e *Envelope[T]) IsEcho() bool                   { return e.Echo }
func (e *Envelope[T]) GetOccurredAt() int64           { return e.OccurredAt }
func (e *Envelope[T]) IsPushable() bool               { return e.CanPush }
func (e *Envelope[T]) GetCached() any                 { return e.cached.Load() }
func (e *Envelope[T]) SetCached(v any)                { e.cached.Store(v) }
