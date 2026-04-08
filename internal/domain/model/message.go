package model

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// Message is the central entity for chat communications.
type Message struct {
	ID       uuid.UUID `json:"id"`
	SendID   string    `json:"send_id"`
	ThreadID uuid.UUID `json:"thread_id"`
	DomainID int64     `json:"domain_id"`
	From     Peer      `json:"from"`
	// To is an optional pointer to the recipient peer.
	// Can be nil if the message is broadcast or system-oriented.
	To        *Peer          `json:"to"`
	Text      string         `json:"text"`
	CreatedAt int64          `json:"created_at"`
	EditedAt  int64          `json:"updated_at,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
	Documents []*Document    `json:"documents,omitempty"`
	Images    []*Image       `json:"images,omitempty"`
}

// RoutingKey generates an AMQP routing key based on the message destination.
// It handles nil recipients safely to prevent runtime panics.
func (m *Message) RoutingKey() string {
	peerType := "contact"
	// Default 'sub' identifier for system or unassigned routes
	destinationSub := "system"

	// Safety check: only access fields if the pointer is non-nil
	if m.To != nil {
		destinationSub = m.To.Sub

		issuer := strings.ToLower(m.To.Issuer)
		// Categorize as 'bot' if recipient is a bot or a workflow schema
		if strings.Contains(issuer, "bot") || strings.Contains(issuer, "schema") {
			peerType = "bot"
		}
	}

	// Format: im_delivery.v1.{domain_id}.{peer_type}.{sub}.message.created
	return fmt.Sprintf(
		"im_delivery.v1.%d.%s.%s.message.created",
		m.DomainID,
		peerType,
		destinationSub,
	)
}

// NotificationTitle returns a friendly display name for push notifications.
func (m *Message) NotificationTitle() string {
	if m.From.Name != "" {
		return m.From.Name
	}
	return "New Message"
}

// NotificationBody prepares a short preview of the message content.
func (m *Message) NotificationBody() string {
	// 1. Prefer plain text if available
	if m.Text != "" {
		return m.Text
	}

	// 2. Fallback to media indicators
	if len(m.Images) > 0 {
		return "Photo"
	}

	if len(m.Documents) > 0 {
		// Use the first document's filename as a hint
		return m.Documents[0].FileName
	}

	return "Sent a message"
}

// Document defines a generic file attachment metadata.
type Document struct {
	ID       string `json:"id"`
	FileName string `json:"file_name"`
	MimeType string `json:"mime_type"`
	Size     int64  `json:"size"`
}

// Image defines image-specific attachment metadata.
type Image struct {
	ID       string `json:"id"`
	FileName string `json:"file_name"`
	MimeType string `json:"mime_type"`
	URL      string `json:"url"`
}
