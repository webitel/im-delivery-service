package payload

import (
	"github.com/webitel/im-delivery-service/internal/domain/model"
	"github.com/webitel/im-delivery-service/internal/domain/util"
)

type MessageEditedV1 struct {
	MessageID  string         `json:"message_id"`
	ThreadID   string         `json:"thread_id"`
	DomainID   int32          `json:"domain_id"`
	EditedBy   Peer           `json:"edited_by"`
	To         []Recipient    `json:"to"`
	Body       string         `json:"body"`
	Type       string         `json:"type"`
	Version    int32          `json:"version"`
	CreatedAt  string         `json:"created_at"`
	OccurredAt string         `json:"occurred_at"`
	Metadata   map[string]any `json:"metadata,omitempty"`
}

func (d *MessageEditedV1) ToDomain() *model.MessageEdited {
	editedAt := util.SafeParseRFC3339(d.OccurredAt)

	var createdAt int64
	if d.CreatedAt != "" {
		createdAt = util.SafeParseRFC3339(d.CreatedAt)
	}

	msg := &model.MessageEdited{
		ID:        util.SafeParseUUID(d.MessageID),
		ThreadID:  util.SafeParseUUID(d.ThreadID),
		DomainID:  int64(d.DomainID),
		Text:      d.Body,
		Type:      d.Type,
		Metadata:  d.Metadata,
		Version:   d.Version,
		CreatedAt: createdAt,
		EditedAt:  editedAt,
		EditedBy: model.Peer{
			ID:       util.SafeParseUUID(d.EditedBy.ContactID),
			MemberID: d.EditedBy.MemberID,
			Role:     int32(d.EditedBy.Role),
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
