package wsmarshaller

import "github.com/webitel/im-delivery-service/internal/domain/model"

// WSMessageEdited is the transport shape of an edit: clients match the message
// by id and replace its content, so it carries no attachments or reply context.
type WSMessageEdited struct {
	ID        string         `json:"id"`
	ThreadID  string         `json:"thread_id"`
	EditedBy  *WSPeer        `json:"edited_by"`
	Body      string         `json:"body"`
	Type      string         `json:"type"`
	Version   int32          `json:"version,omitempty"`
	CreatedAt int64          `json:"created_at"`
	EditedAt  int64          `json:"edited_at"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

func mapMessageEdited(m *model.MessageEdited) *WSMessageEdited {
	return &WSMessageEdited{
		ID:        m.ID.String(),
		ThreadID:  m.ThreadID.String(),
		EditedBy:  mapPeer(&m.EditedBy),
		Body:      m.Text,
		Type:      m.Type,
		Version:   m.Version,
		CreatedAt: m.CreatedAt,
		EditedAt:  m.EditedAt,
		Metadata:  m.Metadata,
	}
}
