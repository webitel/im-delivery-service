package wsmarshaller

import (
	"strings"

	"github.com/webitel/im-delivery-service/internal/domain/model"
)

// WSPeer represents a participant in the chat.
type WSPeer struct {
	ID     string `json:"id"`
	Type   string `json:"type"` // [ENUM] "user", "group", "channel"
	Name   string `json:"name,omitempty"`
	Issuer string `json:"issuer,omitempty"`
}

// mapPeer transforms domain Peer model to a flat JSON structure.
func mapPeer(p model.Peer) *WSPeer {
	// [STRINGER] Get the string representation (e.g., "PeerUser")
	typeName := p.Type.String()

	// [CLEANUP] Remove "Peer" prefix and convert to lowercase
	cleanType := strings.ToLower(strings.TrimPrefix(typeName, "Peer"))

	res := &WSPeer{
		ID:   p.Sub,
		Type: cleanType,
	}

	// [IDENTITY_ENRICHMENT]
	if p.IsEnriched() {
		res.Name = p.Name
		res.Issuer = p.Issuer
	}

	return res
}

// WSMessage with full Peer support.
type WSMessage struct {
	ID        string         `json:"id"`
	ThreadID  string         `json:"thread_id"`
	Text      string         `json:"text"`
	CreatedAt int64          `json:"created_at"`
	UpdatedAt int64          `json:"updated_at,omitempty"`
	From      *WSPeer        `json:"from"` //  Full peer object
	To        *WSPeer        `json:"to"`   //  Full peer object
	Type      string         `json:"type"`
	Media     any            `json:"media,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

func mapMessage(m *model.Message) *WSMessage {
	msg := &WSMessage{
		ID:        m.ID.String(),
		ThreadID:  m.ThreadID.String(),
		Text:      m.Text,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.EditedAt,
		From:      mapPeer(m.From), // [MAPPING] Use the new peer mapper
		To:        mapPeer(m.To),   // [MAPPING] Use the new peer mapper
		Metadata:  m.Metadata,
		Type:      "text",
	}

	// [MEDIA_SELECTION]
	switch {
	case len(m.Images) > 0:
		msg.Type = "image"
		msg.Media = m.Images[0]
	case len(m.Documents) > 0:
		msg.Type = "document"
		msg.Media = m.Documents[0]
	}

	return msg
}
