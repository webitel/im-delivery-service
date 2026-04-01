package wsmarshaller

import (
	"strings"

	"github.com/webitel/im-delivery-service/internal/domain/model"
)

// WSPeer represents a participant in the transport layer.
// [SYNC] Using 'sub' and 'iss' to align with external API standards.
type WSPeer struct {
	Sub   string `json:"sub"`
	Iss   string `json:"iss,omitempty"`
	Name  string `json:"name,omitempty"`
	Type  string `json:"type"`
	IsBot bool   `json:"is_bot"`
}

type WSMessage struct {
	ID        string         `json:"id"`
	SendID    string         `json:"send_id"`
	ThreadID  string         `json:"thread_id"`
	Sender    *WSPeer        `json:"sender"`
	To        *WSPeer        `json:"to"`
	CreatedAt int64          `json:"created_at"`
	EditedAt  int64          `json:"edited_at,omitempty"`
	Body      string         `json:"body"`
	Type      string         `json:"type"`
	Content   map[string]any `json:"content,omitempty"`
}

// mapPeer normalizes domain Peer into a transport-friendly DTO.
// [POINTER_LOGIC] Accepting a pointer to handle optional peers (like 'To').
func mapPeer(p *model.Peer) *WSPeer {
	if p == nil {
		return nil
	}
	return &WSPeer{
		Sub:  p.Sub,
		Iss:  p.Issuer,
		Name: p.Name,
		// [FORMAT] Consistent lowercase naming for types.
		Type:  strings.ToLower(strings.TrimPrefix(p.Type.String(), "Peer")),
		IsBot: p.IsBot,
	}
}

func mapMessage(m *model.Message) *WSMessage {
	msg := &WSMessage{
		ID:       m.ID.String(),
		SendID:   m.SendID,
		ThreadID: m.ThreadID.String(),
		// [FIX] Use address operator '&' because m.From is a struct, but mapPeer expects *model.Peer.
		Sender:    mapPeer(&m.From),
		To:        mapPeer(m.To),
		CreatedAt: m.CreatedAt,
		EditedAt:  m.EditedAt,
		Body:      m.Text,
		Type:      "text",
	}

	// [MEDIA_MAPPING] Handle image/document attachments.
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
