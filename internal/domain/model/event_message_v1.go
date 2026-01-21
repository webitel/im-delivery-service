// internal/domain/model/message_event.go
package model

import (
	"fmt"

	"github.com/google/uuid"
)

// Interface guard
var _ Eventer = (*MessageV1Event)(nil)

// MessageV1Event is a domain event wrapper that facilitates the "Fan-out" delivery pattern.
//
// [STRATEGY]
// It distinguishes between:
//   - [BUSINESS_PEERS] (message.From/To): Logical participants (The "Who").
//   - [ROUTING_TARGET] (userID): The physical recipient of this event instance (The "Where").
//
// This allows "Stateless Horizontal Scaling" where every node can check
// hub.IsConnected(userID) to decide if it should handle the delivery.
type MessageV1Event struct {
	message *Message
	userID  uuid.UUID // [PHYSICAL_RECIPIENT] Target user ID for infrastructure routing
	cached  any
}

// NewMessageV1Event initializes the event and binds enriched peers.
//
// [NOTE] Even if the message is sent to a Group (message.To),
// the 'userID' must be the ID of the individual subscriber.
func NewMessageV1Event(msg *Message, userID uuid.UUID, from, to Peer) *MessageV1Event {
	// Enrich the message entity with full Peer profiles (Name, Avatar, etc.)
	msg.From = from
	msg.To = to

	return &MessageV1Event{
		message: msg,
		userID:  userID, // Used by the Hub to find the local WebSocket connection
	}
}

func (e *MessageV1Event) GetID() string              { return e.message.ID.String() }
func (e *MessageV1Event) GetPayload() any            { return e.message }
func (e *MessageV1Event) GetUserID() uuid.UUID       { return e.userID }
func (e *MessageV1Event) GetOccurredAt() int64       { return e.message.CreatedAt }
func (e *MessageV1Event) GetKind() EventKind         { return MessageCreated }
func (e *MessageV1Event) GetPriority() EventPriority { return PriorityHigh }

func (e *MessageV1Event) GetCached() any  { return e.cached }
func (e *MessageV1Event) SetCached(v any) { e.cached = v }

// GetRoutingKey generates RabbitMQ routing topic: im_delivery.message.v1.{sub}.{issuer}.{domain}.processed
func (e *MessageV1Event) GetRoutingKey() string {
	sub, issuer := e.message.From.GetRoutingParts()
	return fmt.Sprintf("im_delivery.message.v1.%s.%s.processed", sub, issuer)
}
