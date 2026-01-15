package model

import "github.com/google/uuid"

//go:generate stringer -type=PeerType
type PeerType int16

const (
	// [ZERO_VALUE_GUARD] WE START FROM 1 TO DISTINGUISH FROM UNINITIALIZED DATA
	PeerUser PeerType = iota + 1
	PeerBot
	PeerChat
	PeerChannel
)

type Peer struct {
	ID   uuid.UUID
	Type PeerType
}

// [MESSAGE] CORE ENTITY REPRESENTING A CONVERSATION ELEMENT
type Message struct {
	ID        uuid.UUID
	ThreadID  uuid.UUID
	From      Peer
	To        Peer
	Text      string
	CreatedAt int64
	UpdatedAt int64
	Documents  []*Document
	Images     []*Image
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
