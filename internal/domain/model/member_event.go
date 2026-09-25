package model

import "github.com/google/uuid"

type MemberEvent struct {
	ThreadID  uuid.UUID      `json:"thread_id"`
	ContactID uuid.UUID      `json:"contact_id"`
	Metadata  map[string]any `json:"metadata,omitempty"`
	// Per-thread update_seq (GetThreadUpdates cursor); clients use it for gap detection.
	UpdateSeq int64 `json:"update_seq,omitempty"`
	// Action: "joined" or "left".
	Action string `json:"action,omitempty"`
}
