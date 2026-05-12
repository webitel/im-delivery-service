package model

import "github.com/google/uuid"

type MemberEvent struct {
	ThreadID  uuid.UUID      `json:"thread_id"`
	ContactID uuid.UUID      `json:"contact_id"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}
