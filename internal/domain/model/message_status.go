package model

import (
	"github.com/google/uuid"
)

// MessageStatusUpdate notifies thread participants that the per-recipient
// delivery status of one or more messages changed. Read receipts are bulk
// ("read up to"), so a single update may cover multiple messages of the
// same recipient in the same thread.
type MessageStatusUpdate struct {
	ThreadID uuid.UUID `json:"thread_id"`
	// MemberID is the recipient contact id whose statuses changed.
	MemberID uuid.UUID `json:"-"`
	// Member is that recipient in the same shape as a message sender.
	Member     *Peer       `json:"member,omitempty"`
	MessageIDs []uuid.UUID `json:"message_ids"`
	// Status is the new delivery state: delivered|read|failed.
	Status string `json:"status"`
	// Via is the confirmation source: ws|push|provider|bot.
	Via string `json:"via,omitempty"`
	// Error carries provider error details for failed statuses.
	Error map[string]any `json:"error,omitempty"`
	// OccurredAt is the status change time (Unix ms).
	OccurredAt int64 `json:"occurred_at"`
	// UpToMessageID is the highest message id covered; nil (omitted, not zeros)
	// when the producer sends only UpToSeq, as thread-service does.
	UpToMessageID *uuid.UUID `json:"up_to_message_id,omitempty"`
	// UpToSeq is the per-thread sequence number of the delivered/read-up-to boundary
	// (preferred watermark; supercedes UpToMessageID).
	UpToSeq int64 `json:"up_to_seq,omitempty"`
	// UpdatesCursor is the recipient's GetUpdates cursor after this change; reads only.
	UpdatesCursor string `json:"updates_cursor,omitempty"`
}

// EventMessageRef is the message context of a fan-out event envelope, kept
// so a client ACK (which references only the envelope id) can be resolved
// into a per-recipient MarkDelivered report for im-thread-service.
type EventMessageRef struct {
	MessageID uuid.UUID `json:"message_id"`
	ThreadID  uuid.UUID `json:"thread_id"`
	// MemberID is the recipient contact id the envelope was addressed to.
	MemberID uuid.UUID `json:"member_id"`
	DomainID int64     `json:"domain_id"`
}

// Failure is the provider's reason as {code, message}; ok is false when nothing failed.
func (m *MessageStatusUpdate) Failure() (code, message string, ok bool) {
	if len(m.Error) == 0 {
		return "", "", false
	}

	return errorText(m.Error, "code"), errorText(m.Error, "message"), true
}

func errorText(details map[string]any, key string) string {
	if v, ok := details[key].(string); ok {
		return v
	}

	return ""
}

// FailedMessageID is the message a failure is about; empty for delivered/read, which move a
// horizon (up_to_seq) instead of naming messages.
func (m *MessageStatusUpdate) FailedMessageID() string {
	if m.Status != "failed" || len(m.MessageIDs) == 0 {
		return ""
	}

	return m.MessageIDs[0].String()
}
