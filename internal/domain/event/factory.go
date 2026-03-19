package event

import (
	"time"

	"github.com/google/uuid"
	"github.com/webitel/im-delivery-service/internal/domain/model"
)

// [INTERFACE_GUARDS] Ensure Envelope correctly implements Eventer.
var _ Eventer = (*Envelope[any])(nil)

// [NEW_MESSAGE_EVENT] Factory for core messaging business logic.
func NewMessageEvent(
	msg *model.Message,
	targetID uuid.UUID,
	opts ...Option[*model.Message],
) Eventer {
	e := &Envelope[*model.Message]{
		id:         uuid.New(),
		payload:    msg,
		userID:     targetID,
		domainID:   msg.DomainID,
		kind:       MessageCreated,
		priority:   PriorityHigh,
		occurredAt: msg.CreatedAt,
		canPush:    true, // Messages are trackable by default.
	}

	for _, apply := range opts {
		apply(e)
	}
	return e
}

// [NEW_THREAD_EVENT] Factory for room/thread lifecycle.
func NewThreadEvent(
	thread *model.Thread,
	targetID uuid.UUID,
	opts ...Option[*model.Thread],
) Eventer {
	e := &Envelope[*model.Thread]{
		id:         uuid.New(),
		payload:    thread,
		userID:     targetID,
		domainID:   int64(thread.DomainID),
		kind:       ThreadCreated,
		priority:   PriorityNormal,
		occurredAt: thread.CreatedAt,
	}

	for _, apply := range opts {
		apply(e)
	}
	return e
}

// [NEW_SYSTEM_EVENT] Helper for generic system triggers.
func NewSystemEvent[T any](
	userID uuid.UUID,
	kind EventKind,
	payload T,
	opts ...Option[T],
) Eventer {
	e := &Envelope[T]{
		id:         uuid.New(),
		payload:    payload,
		userID:     userID,
		kind:       kind,
		priority:   PriorityLow,
		occurredAt: time.Now().UnixMilli(),
	}

	for _, apply := range opts {
		apply(e)
	}
	return e
}

// [NEW_READ_EVENT] Optimized read confirmation event.
func NewReadEvent(eventID, userID uuid.UUID) Eventer {
	return &Envelope[*model.MessageReadPayload]{
		id:         eventID, // Reuse original message ID for easy lookup.
		userID:     userID,
		kind:       MessageRead,
		priority:   PriorityLow,
		occurredAt: time.Now().UnixMilli(),
		payload: &model.MessageReadPayload{
			MessageID: eventID,
		},
	}
}
