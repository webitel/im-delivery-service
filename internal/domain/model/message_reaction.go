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

	// Reactions is the authoritative per-emoji aggregate on the message AFTER
	// this change, letting a client replace its reaction state instead of
	// applying the (Emoji, Removed) delta. reacted_by_me is derived client-side
	// from ReactorIDs.
	Reactions []ReactionAggregate `json:"reactions,omitempty"`

	// To is the recipient set used for fan-out only; it never reaches clients.
	To []Peer `json:"-"`
}

// ReactionAggregate is one emoji's state on a message: how many members hold it
// and a capped sample of their contact ids.
type ReactionAggregate struct {
	Emoji         string   `json:"emoji"`
	Count         int32    `json:"count"`
	ReactorIDs    []string `json:"reactor_ids"`
	LastReactedAt int64    `json:"last_reacted_at"`
}

func (m *MessageReaction) RoutingKey() string {
	return fmt.Sprintf("im_delivery.v1.%d.message.reaction", m.DomainID)
}
