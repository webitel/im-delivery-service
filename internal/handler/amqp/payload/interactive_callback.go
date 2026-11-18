package payload

import (
	"github.com/google/uuid"

	"github.com/webitel/webitel-go-kit/pkg/errors"

	"github.com/webitel/im-delivery-service/internal/domain/model"
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
	Receiver     string    `json:"receiver"`

	// ReactedByMember and ReceiverMember carry the thread memberships (member id
	// and role) behind ReactedBy and Receiver. Older im-thread-service builds do
	// not send them; the event then has no member ids.
	ReactedByMember *Recipient `json:"reacted_by_member,omitempty"`
	ReceiverMember  *Recipient `json:"receiver_member,omitempty"`

	UpdatesCursor string `json:"updates_cursor,omitempty"`
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

	receiverID, err := uuid.Parse(c.Receiver)
	if err != nil {
		return nil, errors.InvalidArgument(
			"parsing receiver ID",
			errors.WithCause(err),
			errors.WithID("payload.interactive_callback.to_domain.parsing_receiver_id"),
		)
	}

	return &model.InteractiveCallback{
		ReactedBy:     withMembership(model.Peer{ID: id, Type: model.PeerUser}, c.ReactedByMember),
		Receiver:      withMembership(model.Peer{ID: receiverID, Type: model.PeerUser}, c.ReceiverMember),
		InReplyTo:     c.InReplyTo,
		ButtonCode:    c.ButtonCode,
		CallbackData:  c.CallbackData,
		ReactedAt:     c.ReactedAt,
		DomainID:      c.DomainID,
		ThreadID:      c.ThreadID,
		UpdatesCursor: c.UpdatesCursor,
	}, nil
}

// withMembership overlays the member id, role and bot flag of m onto p. A
// missing or foreign membership (another contact id) leaves p untouched.
func withMembership(p model.Peer, m *Recipient) model.Peer {
	if m == nil || m.ContactID != p.ID.String() || m.MemberID == "" {
		return p
	}

	p.MemberID = m.MemberID
	p.Role = int32(m.Role)
	p.IsBot = m.IsBot

	return p
}
