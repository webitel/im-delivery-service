package wsmarshaller

import "github.com/webitel/im-delivery-service/internal/domain/model"

// WSThread represents the thread structure for WebSocket transport.
type WSThread struct {
	ID        string   `json:"id"`
	DomainID  int32    `json:"domain_id"`
	CreatedAt int64    `json:"created_at"`
	Subject   string   `json:"subject"`
	Type      string   `json:"type"`
	Members   []WSPeer `json:"members,omitempty"` // [FIX] Now using unified WSPeer
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

	// [ITERATION] Map enriched members using the shared mapPeer logic.
	if len(t.Members) > 0 {
		res.Members = make([]WSPeer, len(t.Members))
		for i, p := range t.Members {
			res.Members[i] = *mapPeer(&p)
		}
	}

	return res
}
