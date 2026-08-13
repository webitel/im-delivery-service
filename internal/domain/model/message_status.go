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
	MemberID   uuid.UUID   `json:"member_id"`
	MessageIDs []uuid.UUID `json:"message_ids"`
	// Status is the new delivery state: delivered|read|failed.
	Status string `json:"status"`
	// Via is the confirmation source: ws|push|provider|bot.
	Via string `json:"via,omitempty"`
	// Error carries provider error details for failed statuses.
	Error map[string]any `json:"error,omitempty"`
	// OccurredAt is the status change time (Unix ms).
	OccurredAt int64 `json:"occurred_at"`
	// UpToMessageID is the watermark: highest message id covered by this
	// status change (delivered/read-up-to).
	UpToMessageID uuid.UUID `json:"up_to_message_id,omitempty"`
	// UpToSeq is the per-thread sequence number of the delivered/read-up-to boundary
	// (preferred watermark; supercedes UpToMessageID).
	UpToSeq int64 `json:"up_to_seq,omitempty"`
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
