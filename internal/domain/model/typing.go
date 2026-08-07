package model

// Typing is the client-shaped payload of an ephemeral "…is typing" indicator.
// PreviewText is populated only for sessions authorized to see the sender's
// unsent draft; it is empty for everyone else.
type Typing struct {
	ThreadID    string `json:"thread_id"`
	MemberID    string `json:"member_id"`
	TimeoutMs   int32  `json:"timeout_ms"`
	PreviewText string `json:"preview_text,omitempty"`
}
