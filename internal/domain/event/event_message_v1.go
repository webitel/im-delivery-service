package event

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/webitel/im-delivery-service/internal/domain/model"
)

var (
	_ Eventer    = (*MessageV1Event)(nil)
	_ Exportable = (*MessageV1Event)(nil)
)

// MessageV1Event is a domain event wrapper that facilitates the "Fan-out" delivery pattern.
//
// [STRATEGY]
// It distinguishes between:
//   - [BUSINESS_PEERS] (message.From/To): Logical participants (The "Who").
//   - [ROUTING_TARGET] (UserID): The physical recipient of this event instance (The "Where").
//
// This allows "Stateless Horizontal Scaling" where every node can check
// hub.IsConnected(UserID) to decide if it should handle the delivery.
type MessageV1Event struct {
	ID       uuid.UUID
	Message  *model.Message `json:"message"`
	UserID   uuid.UUID      `json:"user_id"` // [PHYSICAL_RECIPIENT] Target user ID
	DomainID int64          `json:"domain_id"`
	Cached   any            `json:"-"` // [INTERNAL] Not for serialization
}

// NewMessageV1Event initializes the event and binds enriched peers.
//
// [NOTE] Even if the message is sent to a Group (message.To),
// the 'UserID' must be the ID of the individual subscriber.
func NewMessageV1Event(msg *model.Message, userID uuid.UUID, from, to model.Peer) *MessageV1Event {
	// Enrich the message entity with full Peer profiles (Name, Avatar, etc.)
	msg.From = from
	msg.To = to

	return &MessageV1Event{
		ID:       uuid.New(),
		Message:  msg,
		UserID:   userID, // Used by the Hub to find the local WebSocket connection
		DomainID: msg.DomainID,
	}
}

func (e *MessageV1Event) GetID() string              { return e.ID.String() }
func (e *MessageV1Event) GetPayload() any            { return e.Message }
func (e *MessageV1Event) GetUserID() uuid.UUID       { return e.UserID }
func (e *MessageV1Event) GetOccurredAt() int64       { return e.Message.CreatedAt }
func (e *MessageV1Event) GetKind() EventKind         { return MessageCreated }
func (e *MessageV1Event) GetPriority() EventPriority { return PriorityHigh }
func (e *MessageV1Event) GetCached() any             { return e.Cached }
func (e *MessageV1Event) SetCached(v any)            { e.Cached = v }

// GetRoutingKey generates RabbitMQ routing topic based on domain requirements.
// [PATTERN] im_delivery.v1.{domain_id}.{recipient_subject}.enriched
func (e *MessageV1Event) GetRoutingKey() string {
	return fmt.Sprintf("im_delivery.v1.%d.%s.message.created", e.Message.DomainID, e.Message.To.Sub)
}
