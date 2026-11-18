package model

import "fmt"

type InteractiveCallback struct {
	ReactedBy    Peer   `json:"reacted_by"`
	InReplyTo    string `json:"in_reply_to"`
	ButtonCode   string `json:"button_code"`
	CallbackData string `json:"callback_data"`
	ReactedAt    string `json:"reacted_at"`
	DomainID     int    `json:"domain_id"`
	ThreadID     string `json:"thread_id"`
	Receiver     Peer   `json:"receiver"`

	// UpdatesCursor is the recipient's GetUpdates cursor after this callback
	// (set by im-thread-service once the callback is journaled).
	UpdatesCursor string `json:"updates_cursor,omitempty"`
}

func (c *InteractiveCallback) RoutingKey() string {
	return fmt.Sprintf("im_delivery.v1.%d.interactive_callback.reacted", c.ReactedBy.DomainID)
}
