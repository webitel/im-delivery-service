package wsmarshaller

import "github.com/webitel/im-delivery-service/internal/domain/model"

// WSThread represents the thread structure for WebSocket transport.
type WSThread struct {
	ID        string   `json:"id"`
	DomainID  int32    `json:"domain_id"`
	CreatedAt int64    `json:"created_at"`
	Subject   string   `json:"subject"`
	Type      string   `json:"type"`
	Members   []WSPeer `json:"members,omitempty"`
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
	}

	// Mapping members with the same nested logic as in mapMessage
	if len(t.Members) > 0 {
		res.Members = make([]WSPeer, 0, len(t.Members))
		for i := range t.Members {
			peerDTO := mapPeer(&t.Members[i])
			if peerDTO != nil {
				res.Members = append(res.Members, *peerDTO)
			}
		}
	}

	return res
}
