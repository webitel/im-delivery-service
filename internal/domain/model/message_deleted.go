package model

import (
	"fmt"

	"github.com/google/uuid"
)

// MessageDeleted tells connected clients to drop a message from a thread.
// It is a distinct type from Message on purpose: the marshallers dispatch on
// the payload's Go type, so reusing Message would emit the deletion under the
// regular message key and clients would render it as a new message.
type MessageDeleted struct {
	ID        uuid.UUID `json:"id"`
	ThreadID  uuid.UUID `json:"thread_id"`
	DomainID  int64     `json:"domain_id"`
	DeletedBy Peer      `json:"deleted_by"`
	CreatedAt int64     `json:"created_at"`
	DeletedAt int64     `json:"deleted_at"`

	// To is the recipient set used for fan-out only; it never reaches clients.
	To []Peer `json:"-"`
}

func (m *MessageDeleted) RoutingKey() string {
	return fmt.Sprintf("im_delivery.v1.%d.message.deleted", m.DomainID)
}
