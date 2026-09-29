package model

import "github.com/google/uuid"

type MemberEvent struct {
	ThreadID  uuid.UUID      `json:"thread_id"`
	ContactID uuid.UUID      `json:"contact_id"`
	Metadata  map[string]any `json:"metadata,omitempty"`
	// UpdatesCursor is the GetUpdates position just before this change (set by im-thread-service).
	UpdatesCursor string `json:"updates_cursor,omitempty"`
	// Action: "joined" or "left".
	Action string `json:"action,omitempty"`
}
