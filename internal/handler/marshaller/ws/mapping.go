package wsmarshaller

import (
	"strings"

	"github.com/webitel/im-delivery-service/internal/domain/model"
)

type WSPeer struct {
	ID     string `json:"id"`
	Type   string `json:"type"`
	Name   string `json:"name,omitempty"`
	Issuer string `json:"issuer,omitempty"`
}

type WSMessage struct {
	ID        string         `json:"id"`
	ThreadID  string         `json:"thread_id"`
	From      *WSPeer        `json:"from"`
	To        *WSPeer        `json:"to"`
	CreatedAt int64          `json:"created_at"`
	EditedAt  int64          `json:"edited_at,omitempty"`
	Text      string         `json:"text"`
	Type      string         `json:"type"`
	Content   map[string]any `json:"content,omitempty"`
}

func mapPeer(p model.Peer) *WSPeer {
	return &WSPeer{
		ID:     p.Sub,
		Type:   strings.ToLower(strings.TrimPrefix(p.Type.String(), "Peer")),
		Name:   p.Name,
		Issuer: p.Issuer,
	}
}

func mapMessage(m *model.Message) *WSMessage {
	msg := &WSMessage{
		ID:        m.ID.String(),
		ThreadID:  m.ThreadID.String(),
		From:      mapPeer(m.From),
		To:        mapPeer(m.To),
		CreatedAt: m.CreatedAt,
		EditedAt:  m.EditedAt,
		Text:      m.Text,
		Type:      "text",
	}

	// [MEDIA_MAPPING] Handle image/document attachments
	switch {
	case len(m.Images) > 0:
		msg.Type = "image"
		msg.Content = map[string]any{"image": m.Images[0]}
	case len(m.Documents) > 0:
		msg.Type = "document"
		msg.Content = map[string]any{"document": m.Documents[0]}
	}
	return msg
}
