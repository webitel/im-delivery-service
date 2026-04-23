package model

import (
	"fmt"

	"github.com/google/uuid"
)

type Message struct {
	ID        uuid.UUID      `json:"id"`
	SendID    string         `json:"send_id"`
	ThreadID  uuid.UUID      `json:"thread_id"`
	DomainID  int64          `json:"domain_id"`
	From      Peer           `json:"from"`
	To        []Peer         `json:"to"`
	Text      string         `json:"text"`
	CreatedAt int64          `json:"created_at"`
	EditedAt  int64          `json:"updated_at,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
	Documents []*Document    `json:"documents,omitempty"`
	Images    []*Image       `json:"images,omitempty"`
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
