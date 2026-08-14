package model

// Typing is the client-shaped payload of an ephemeral "…is typing" indicator.
// From is the enriched typing participant — the same Peer shape as a message
// sender (resolved via the contact enricher). PreviewText is populated only for
// sessions authorized to see the sender's unsent draft; empty for everyone else.
type Typing struct {
	ThreadID    string `json:"thread_id"`
	TimeoutMs   int32  `json:"timeout_ms"`
	From        Peer   `json:"from"`
	PreviewText string `json:"preview_text,omitempty"`
}
