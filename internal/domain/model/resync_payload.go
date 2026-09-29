package model

// ResyncPayload tells the client events for its connection were dropped; it catches up via
// GetUpdates from Cursor, or from its saved cursor when Cursor is empty.
type ResyncPayload struct {
	Cursor string `json:"cursor,omitempty"`
}

// UpdatesCursorOf is the GetUpdates position a live payload carries; empty when it has none.
func UpdatesCursorOf(payload any) string {
	switch p := payload.(type) {
	case *Message:
		return p.UpdatesCursor
	case *MessageEdited:
		return p.UpdatesCursor
	case *MessageDeleted:
		return p.UpdatesCursor
	case *MessageReaction:
		return p.UpdatesCursor
	case *MemberEvent:
		return p.UpdatesCursor
	}

	return ""
}
