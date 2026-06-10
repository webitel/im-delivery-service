package payload

import (
	"github.com/google/uuid"
	"github.com/webitel/im-delivery-service/internal/domain/model"
	"github.com/webitel/webitel-go-kit/pkg/errors"
)

type ReactedBy struct {
	ID   string `json:"id"`
	Type int    `json:"type"`
}

type InteractiveCallbackV1 struct {
	ReactedBy    ReactedBy `json:"reacted_by"`
	InReplyTo    string    `json:"in_reply_to"`
	ButtonCode   string    `json:"button_code"`
	CallbackData string    `json:"callback_data"`
	ReactedAt    string    `json:"reacted_at"`
	DomainID     int       `json:"domain_id"`
	ThreadID     string    `json:"thread_id"`
}

func (c *InteractiveCallbackV1) ToDomain() (*model.InteractiveCallback, error) {
	if c == nil {
		return nil, errors.InvalidArgument(
			"received nil pointer interactive callback payload",
			errors.WithID("payload.interactive_callback.to_domain.nil_pointer_receiver"),
		)
	}

	id, err := uuid.Parse(c.ReactedBy.ID)
	if err != nil {
		return nil, errors.InvalidArgument(
			"parsing reacted by ID",
			errors.WithCause(err),
			errors.WithID("payload.interactive_callback.to_domain.parsing_reacted_by_id"),
		)
	}

	return &model.InteractiveCallback{
		ReactedBy:    model.Peer{ID: id, Type: model.PeerUser},
		InReplyTo:    c.InReplyTo,
		ButtonCode:   c.ButtonCode,
		CallbackData: c.CallbackData,
		ReactedAt:    c.ReactedAt,
		DomainID:     c.DomainID,
		ThreadID:     c.ThreadID,
	}, nil
}
