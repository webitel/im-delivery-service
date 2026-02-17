package wsmarshaller

import "github.com/webitel/im-delivery-service/internal/domain/model"

type Recipient struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// [DTO] WSThread represents the thread structure for WebSocket transport.
type WSThread struct {
	ID        string    `json:"id"`
	DomainID  int32     `json:"domain_id"`
	CreatedAt int64     `json:"created_at"`
	Subject   string    `json:"subject"`
	Recipient Recipient `json:"recipient"` // [REQUIRED] Target owner of the thread
}

// [MAPPER] mapThread converts domain model to WebSocket DTO.
func mapThread(t *model.Thread) *WSThread {
	if t == nil {
		return nil
	}

	return &WSThread{
		ID:        t.ID.String(),
		DomainID:  t.DomainID,
		CreatedAt: t.CreatedAt,
		Subject:   t.Subject,
		Recipient: Recipient{
			ID:   t.Recipient.ID.String(),
			Name: t.Recipient.Name,
		},
	}
}
