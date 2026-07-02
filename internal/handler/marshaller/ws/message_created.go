package wsmarshaller

import (
	"encoding/json"

	"github.com/webitel/im-delivery-service/internal/domain/model"
)

// WSContact holds the identity provider information.
type WSContact struct {
	Sub      string `json:"sub"`
	Iss      string `json:"iss,omitempty"`
	Name     string `json:"name,omitempty"`
	Username string `json:"username,omitempty"`
	Type     string `json:"type"`
	IsBot    bool   `json:"is_bot"`
}

// WSPeer represents a participant with nested contact info and role.
type WSPeer struct {
	ID      string     `json:"id"`
	Contact *WSContact `json:"contact"`
	Role    string     `json:"role"`
}

type WSMessage struct {
	ID          string            `json:"id"`
	SendID      string            `json:"send_id"`
	ThreadID    string            `json:"thread_id"`
	Sender      *WSPeer           `json:"sender"`
	To          []*WSPeer         `json:"to,omitempty"`
	CreatedAt   int64             `json:"created_at"`
	EditedAt    int64             `json:"edited_at,omitempty"`
	Body        string            `json:"body"`
	Type        string            `json:"type"`
	Images      []*model.Image    `json:"images,omitempty"`
	Documents   []*model.Document `json:"documents,omitempty"`
	Interactive json.RawMessage   `json:"interactive,omitempty"`
	Contact     *model.Contact    `json:"contact,omitempty"`
	Location    *model.Location   `json:"location,omitempty"`
	System      *model.System     `json:"system,omitempty"`
	Metadata    map[string]any    `json:"metadata,omitempty"`
}

// mapPeer converts internal model.Peer to the nested WSPeer structure.
func mapPeer(p *model.Peer) *WSPeer {
	if p == nil {
		return nil
	}

	return &WSPeer{
		ID:   p.MemberID, // mapping member_id to id
		Role: model.ParseRole(p.Role).String(),
		Contact: &WSContact{
			Sub:      p.Sub,
			Iss:      p.Issuer,
			Name:     p.Name,
			Username: p.Name, // Using name as username as fallback
			Type:     p.ContactType,
			IsBot:    p.IsBot, // Assigned inside the contact object
		},
	}
}

// mapMessage transforms the internal message domain into a WebSocket DTO.
func mapMessage(m *model.Message) *WSMessage {
	msg := &WSMessage{
		ID:        m.ID.String(),
		SendID:    m.SendID,
		ThreadID:  m.ThreadID.String(),
		Sender:    mapPeer(&m.From),
		To:        nil, // Kept empty as per current requirement
		CreatedAt: m.CreatedAt,
		EditedAt:  m.EditedAt,
		Body:      m.Text,
		Type:      m.Type,
		Contact:   m.Contact,
		Location:  m.Location,
		Metadata:  m.Metadata,
		System:    m.System,
	}

	if len(m.Images) > 0 {
		msg.Images = m.Images
	}

	if len(m.Documents) > 0 {
		msg.Documents = m.Documents
	}

	if m.Interactive != nil {
		msg.Interactive = m.Interactive
	}

	return msg
}
