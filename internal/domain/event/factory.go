package event

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/webitel/im-delivery-service/internal/domain/model"
)

// INTERFACE GUARD [MESSAGE_EVENT]
var (
	_ Eventer    = (*ExportableEnvelope[*model.Message])(nil)
	_ Exportable = (*ExportableEnvelope[*model.Message])(nil)
)

// INTERFACE GUARD [THREAD_EVENT]
var _ Eventer = (*Envelope[*model.Thread])(nil)

// INTERFACE GUARD [SYSTEM_EVENT]
var (
	_ Eventer = (*Envelope[*model.ConnectedPayload])(nil)
	_ Eventer = (*Envelope[*model.DisconnectedPayload])(nil)
)

// --- OPTIONS ---

func WithPriority[T any](p EventPriority) Option[T] {
	return func(e *Envelope[T]) { e.Priority = p }
}

func WithOccurredAt[T any](t int64) Option[T] {
	return func(e *Envelope[T]) { e.OccurredAt = t }
}

func WithTraceID[T any](traceID string) Option[T] {
	return func(e *Envelope[T]) { e.TraceID = traceID }
}

// --- FACTORIES ---

// [NEW_MESSAGE_EVENT] Returns an ExportableEnvelope.
func NewMessageEvent(msg *model.Message, targetID uuid.UUID, opts ...Option[*model.Message]) Eventer {
	e := &ExportableEnvelope[*model.Message]{
		Envelope: Envelope[*model.Message]{
			ID:         uuid.New(),
			Payload:    msg,
			UserID:     targetID,
			DomainID:   msg.DomainID,
			Kind:       MessageCreated,
			Priority:   PriorityHigh,
			OccurredAt: msg.CreatedAt,
		},
		RoutingKeyFunc: func(m *model.Message) string {
			peerType := "contact"
			if m.To != nil {
				issuer := strings.ToLower(m.To.Issuer)
				if strings.Contains(issuer, "bot") || strings.Contains(issuer, "schema") {
					peerType = "bot"
				}
			}
			sub := "unknown"
			if m.To != nil {
				sub = m.To.Sub
			}
			return fmt.Sprintf("im_delivery.v1.%d.%s.%s.message.created", m.DomainID, peerType, sub)
		},
	}

	for _, opt := range opts {
		opt(&e.Envelope)
	}
	return e
}

// [NEW_THREAD_EVENT] Returns a base Envelope (Not Exportable).
func NewThreadEvent(thread *model.Thread, targetID uuid.UUID, opts ...Option[*model.Thread]) Eventer {
	e := &Envelope[*model.Thread]{
		ID:         uuid.New(),
		Payload:    thread,
		UserID:     targetID,
		DomainID:   int64(thread.DomainID),
		Kind:       ThreadCreated,
		Priority:   PriorityNormal,
		OccurredAt: thread.CreatedAt,
	}

	for _, opt := range opts {
		opt(e)
	}
	return e
}

// [NEW_SYSTEM_EVENT] Returns a base Envelope (Not Exportable).
func NewSystemEvent[T any](userID uuid.UUID, kind EventKind, payload T, opts ...Option[T]) Eventer {
	e := &Envelope[T]{
		ID:         uuid.New(),
		Payload:    payload,
		UserID:     userID,
		Kind:       kind,
		Priority:   PriorityLow,
		OccurredAt: time.Now().UnixMilli(),
	}

	for _, opt := range opts {
		opt(e)
	}
	return e
}
