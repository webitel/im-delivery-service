package model

import (
	"fmt"

	"github.com/google/uuid"
)

// MessageEdited tells connected clients to replace the content of a message
// they already hold. It is a distinct type from Message on purpose: the
// marshallers dispatch on the payload's Go type, so reusing Message would emit
// the edit under the regular message key and clients would render it as a new
// message.
type MessageEdited struct {
	ID       uuid.UUID      `json:"id"`
	ThreadID uuid.UUID      `json:"thread_id"`
	DomainID int64          `json:"domain_id"`
	EditedBy Peer           `json:"edited_by"`
	To       []Peer         `json:"to,omitempty"`
	Text     string         `json:"text"`
	Type     string         `json:"type"`
	Metadata map[string]any `json:"metadata,omitempty"`

	// Version is the position of this body in the message's change history, the
	// number GetMessageRevisions reports for it: 2 right after the first edit.
	// Clients order concurrent edits by it.
	Version   int32 `json:"version"`
	CreatedAt int64 `json:"created_at"`
	EditedAt  int64 `json:"edited_at"`
}

func (m *MessageEdited) RoutingKey() string {
	return fmt.Sprintf("im_delivery.v1.%d.message.edited", m.DomainID)
}
