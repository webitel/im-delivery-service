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

// WSMessage represents the message structure sent over WebSockets.
type WSMessage struct {
	ID        string            `json:"id"`
	SendID    string            `json:"send_id"`
	ThreadID  string            `json:"thread_id"`
	Sender    *WSPeer           `json:"sender"`
	To        *WSPeer           `json:"to"`
	CreatedAt int64             `json:"created_at"`
	EditedAt  int64             `json:"edited_at,omitempty"`
	Body      string            `json:"body"`
	Type      string            `json:"type"`
	Images    []*model.Image    `json:"images,omitempty"`
	Documents []*model.Document `json:"documents,omitempty"`
}

// mapPeer normalizes domain Peer into a transport-friendly DTO.
func mapPeer(p *model.Peer) *WSPeer {
	if p == nil {
		return nil
	}
	return &WSPeer{
		Sub:   p.Sub,
		Iss:   p.Issuer,
		Name:  p.Name,
		Type:  strings.ToLower(strings.TrimPrefix(p.Type.String(), "Peer")),
		IsBot: p.IsBot,
	}
}

// mapMessage converts domain model.Message to WSMessage DTO.
func mapMessage(m *model.Message) *WSMessage {
	msg := &WSMessage{
		ID:        m.ID.String(),
		SendID:    m.SendID,
		ThreadID:  m.ThreadID.String(),
		Sender:    mapPeer(&m.From),
		To:        mapPeer(m.To),
		CreatedAt: m.CreatedAt,
		EditedAt:  m.EditedAt,
		Body:      m.Text,
		Type:      "text",
	}

	// [MEDIA_ASSIGNMENT] Directly assign slices to explicit fields.
	if len(m.Images) > 0 {
		msg.Type = "image"
		msg.Images = m.Images
	}

	if len(m.Documents) > 0 {
		// If there are both images and documents, we prioritize 'document' type
		// or keep it as 'image' based on your business logic preferences.
		if msg.Type == "text" {
			msg.Type = "document"
		}
		msg.Documents = m.Documents
	}

	return msg
}
