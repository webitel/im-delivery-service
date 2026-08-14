package payload

// TypingV1 is the ephemeral typing indicator published by im-thread on
// im_message.<thread_id>.typing.v1 (fire-and-forget, bypasses the outbox).
type TypingV1 struct {
	ThreadID   string `json:"thread_id"`
	MemberID   string `json:"member_id"` // who is typing (contact id)
	Role       int32  `json:"role"`      // sender's thread role; overlaid onto the enriched peer
	DomainID   int32  `json:"domain_id"` // used to enrich the sender via the contact resolver
	TimeoutMS  int32  `json:"timeout_ms"`
	OccurredAt string `json:"occurred_at"`

	// To are the recipient contact ids for the indicator. When absent the
	// preview allow-list is used as the recipient set (forward-compatible with
	// im-thread emitting an explicit recipient list later).
	To []string `json:"to,omitempty"`

	// PreviewText is the sender's unsent draft (Live Typing Preview). It is
	// attached to a delivered session only if that session's member is in
	// PreviewVisibleTo.
	PreviewText string `json:"preview_text,omitempty"`

	// PreviewVisibleTo are the contact ids allowed to receive PreviewText.
	PreviewVisibleTo []string `json:"preview_visible_to,omitempty"`
}

// Recipients returns the contact ids that should receive the indicator: the
// explicit To list when present, otherwise the preview allow-list.
func (t *TypingV1) Recipients() []string {
	if len(t.To) > 0 {
		return t.To
	}

	return t.PreviewVisibleTo
}
