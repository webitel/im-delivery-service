// internal/domain/model/message_v2_event.go
package event

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/webitel/im-delivery-service/internal/domain/model"
)

// Interface guard
var (
	_ Eventer    = (*MessageCreatedV2Event)(nil)
	_ Exportable = (*MessageCreatedV1Event)(nil)
)

// MessageCreatedV2Event represents the enhanced V2 domain event
type MessageCreatedV2Event struct {
	ID      uuid.UUID
	message *model.Message
	userID  uuid.UUID
	cached  any
}

// NewMessageCreatedV2Event initializes the event with pre-resolved peers and domain entity
func NewMessageCreatedV2Event(msg *model.Message, userID uuid.UUID, from, to model.Peer) *MessageCreatedV2Event {
	msg.From = from
	msg.To = to
	return &MessageCreatedV2Event{
		ID:      uuid.New(),
		message: msg,
		userID:  userID,
	}
}

func (e *MessageCreatedV2Event) GetID() string              { return e.ID.String() }
func (e *MessageCreatedV2Event) GetPayload() any            { return e.message }
func (e *MessageCreatedV2Event) GetUserID() uuid.UUID       { return e.userID }
func (e *MessageCreatedV2Event) GetOccurredAt() int64       { return e.message.CreatedAt }
func (e *MessageCreatedV2Event) GetKind() EventKind         { return MessageCreated }
func (e *MessageCreatedV2Event) GetPriority() EventPriority { return PriorityHigh }
func (e *MessageCreatedV2Event) GetCached() any             { return e.cached }
func (e *MessageCreatedV2Event) SetCached(v any)            { e.cached = v }

// GetRoutingKey for V2: im_delivery.message.v2.{sub}.{issuer}.{domain}.processed
func (e *MessageCreatedV2Event) GetRoutingKey() string {
	sub, issuer := e.message.From.GetRoutingParts()
	return fmt.Sprintf("im_delivery.message.v2.%s.%s.processed", sub, issuer)
}
