package model

import (
	"fmt"

	"github.com/google/uuid"
)

// MessageReaction tells connected clients about an emoji reaction on a message.
// It is a distinct type from Message on purpose: the marshallers dispatch on
// the payload's Go type, so reusing Message would emit the reaction under the
// regular message key and clients would render it incorrectly.
type MessageReaction struct {
	ID        uuid.UUID `json:"id"`
	ThreadID  uuid.UUID `json:"thread_id"`
	DomainID  int64     `json:"domain_id"`
	Reactor   Peer      `json:"reactor"`
	Emoji     string    `json:"emoji"`
	Removed   bool      `json:"removed"`
	ReactedAt int64     `json:"reacted_at"`
	SendId    string    `json:"send_id"`

	// To is the recipient set used for fan-out only; it never reaches clients.
	To []Peer `json:"-"`
}

func (m *MessageReaction) RoutingKey() string {
	return fmt.Sprintf("im_delivery.v1.%d.message.reaction", m.DomainID)
}
