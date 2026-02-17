package event

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/webitel/im-delivery-service/internal/domain/model"
)

var (
	_ Eventer    = (*MessageCreatedV1Event)(nil)
	_ Exportable = (*MessageCreatedV1Event)(nil)
	_ Multicast  = (*MessageCreatedV1Event)(nil)
)

// MessageCreatedV1Event is a domain event wrapper that facilitates the "Fan-out" delivery pattern.
//
// [STRATEGY]
// It distinguishes between:
//   - [BUSINESS_PEERS] (message.From/To): Logical participants (The "Who").
//   - [ROUTING_TARGET] (UserID): The physical recipient of this event instance (The "Where").
//
// This allows "Stateless Horizontal Scaling" where every node can check
// hub.IsConnected(UserID) to decide if it should handle the delivery.
type MessageCreatedV1Event struct {
	ID       uuid.UUID
	Message  *model.Message `json:"message"`
	UserID   uuid.UUID      `json:"user_id"` // [PHYSICAL_RECIPIENT] Target user ID
	DomainID int64          `json:"domain_id"`
	Cached   any            `json:"-"` // [INTERNAL] Not for serialization
}

// NewMessageCreatedV1Event initializes the event and binds enriched peers.
//
// [NOTE] Even if the message is sent to a Group (message.To),
// the 'UserID' must be the ID of the individual subscriber.
func NewMessageCreatedV1Event(msg *model.Message, userID uuid.UUID, from, to model.Peer) *MessageCreatedV1Event {
	// Enrich the message entity with full Peer profiles (Name, Avatar, etc.)
	msg.From = from
	msg.To = to

	return &MessageCreatedV1Event{
		ID:       uuid.New(),
		Message:  msg,
		UserID:   userID, // Used by the Hub to find the local WebSocket connection
		DomainID: msg.DomainID,
	}
}

func (e *MessageCreatedV1Event) GetID() string              { return e.ID.String() }
func (e *MessageCreatedV1Event) GetPayload() any            { return e.Message }
func (e *MessageCreatedV1Event) GetUserID() uuid.UUID       { return e.UserID }
func (e *MessageCreatedV1Event) GetOccurredAt() int64       { return e.Message.CreatedAt }
func (e *MessageCreatedV1Event) GetKind() EventKind         { return MessageCreated }
func (e *MessageCreatedV1Event) GetPriority() EventPriority { return PriorityHigh }
func (e *MessageCreatedV1Event) GetCached() any             { return e.Cached }
func (e *MessageCreatedV1Event) SetCached(v any)            { e.Cached = v }

// GetRoutingKey generates RabbitMQ routing topic based on domain requirements.
// Pattern: im_delivery.v1.{domain_id}.{peer_type}.{subject}.message.created
func (e *MessageCreatedV1Event) GetRoutingKey() string {
	// Default peer type is contact
	peerType := "contact"

	// Normalize issuer to lowercase for reliable comparison
	issuer := strings.ToLower(e.Message.To.Issuer)

	// [STRATEGY] If issuer contains 'bot' or 'schema', classify as bot routing
	if strings.Contains(issuer, "bot") || strings.Contains(issuer, "schema") {
		peerType = "bot"
	}

	// [ROUTING] Build the final RabbitMQ topic string
	return fmt.Sprintf("im_delivery.v1.%d.%s.%s.message.created",
		e.Message.DomainID,
		peerType,
		e.Message.To.Sub,
	)
}

func (e *MessageCreatedV1Event) GetRecipients() []uuid.UUID {
	// If user send message to himself (draft messages, saved) - just save one ID to not duplicate
	if e.UserID == e.Message.From.ID {
		return []uuid.UUID{e.UserID}
	}

	return []uuid.UUID{e.UserID, e.Message.From.ID}
}
