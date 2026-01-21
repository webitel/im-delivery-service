// internal/domain/model/message_v2_event.go
package model

import (
	"fmt"

	"github.com/google/uuid"
)

// Interface guard
var _ Eventer = (*MessageV2Event)(nil)

// MessageV2Event represents the enhanced V2 domain event
type MessageV2Event struct {
	message *Message
	userID  uuid.UUID
	cached  any
}

// NewMessageV2Event initializes the event with pre-resolved peers and domain entity
func NewMessageV2Event(msg *Message, userID uuid.UUID, from, to Peer) *MessageV2Event {
	msg.From = from
	msg.To = to
	return &MessageV2Event{
		message: msg,
		userID:  userID,
	}
}

func (e *MessageV2Event) GetID() string              { return e.message.ID.String() }
func (e *MessageV2Event) GetPayload() any            { return e.message }
func (e *MessageV2Event) GetUserID() uuid.UUID       { return e.userID }
func (e *MessageV2Event) GetOccurredAt() int64       { return e.message.CreatedAt }
func (e *MessageV2Event) GetKind() EventKind         { return MessageCreated }
func (e *MessageV2Event) GetPriority() EventPriority { return PriorityHigh }
func (e *MessageV2Event) GetCached() any             { return e.cached }
func (e *MessageV2Event) SetCached(v any)            { e.cached = v }

// GetRoutingKey for V2: im_delivery.message.v2.{sub}.{issuer}.{domain}.processed
func (e *MessageV2Event) GetRoutingKey() string {
	sub, issuer := e.message.From.GetRoutingParts()
	return fmt.Sprintf("im_delivery.message.v2.%s.%s.processed", sub, issuer)
}
