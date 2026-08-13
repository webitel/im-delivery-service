package payload

import (
	"github.com/webitel/im-delivery-service/internal/domain/model"
	"github.com/webitel/im-delivery-service/internal/domain/util"
)

// MessageReactionV1 is the im_message.<thread_id>.message.reaction.v1 event
// published by im-thread-service. It notifies of emoji reactions on messages.
type MessageReactionV1 struct {
	MessageID  string      `json:"message_id"`
	ThreadID   string      `json:"thread_id"`
	DomainID   int32       `json:"domain_id"`
	Reactor    Peer        `json:"reactor"`
	To         []Recipient `json:"to"`
	Emoji      string      `json:"emoji"`
	Action     string      `json:"action"`
	OccurredAt string      `json:"occurred_at"`
	SendId     string      `json:"send_id"`
}

func (r *MessageReactionV1) ToDomain() *model.MessageReaction {
	msg := &model.MessageReaction{
		ID:        util.SafeParseUUID(r.MessageID),
		ThreadID:  util.SafeParseUUID(r.ThreadID),
		DomainID:  int64(r.DomainID),
		Emoji:     r.Emoji,
		Removed:   r.Action == "removed",
		ReactedAt: util.SafeParseRFC3339(r.OccurredAt),
		SendId:    r.SendId,
		Reactor: model.Peer{
			ID:       util.SafeParseUUID(r.Reactor.ContactID),
			MemberID: r.Reactor.MemberID,
			Role:     int32(r.Reactor.Role),
		},
	}

	msg.To = make([]model.Peer, 0, len(r.To))
	for _, rec := range r.To {
		msg.To = append(msg.To, model.Peer{
			ID:       util.SafeParseUUID(rec.ContactID),
			MemberID: rec.MemberID,
			Role:     int32(rec.Role),
		})
	}

	return msg
}
