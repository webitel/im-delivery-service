package event

import (
	"sync/atomic"

	"github.com/google/uuid"
)

// [OPTION_PATTERN]
type Option[T any] func(e *Envelope[T])

// [BASE_ENVELOPE]
// Implements Eventer. Used for local-only events like Threads or System signals.
type Envelope[T any] struct {
	ID         uuid.UUID     `json:"id"`
	Payload    T             `json:"payload"`
	UserID     uuid.UUID     `json:"user_id"`
	DomainID   int64         `json:"domain_id"`
	Kind       EventKind     `json:"kind"`
	Priority   EventPriority `json:"priority"`
	OccurredAt int64         `json:"occurred_at"`
	TraceID    string        `json:"trace_id,omitempty"`
	cached     atomic.Pointer[any]
}

// [EXPORTABLE_ENVELOPE]
// Wraps Envelope and implements Exportable. Used for global RabbitMQ events.
type ExportableEnvelope[T any] struct {
	Envelope[T] // [EMBEDDING]

	// RoutingKeyFunc defines the AMQP topic logic.
	RoutingKeyFunc func(payload T) string `json:"-"`
}

// [EVENTER_IMPLEMENTATION]
func (e *Envelope[T]) GetID() string              { return e.ID.String() }
func (e *Envelope[T]) GetPayload() any            { return e.Payload }
func (e *Envelope[T]) GetUserID() uuid.UUID       { return e.UserID }
func (e *Envelope[T]) GetKind() EventKind         { return e.Kind }
func (e *Envelope[T]) GetPriority() EventPriority { return e.Priority }
func (e *Envelope[T]) GetOccurredAt() int64       { return e.OccurredAt }
func (e *Envelope[T]) GetCached() any {
	if v := e.cached.Load(); v != nil {
		return *v
	}
	return nil
}
func (e *Envelope[T]) SetCached(v any) { e.cached.Store(&v) }

// [EXPORTABLE_IMPLEMENTATION]
// Only ExportableEnvelope has this method, satisfying the Exportable interface.
func (e *ExportableEnvelope[T]) GetRoutingKey() string {
	if e.RoutingKeyFunc != nil {
		return e.RoutingKeyFunc(e.Payload)
	}
	return ""
}
