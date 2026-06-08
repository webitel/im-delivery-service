package event

import (
	"time"

	"github.com/google/uuid"
	"github.com/webitel/im-delivery-service/internal/domain/model"
)

// [INTERFACE_GUARDS] Ensure Envelope correctly implements Eventer.
var _ Eventer = (*Envelope[any])(nil)

// [NEW_MESSAGE_EVENT] Factory for core messaging business logic.
// Automatically extracts metadata for push notifications to avoid circular imports.
func NewMessageEvent(
	msg *model.Message,
	targetID uuid.UUID,
	opts ...Option[*model.Message],
) Eventer {
	e := &Envelope[*model.Message]{
		ID:         uuid.New(),
		Payload:    msg,
		UserID:     targetID,
		DomainID:   msg.DomainID,
		Kind:       MessageCreated,
		Priority:   PriorityHigh,
		OccurredAt: msg.CreatedAt,
		CanPush:    true,
		Metadata: map[string]string{
			"sender_name": msg.NotificationTitle(),
			"text":        msg.NotificationBody(),
		},
	}

	for _, apply := range opts {
		apply(e)
	}
	return e
}

// [NEW_THREAD_EVENT] Factory for room/thread lifecycle events.
func NewThreadEvent(
	thread *model.Thread,
	targetID uuid.UUID,
	opts ...Option[*model.Thread],
) Eventer {
	e := &Envelope[*model.Thread]{
		ID:         uuid.New(),
		Payload:    thread,
		UserID:     targetID,
		DomainID:   int64(thread.DomainID),
		Kind:       ThreadCreated,
		Priority:   PriorityNormal,
		OccurredAt: thread.CreatedAt,
		Metadata: map[string]string{
			"sender_name": "System",
			"text":        "New chat conversation started",
		},
	}

	for _, apply := range opts {
		apply(e)
	}
	return e
}

// [NEW_SYSTEM_EVENT] Senior-level generic helper for internal triggers.
// Allows passing any T as payload while maintaining strict event contract.
func NewSystemEvent[T any](
	userID uuid.UUID,
	kind EventKind,
	payload T,
	opts ...Option[T],
) Eventer {
	e := &Envelope[T]{
		ID:         uuid.New(),
		Payload:    payload,
		UserID:     userID,
		Kind:       kind,
		Priority:   PriorityLow, // Default for system tasks
		OccurredAt: time.Now().UnixMilli(),
	}

	for _, apply := range opts {
		apply(e)
	}
	return e
}

// [NEW_MEMBER_EVENT] Factory for member lifecycle events (added/left).
func NewMemberEvent(m *model.MemberEvent, targetID uuid.UUID, kind EventKind) Eventer {
	return &Envelope[*model.MemberEvent]{
		ID:         uuid.New(),
		Payload:    m,
		UserID:     targetID,
		Kind:       kind,
		Priority:   PriorityNormal,
		OccurredAt: time.Now().UnixMilli(),
	}
}

// [NEW_READ_EVENT] Optimized factory for message read confirmations.
func NewReadEvent(eventID, userID uuid.UUID) Eventer {
	return &Envelope[*model.MessageReadPayload]{
		ID:         eventID,
		UserID:     userID,
		Kind:       MessageRead,
		Priority:   PriorityLow,
		OccurredAt: time.Now().UnixMilli(),
		Payload: &model.MessageReadPayload{
			MessageID: eventID,
		},
	}
}

func NewVariableEvent(
	payload *model.VariablesPayload,
	targetID uuid.UUID,
	domainID int64,
	kind EventKind, // Should be VariableSet or VariableFlush
) Eventer {
	return &Envelope[*model.VariablesPayload]{
		ID:         uuid.New(),
		Payload:    payload,
		UserID:     targetID,
		DomainID:   domainID,
		Kind:       kind,
		Priority:   PriorityLow,
		OccurredAt: time.Now().UnixMilli(),
	}
}

func NewInteractiveCallbackEvent(payload *model.InteractiveCallback) Eventer {
	return &Envelope[*model.InteractiveCallback]{
		ID:         uuid.New(),
		Payload:    payload,
		UserID:     payload.ReactedBy.ID,
		Priority:   PriorityLow,
		OccurredAt: time.Now().UTC().UnixMilli(),
	}
}
