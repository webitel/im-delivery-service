package wsmarshaller

import "github.com/webitel/im-delivery-service/internal/domain/model"

// WSMessageDeleted is the transport shape of a deletion: clients match the
// message by id and remove it, so it carries no body or attachments. DeletedBy
// is marshaled with the SAME helper as a message sender (mapPeer) — identical
// shape to a NewMessageEvent's `sender` and an edit's `edited_by`.
type WSMessageDeleted struct {
	ID        string  `json:"id"`
	ThreadID  string  `json:"thread_id"`
	DeletedBy *WSPeer `json:"deleted_by"`
	CreatedAt int64   `json:"created_at"`
	DeletedAt int64   `json:"deleted_at"`
}

func mapMessageDeleted(m *model.MessageDeleted) *WSMessageDeleted {
	return &WSMessageDeleted{
		ID:        m.ID.String(),
		ThreadID:  m.ThreadID.String(),
		DeletedBy: mapPeer(&m.DeletedBy),
		CreatedAt: m.CreatedAt,
		DeletedAt: m.DeletedAt,
	}
}
