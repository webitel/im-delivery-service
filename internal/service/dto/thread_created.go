package dto

import (
	"github.com/webitel/im-delivery-service/internal/domain/model"
	"github.com/webitel/im-delivery-service/internal/domain/util"
)

type Recipient struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ThreadCreatedV1 struct {
	ThreadID  string    `json:"id"`
	DomainID  int32     `json:"domain_id"`
	CreatedAt string    `json:"created_id"`
	Subject   string    `json:"subject"`
	Recipient Recipient `json:"recipient"`
}

func (t *ThreadCreatedV1) ToDomain() *model.Thread {
	return &model.Thread{
		ID:        util.SafeParseUUID(t.ThreadID),
		DomainID:  t.DomainID,
		CreatedAt: util.SafeParseRFC3339(t.CreatedAt),
		Subject:   t.Subject,
		Recipient: model.Recipient{
			ID:   util.SafeParseUUID(t.Recipient.ID),
			Name: t.Recipient.Name,
		},
	}
}
