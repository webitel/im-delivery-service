package model

import "fmt"

type InteractiveCallback struct {
	ReactedBy    Peer   `json:"reacted_by"`
	InReplyTo    string `json:"in_reply_to"`
	ButtonCode   string `json:"button_code"`
	CallbackData string `json:"callback_data"`
	ReactedAt    string `json:"reacted_at"`
}

func (c *InteractiveCallback) RoutingKey() string {
	return fmt.Sprintf("im_delivery.v1.%d.interactive_callback.reacted", c.ReactedBy.DomainID)
}
