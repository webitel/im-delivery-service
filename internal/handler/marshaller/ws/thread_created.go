package wsmarshaller

import "github.com/webitel/im-delivery-service/internal/domain/model"

type Recipient struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Peer represents a participant in the transport layer.
type Peer struct {
	Type   string `json:"type"`
	Name   string `json:"name"`
	Sub    string `json:"sub,omitempty"`
	Issuer string `json:"iss,omitempty"`
}

// WSThread represents the thread structure for WebSocket transport.
type WSThread struct {
	ID        string    `json:"id"`
	DomainID  int32     `json:"domain_id"`
	CreatedAt int64     `json:"created_at"`
	Subject   string    `json:"subject"`
	Recipient Recipient `json:"recipient"`
	Type      string    `json:"type"`
	Members   []Peer    `json:"members,omitempty"`
}

// mapThread converts domain model to WebSocket DTO with members.
func mapThread(t *model.Thread) *WSThread {
	if t == nil {
		return nil
	}

	res := &WSThread{
		ID:        t.ID.String(),
		DomainID:  t.DomainID,
		CreatedAt: t.CreatedAt,
		Subject:   t.Subject,
		Type:      t.Type,
		Recipient: Recipient{
			ID:   t.Recipient.ID.String(),
			Name: t.Recipient.Name,
		},
	}

	// Map enriched members
	if len(t.Members) > 0 {
		res.Members = make([]Peer, len(t.Members))
		for i, p := range t.Members {
			res.Members[i] = Peer{
				Type:   p.Type.String(),
				Name:   p.Name,
				Sub:    p.Sub,
				Issuer: p.Issuer,
			}
		}
	}

	return res
}
