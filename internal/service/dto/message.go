package dto

// [RABBIT_V1] THE CURRENT PAYLOAD STRUCTURE FROM IM-THREAD-SERVICE
type MessageV1 struct {
	MessageID  string        `json:"message_id"`
	ThreadID   string        `json:"thread_id"`
	FromID     string        `json:"from_id"`
	FromType   int           `json:"from_type"`
	ToID       string        `json:"to_id"`
	ToType     int           `json:"to_type"`
	Body       string        `json:"body"`
	OccurredAt string        `json:"occurred_at"`
	Images     []ImageDTO    `json:"images"`
	Documents  []DocumentDTO `json:"documents"`
}

type MessageV2 struct {
	MessageID  string            `json:"message_id"`
	ThreadID   string            `json:"thread_id"`
	FromID     string            `json:"from_id"`
	Body       string            `json:"body"`
	OccurredAt int64             `json:"occurred_at"` // Changed to Unix timestamp
	EditedAt   int64             `json:"edited_at"`   // New: Track edits
	IsForward  bool              `json:"is_forward"`  // New: Forwarding flag
	Metadata   map[string]string `json:"metadata"`    // New: Flexible key-value pairs
	Images     []ImageDTO        `json:"images"`
	Documents  []DocumentDTO     `json:"documents"`
}

type ImageDTO struct {
	FileID int64  `json:"file_id"`
	Mime   string `json:"mime"`
	Name   string `json:"name"`
}

type DocumentDTO struct {
	FileID int64  `json:"file_id"`
	Mime   string `json:"mime"`
	Name   string `json:"name"`
	Size   int64  `json:"size"`
}
