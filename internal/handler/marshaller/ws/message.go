package wsmarshaller

import (
	"github.com/webitel/im-delivery-service/internal/domain/model"
)

type WSMessage struct {
	ID        string         `json:"id"`
	ThreadID  string         `json:"thread_id"`
	Text      string         `json:"text"`
	CreatedAt int64          `json:"created_at"`
	UpdatedAt int64          `json:"updated_at,omitempty"`
	From      string         `json:"from_id"`
	Type      string         `json:"type"` // "text", "image", "document"
	Media     any            `json:"media,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

func mapMessage(m *model.Message) *WSMessage {
	msg := &WSMessage{
		ID:        m.ID.String(),
		ThreadID:  m.ThreadID.String(),
		Text:      m.Text,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
		From:      m.From.ID.String(),
		Metadata:  m.Metadata,
		Type:      "text",
	}

	// Handle Media (Simplified for JSON)
	if len(m.Images) > 0 {
		msg.Type = "image"
		msg.Media = m.Images[0]
	} else if len(m.Documents) > 0 {
		msg.Type = "document"
		msg.Media = m.Documents[0]
	}

	return msg
}
