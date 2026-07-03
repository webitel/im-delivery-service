package model

import (
	"fmt"

	"github.com/google/uuid"
)

// BotControlReleased is the delivery-side representation of a bot losing control of a
// thread (forwarded from im-thread-service). It is broadcast so flow_manager can stop
// the running schema for the thread.
type BotControlReleased struct {
	ThreadID     uuid.UUID  `json:"thread_id"`
	DomainID     int64      `json:"domain_id"`
	MemberID     uuid.UUID  `json:"member_id"`
	NextMemberID *uuid.UUID `json:"next_member_id,omitempty"`
	Reason       string     `json:"reason"`
	OccurredAt   int64      `json:"occurred_at"`
}

// RoutingKey routes the event to the im_delivery.broadcast exchange using the same
// "im_delivery.v1.<domain>.<...>" convention as message events.
func (b *BotControlReleased) RoutingKey() string {
	return fmt.Sprintf("im_delivery.v1.%d.bot.control.released", b.DomainID)
}
