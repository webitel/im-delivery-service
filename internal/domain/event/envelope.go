package event

import (
	"sync/atomic"

	"github.com/google/uuid"
)

// [INTERFACE GUARDS]
var (
	_ Eventer    = (*Envelope[any])(nil)
	_ IsPushable = (*Envelope[any])(nil)
)

type Option[T any] func(e *Envelope[T])

// [ENVELOPE] Generic event container. Fields are private for strict immutability.
type Envelope[T any] struct {
	id         uuid.UUID
	payload    T
	userID     uuid.UUID
	domainID   int64
	kind       EventKind
	priority   EventPriority
	occurredAt int64
	traceID    string
	canPush    bool
	echo       bool

	// [INTERNAL_CACHE] Thread-safe storage for transport-level data.
	cached atomic.Value
}

// [GETTERS]
func (e *Envelope[T]) GetID() string              { return e.id.String() }
func (e *Envelope[T]) GetKind() EventKind         { return e.kind }
func (e *Envelope[T]) GetUserID() uuid.UUID       { return e.userID }
func (e *Envelope[T]) GetPriority() EventPriority { return e.priority }
func (e *Envelope[T]) GetOccurredAt() int64       { return e.occurredAt }
func (e *Envelope[T]) GetPayload() any            { return e.payload }
func (e *Envelope[T]) CanPush() bool              { return e.canPush }
func (e *Envelope[T]) IsEcho() bool               { return e.echo }

// [CACHE_ACCESS]
func (e *Envelope[T]) GetCached() any  { return e.cached.Load() }
func (e *Envelope[T]) SetCached(v any) { e.cached.Store(v) }
