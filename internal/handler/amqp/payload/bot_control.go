package payload

import (
	"github.com/google/uuid"

	"github.com/webitel/im-delivery-service/internal/domain/model"
	"github.com/webitel/im-delivery-service/internal/domain/util"
)

// BotControlReleasedV1 mirrors the im-thread-service "im.thread.bot.control.released"
// event payload.
type BotControlReleasedV1 struct {
	ThreadID     string `json:"thread_id"`
	DomainID     int64  `json:"domain_id"`
	MemberID     string `json:"member_id"`
	Position     int    `json:"position"`
	Reason       string `json:"reason"`
	NextMemberID string `json:"next_member_id,omitempty"`
	OccurredAt   string `json:"occurred_at"`
}

func (b *BotControlReleasedV1) ToDomain() *model.BotControlReleased {
	out := &model.BotControlReleased{
		ThreadID: util.SafeParseUUID(b.ThreadID),
		DomainID: b.DomainID,
		MemberID: util.SafeParseUUID(b.MemberID),
		Reason:   b.Reason,
	}

	if b.NextMemberID != "" {
		if next := util.SafeParseUUID(b.NextMemberID); next != uuid.Nil {
			out.NextMemberID = &next
		}
	}

	return out
}
