package payload

import (
	"github.com/webitel/im-delivery-service/internal/domain/model"
	"github.com/webitel/im-delivery-service/internal/domain/util"
)

// MessageDeletedV1 is the im_message.<thread_id>.message.deleted.v1 event
// published by im-thread-service. It carries no content by design.
type MessageDeletedV1 struct {
	MessageID  string      `json:"message_id"`
	ThreadID   string      `json:"thread_id"`
	DomainID   int32       `json:"domain_id"`
	DeletedBy  Peer        `json:"deleted_by"`
	To         []Recipient `json:"to"`
	Type       string      `json:"type"`
	CreatedAt  string      `json:"created_at"`
	OccurredAt string      `json:"occurred_at"`
}

func (d *MessageDeletedV1) ToDomain() *model.MessageDeleted {
	var createdAt int64
	if d.CreatedAt != "" {
		createdAt = util.SafeParseRFC3339(d.CreatedAt)
	}

	msg := &model.MessageDeleted{
		ID:        util.SafeParseUUID(d.MessageID),
		ThreadID:  util.SafeParseUUID(d.ThreadID),
		DomainID:  int64(d.DomainID),
		CreatedAt: createdAt,
		DeletedAt: util.SafeParseRFC3339(d.OccurredAt),
		DeletedBy: model.Peer{
			ID:       util.SafeParseUUID(d.DeletedBy.ContactID),
			MemberID: d.DeletedBy.MemberID,
			Role:     int32(d.DeletedBy.Role),
		},
	}

	msg.To = make([]model.Peer, 0, len(d.To))
	for _, r := range d.To {
		msg.To = append(msg.To, model.Peer{
			ID:       util.SafeParseUUID(r.ContactID),
			MemberID: r.MemberID,
			Role:     int32(r.Role),
		})
	}

	return msg
}
