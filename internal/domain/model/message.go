package model

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

type Message struct {
	ID          uuid.UUID       `json:"id"`
	SendID      string          `json:"send_id"`
	ThreadID    uuid.UUID       `json:"thread_id"`
	DomainID    int64           `json:"domain_id"`
	From        Peer            `json:"from"`
	To          []Peer          `json:"to"`
	Text        string          `json:"text"`
	Type        string          `json:"type"`
	CreatedAt   int64           `json:"created_at"`
	EditedAt    int64           `json:"updated_at,omitempty"`
	Metadata    map[string]any  `json:"metadata,omitempty"`
	Documents   []*Document     `json:"documents,omitempty"`
	Images      []*Image        `json:"images,omitempty"`
	Interactive json.RawMessage `json:"interactive,omitempty"`
	Location    *Location       `json:"location,omitempty"`
	Contact     *Contact        `json:"contact,omitempty"`
	System      *System         `json:"system,omitempty"`
	ReplyTo     *ReplyTo        `json:"reply_to,omitempty"`

	ForwardOrigin *ForwardOrigin `json:"forward_origin,omitempty"`
}

type ForwardOrigin struct {
	Kind            int16      `json:"kind"`
	SenderID        *uuid.UUID `json:"sender_id,omitempty"`
	SenderName      string     `json:"sender_name,omitempty"`
	OriginalSentAt  int64      `json:"original_sent_at,omitempty"`
	SourceMessageID *uuid.UUID `json:"source_message_id,omitempty"`
}

type ReplyTo struct {
	MessageID      uuid.UUID `json:"message_id"`
	SenderID       uuid.UUID `json:"sender_id"`
	Type           string    `json:"type"`
	Body           string    `json:"body"`
	CreatedAt      int64     `json:"created_at"`
	AttachmentKind *string   `json:"attachment_kind,omitempty"`
	AttachmentName *string   `json:"attachment_name,omitempty"`
	AttachmentMime *string   `json:"attachment_mime,omitempty"`

	AttachmentAddress *string `json:"attachment_address,omitempty"`
}

func (m *Message) RoutingKey() string {
	return fmt.Sprintf("im_delivery.v1.%d.message.created", m.DomainID)
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
	if m.Text != "" {
		return m.Text
	}

	if len(m.Images) > 0 {
		return "Photo"
	}

	if len(m.Documents) > 0 {
		return m.Documents[0].Name
	}

	return "Sent a message"
}

type Document struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Mime string `json:"mime"`
	Size int64  `json:"size"`
	URL  string `json:"url"`
}

type Image struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Mime string `json:"mime"`
	URL  string `json:"url"`
}

type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Address   *string `json:"address,omitempty"`
	Name      *string `json:"name,omitempty"`
}

type Contact struct {
	Name  *string `json:"name,omitempty"`
	Phone *string `json:"phone,omitempty"`
	Email *string `json:"email,omitempty"`
}

type System struct {
	Type     string         `json:"type"`
	Metadata map[string]any `json:"metadata,omitempty"`
}
