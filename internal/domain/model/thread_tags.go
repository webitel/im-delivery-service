package model

import "github.com/google/uuid"

// ThreadTagsChanged is the recipient's own current tags on a thread.
type ThreadTagsChanged struct {
	ThreadID uuid.UUID `json:"thread_id"`
	Tags     []string  `json:"tags"`
	// UpdatesCursor is the recipient's GetUpdates cursor after this change (set by im-thread-service).
	UpdatesCursor string `json:"updates_cursor,omitempty"`
}
