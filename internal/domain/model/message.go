package model

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// [MESSAGE] Core chat message entity.
type Message struct {
	ID        uuid.UUID
	SendID    string
	ThreadID  uuid.UUID
	DomainID  int64
	From      Peer
	To        *Peer
	Text      string
	CreatedAt int64
	EditedAt  int64
	Metadata  map[string]any
	Documents []*Document
	Images    []*Image
}

// [ROUTING_KEY] Defines AMQP routing key.
func (m *Message) RoutingKey() string {
	peerType := "contact"

	if m.To != nil {
		issuer := strings.ToLower(m.To.Issuer)

		if strings.Contains(issuer, "bot") ||
			strings.Contains(issuer, "schema") {

			peerType = "bot"
		}
	}

	return fmt.Sprintf(
		"im_delivery.v1.%d.%s.%s.message.created",
		m.DomainID,
		peerType,
		m.To.Sub,
	)
}

// [PUSH_TITLE]
func (m *Message) NotificationTitle() string {
	if m.From.Name != "" {
		return m.From.Name
	}

	return "New Message"
}

// [PUSH_BODY]
func (m *Message) NotificationBody() string {
	if m.Text != "" {
		return m.Text
	}

	if len(m.Images) > 0 {
		return "Photo"
	}

	if len(m.Documents) > 0 {
		return m.Documents[0].FileName
	}

	return "Sent a message"
}

type Document struct {
	ID       string `json:"id"`
	FileName string `json:"file_name"`
	MimeType string `json:"mime_type"`
	Size     int64  `json:"size"`
}

type Image struct {
	ID       string `json:"id"`
	FileName string `json:"file_name"`
	MimeType string `json:"mime_type"`
	URL      string `json:"url"`
}
