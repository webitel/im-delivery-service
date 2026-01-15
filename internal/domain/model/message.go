package model

import (
	"time"

	"github.com/google/uuid"
)

//go:generate stringer -type=PeerType
type PeerType int16

const (
	PeerUser PeerType = iota + 1 // Contact
	PeerBot                      // Schems
	PeerChat
	PeerChannel
)

// Message is the SINGLE source of truth for the entire system.
type Message struct {
	ID        uuid.UUID
	ThreadID  uuid.UUID
	From      Peer
	To        Peer
	Text      string
	CreatedAt int64
	// --- NEW FIELDS TO SUPPORT V2+ ---
	UpdatedAt int64          // Time of last edit (0 if never edited)
	IsForward bool           // Flag for forwarded messages
	Metadata  map[string]any // Flexible container for extra data (tags, localized strings, etc.)
	// ---------------------------------
	Documents []*Document
	Images    []*Image
}

// NewMessage remains the "standard" constructor (mostly for V1 and simple cases).
func NewMessage(id, threadID, fromID, text, occurredAt string, recipientID uuid.UUID, images []*Image, docs []*Document) *Message {
	return &Message{
		ID:        safeParseUUID(id),
		ThreadID:  safeParseUUID(threadID),
		Text:      text,
		CreatedAt: safeParseRFC3339(occurredAt),
		From:      Peer{ID: safeParseUUID(fromID), Type: PeerUser},
		To:        Peer{ID: recipientID, Type: PeerUser},
		Images:    images,
		Documents: docs,
		// New fields get safe default values
		Metadata: make(map[string]any),
	}
}

type Peer struct {
	ID   uuid.UUID
	Type PeerType
}

type Document struct {
	ID       string
	URL      string
	FileName string
	MimeType string
	Size     int64
}

type Image struct {
	ID         string
	URL        string
	FileName   string
	MimeType   string
	Thumbnails []string
}

func safeParseUUID(s string) uuid.UUID {
	val, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil
	}
	return val
}

func safeParseRFC3339(s string) int64 {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Now().UnixMilli()
	}
	return t.UnixMilli()
}
