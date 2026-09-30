package wsmarshaller

import "github.com/webitel/im-delivery-service/internal/domain/model"

// WSMessageStatus is a delivered/read/failed change; member has the same shape as a message sender.
type WSMessageStatus struct {
	ThreadID      string         `json:"thread_id"`
	Status        string         `json:"status"`
	Member        *WSPeer        `json:"member"`
	MessageIDs    []string       `json:"message_ids"`
	UpToSeq       int64          `json:"up_to_seq,omitempty"`
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
	ids := make([]string, 0, len(m.MessageIDs))
	for _, id := range m.MessageIDs {
		ids = append(ids, id.String())
	}

	return &WSMessageStatus{
		ThreadID:      m.ThreadID.String(),
		Status:        m.Status,
		Member:        mapPeer(m.Member),
		MessageIDs:    ids,
		UpToSeq:       m.UpToSeq,
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
