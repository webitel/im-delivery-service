package model

import "github.com/google/uuid"

type Recipient struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type Thread struct {
	ID        uuid.UUID `json:"id"`
	DomainID  int32     `json:"domain_id"`
	CreatedAt int64     `json:"created_id"`
	Subject   string    `json:"subject"`
	Recipient Recipient `json:"recipient"`
	Type      string    `json:"type"`
	Members   []Peer    `json:"members,omitempty"`
}
