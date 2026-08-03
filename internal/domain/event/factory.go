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

// [NEW_MESSAGE_EDITED_EVENT] Factory for message edits. Unlike a new message,
// an edit updates an already-delivered message in place (clients match by ID),
// so it is not pushed as a notification (CanPush=false) and its timestamp is the
// edit time.
func NewMessageEditedEvent(
	msg *model.Message,
	targetID uuid.UUID,
	opts ...Option[*model.Message],
) Eventer {
	e := &Envelope[*model.Message]{
		ID:         uuid.New(),
		Payload:    msg,
		UserID:     targetID,
		DomainID:   msg.DomainID,
		Kind:       MessageEdited,
		Priority:   PriorityHigh,
		OccurredAt: msg.EditedAt,
		CanPush:    false,
	}

	for _, apply := range opts {
		apply(e)
	}

	return e
}

// [NEW_MESSAGE_DELETED_EVENT] Factory for message deletions. Clients match the
// message by ID and remove it, so it is not pushed as a notification.
func NewMessageDeletedEvent(
	msg *model.MessageDeleted,
	targetID uuid.UUID,
	opts ...Option[*model.MessageDeleted],
) Eventer {
	e := &Envelope[*model.MessageDeleted]{
		ID:         uuid.New(),
		Payload:    msg,
		UserID:     targetID,
		DomainID:   msg.DomainID,
		Kind:       MessageDeleted,
		Priority:   PriorityHigh,
		OccurredAt: msg.DeletedAt,
		CanPush:    false,
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

// [NEW_MESSAGE_STATUS_EVENT] Factory for per-recipient delivery status
// change notifications. Not pushable: status marks only matter to clients
// that render the chat in real time.
func NewMessageStatusEvent(
	update *model.MessageStatusUpdate,
	targetID uuid.UUID,
	domainID int64,
) Eventer {
	occurredAt := update.OccurredAt
	if occurredAt <= 0 {
		occurredAt = time.Now().UnixMilli()
	}

	return &Envelope[*model.MessageStatusUpdate]{
		ID:         uuid.New(),
		Payload:    update,
		UserID:     targetID,
		DomainID:   domainID,
		Kind:       MessageStatusChanged,
		Priority:   PriorityNormal,
		OccurredAt: occurredAt,
	}
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
		Kind:       InteractiveCallback,
	}
}

// [NEW_TYPING_EVENT] Factory for the ephemeral "…is typing" indicator. It is
// real-time only: CanPush=false keeps it out of the push pipeline, and it is
// never persisted or replayed on reconnect.
func NewTypingEvent(t *model.Typing, targetID uuid.UUID, domainID, occurredAt int64) Eventer {
	if occurredAt <= 0 {
		occurredAt = time.Now().UnixMilli()
	}

	return &Envelope[*model.Typing]{
		ID:         uuid.New(),
		Payload:    t,
		UserID:     targetID,
		DomainID:   domainID,
		Kind:       Typing,
		Priority:   PriorityNormal,
		CanPush:    false,
		OccurredAt: occurredAt,
	}
}
