package model

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type Message struct {
	ID        uuid.UUID      `json:"id"`
	SendID    string         `json:"send_id"`
	ThreadID  uuid.UUID      `json:"thread_id"`
	DomainID  int64          `json:"domain_id"`
	From      Peer           `json:"from"`
	To        *Peer          `json:"to"`
	Text      string         `json:"text"`
	CreatedAt int64          `json:"created_at"`
	EditedAt  int64          `json:"updated_at,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
	Documents []*Document    `json:"documents,omitempty"`
	Images    []*Image       `json:"images,omitempty"`
}

// RoutingKey generates an AMQP routing key based on the message destination.
// If the recipient (To) is missing, it returns a truncated key without trailing dots.
func (m *Message) RoutingKey() string {
	// Base prefix that is always present
	// im_delivery.v1.{domain_id}
	key := fmt.Sprintf("im_delivery.v1.%d", m.DomainID)

	// If there is no recipient, we stop here to avoid empty segments or "system" labels
	if m.To == nil {
		// Results in: im_delivery.v1.1.message.created
		return key + ".message.created"
	}

	// Determine peer type
	peerType := "contact"
	issuer := strings.ToLower(m.To.Issuer)
	if strings.Contains(issuer, "bot") || strings.Contains(issuer, "schema") {
		peerType = "bot"
	}

	// If recipient exists, we provide the full granular path
	// Results in: im_delivery.v1.1.contact.3.message.created
	return fmt.Sprintf("%s.%s.%s.message.created", key, peerType, m.To.Sub)
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
