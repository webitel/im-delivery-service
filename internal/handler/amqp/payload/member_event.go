package payload

import (
	"github.com/webitel/im-delivery-service/internal/domain/model"
	"github.com/webitel/im-delivery-service/internal/domain/util"
)

type MemberEventSystem struct {
	Type     string         `json:"type"`
	Metadata map[string]any `json:"metadata"`
}

type MemberEventV1 struct {
	ThreadID   string            `json:"thread_id"`
	DomainID   int64             `json:"domain_id"`
	ContactID  string            `json:"contact_id"`
	OccurredAt string            `json:"occurred_at"`
	System     MemberEventSystem `json:"system"`
}

func (m *MemberEventV1) ToDomain() *model.MemberEvent {
	return &model.MemberEvent{
		ThreadID:  util.SafeParseUUID(m.ThreadID),
		ContactID: util.SafeParseUUID(m.ContactID),
		Metadata:  m.System.Metadata,
	}
}
