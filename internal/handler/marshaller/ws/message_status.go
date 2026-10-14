package wsmarshaller

import "github.com/webitel/im-delivery-service/internal/domain/model"

// WSMessageStatus is a delivered/read/failed change; member has the same shape as a message sender.
// Delivered/read move the horizon to up_to_seq; failed names its message_id.
type WSMessageStatus struct {
	ThreadID      string         `json:"thread_id"`
	Status        string         `json:"status"`
	Member        *WSPeer        `json:"member"`
	MessageID     string         `json:"message_id,omitempty"`
	UpToSeq       int64          `json:"up_to_seq,omitempty"`
	UnreadCount   *int64         `json:"unread_count,omitempty"`
	Via           string         `json:"via,omitempty"`
	Error         *WSStatusError `json:"error,omitempty"`
	OccurredAt    int64          `json:"occurred_at"`
	UpdatesCursor string         `json:"updates_cursor,omitempty"`
}

// WSStatusError is why a message did not reach the member.
type WSStatusError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func mapMessageStatus(m *model.MessageStatusUpdate) *WSMessageStatus {
	return &WSMessageStatus{
		ThreadID:      m.ThreadID.String(),
		Status:        m.Status,
		Member:        mapPeer(m.Member),
		MessageID:     m.FailedMessageID(),
		UpToSeq:       m.UpToSeq,
		UnreadCount:   m.UnreadCount,
		Via:           m.Via,
		Error:         mapStatusError(m),
		OccurredAt:    m.OccurredAt,
		UpdatesCursor: m.UpdatesCursor,
	}
}

func mapStatusError(m *model.MessageStatusUpdate) *WSStatusError {
	code, message, ok := m.Failure()
	if !ok {
		return nil
	}

	return &WSStatusError{Code: code, Message: message}
}
