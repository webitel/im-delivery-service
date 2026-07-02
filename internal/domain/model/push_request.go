package model

import (
	"fmt"
	"unicode/utf8"

	"github.com/google/uuid"
)

const (
	// MaxBodyLength defines the character limit before truncation.
	// 128 is a safe spot for most mobile notification trays.
	MaxBodyLength = 128
)

// pushSource is a localized interface to prevent circular imports with 'event' package.
type pushSource interface {
	GetID() string
	GetUserID() uuid.UUID
	GetKindName() string
	GetMetadata() map[string]string
}

// PushRequest represents the final transport entity for FCM/APNS gateways.
type PushRequest struct {
	UserID     string
	Devices    []Device
	Title      string
	Body       string
	Data       map[string]string
	IsSilent   bool
	CollapseID string
}

// FillFromEvent populates the PushRequest fields using data from the domain event.
func (r *PushRequest) FillFromEvent(ev pushSource) {
	// Map the recipient's UUID to a string format.
	if uid := ev.GetUserID(); uid != uuid.Nil {
		r.UserID = uid.String()
	}

	// Use event ID as a collapse key to group notifications from the same source.
	r.CollapseID = ev.GetID()

	// Resolve human-readable text based on event type and metadata.
	r.Title, r.Body = formatNotification(ev.GetKindName(), ev.GetMetadata())
}

// formatNotification handles the visual representation of the push.
func formatNotification(kind string, meta map[string]string) (title, body string) {
	sender := meta["sender_name"]
	if sender == "" {
		sender = "Someone"
	}

	switch kind {
	case "MessageCreated":
		// [FORMAT] Title: "New Message from Ihor Ihor", Body: "hello there !"
		title = fmt.Sprintf("New Message from %s", sender)

		rawText := meta["text"]
		if rawText == "" {
			rawText = "sent a file" // Fallback for attachment-only messages.
		}

		// Body now contains only the message content, truncated for safety.
		body = truncate(rawText, MaxBodyLength)

		return title, body

	case "ThreadCreated":
		// [FORMAT] Title: "New Conversation", Body: "Ihor Ihor started a new chat"
		title = "New Conversation"
		text := fmt.Sprintf("%s started a new chat", sender)
		body = truncate(text, MaxBodyLength)

		return title, body

	default:
		// Generic fallback for system or unknown events.
		return "New Update", "You have a new notification"
	}
}

// truncate safely trims a string based on UTF-8 runes instead of raw bytes.
// This prevents breaking multi-byte characters (Cyrillic, Emojis) mid-sequence.
func truncate(s string, max int) string {
	if utf8.RuneCountInString(s) <= max {
		return s
	}

	// Slice by runes to maintain character integrity.
	runes := []rune(s)

	return string(runes[:max]) + "..."
}
