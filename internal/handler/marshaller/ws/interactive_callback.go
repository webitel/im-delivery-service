package wsmarshaller

import (
	"github.com/webitel/im-delivery-service/internal/domain/model"
	"github.com/webitel/im-delivery-service/internal/domain/util"
)

// WSInteractiveCallback is the transport shape of a pressed interactive button.
// ReactedBy and Receiver are marshaled with the SAME helper as a message sender
// (mapPeer), so clients reuse the participant model of every other event.
type WSInteractiveCallback struct {
	ThreadID     string  `json:"thread_id"`
	MessageID    string  `json:"message_id"`
	ButtonCode   string  `json:"button_code"`
	CallbackData string  `json:"callback_data"`
	ReactedAt    int64   `json:"reacted_at"`
	ReactedBy    *WSPeer `json:"reacted_by"`
	Receiver     *WSPeer `json:"receiver"`

	UpdatesCursor string `json:"updates_cursor,omitempty"`
}

func mapInteractiveCallback(c *model.InteractiveCallback) *WSInteractiveCallback {
	return &WSInteractiveCallback{
		ThreadID:     c.ThreadID,
		MessageID:    c.InReplyTo,
		ButtonCode:   c.ButtonCode,
		CallbackData: c.CallbackData,
		ReactedAt:    util.SafeParseRFC3339(c.ReactedAt),
		ReactedBy:    mapPeer(&c.ReactedBy),
		Receiver:     mapPeer(&c.Receiver),

		UpdatesCursor: c.UpdatesCursor,
	}
}
