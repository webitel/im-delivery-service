package payload

import (
	"github.com/webitel/im-delivery-service/internal/domain/model"
	"github.com/webitel/im-delivery-service/internal/domain/util"
)

// MemberContact is the contact im-thread-service resolved for the actor of an
// event, so this service does not have to look it up again.
type MemberContact struct {
	ID       string `json:"id"`
	Sub      string `json:"sub"`
	Iss      string `json:"iss"`
	Name     string `json:"name"`
	Username string `json:"username"`
	Type     string `json:"type"`
	IsBot    bool   `json:"is_bot"`
}

// Member is a thread participant addressed by its membership id, carrying the
// contact behind it. Role is the ThreadRole name, not its ordinal.
type Member struct {
	ID      string         `json:"id"`
	Contact *MemberContact `json:"contact"`
	Role    string         `json:"role"`
}

// ToPeer flattens the member onto the internal peer shape. An event without a
// contact block still yields the routing identity, just without enrichment.
func (m Member) ToPeer() model.Peer {
	peer := model.Peer{
		MemberID: m.ID,
		Role:     int32(model.ParseRoleName(m.Role)),
	}

	if m.Contact == nil {
		return peer
	}

	peer.ID = util.SafeParseUUID(m.Contact.ID)
	peer.Type = model.ParsePeerType(m.Contact.Type)
	peer.Sub = m.Contact.Sub
	peer.Issuer = m.Contact.Iss
	peer.Name = m.Contact.Name
	peer.Username = m.Contact.Username
	peer.ContactType = m.Contact.Type
	peer.IsBot = m.Contact.IsBot

	return peer
}

func (m Member) ContactID() string {
	if m.Contact == nil {
		return ""
	}

	return m.Contact.ID
}

// MessageDeletedV1 is the im_message.<thread_id>.message.deleted.v1 event
// published by im-thread-service. It carries no content by design.
type MessageDeletedV1 struct {
	MessageID  string      `json:"message_id"`
	ThreadID   string      `json:"thread_id"`
	DomainID   int32       `json:"domain_id"`
	DeletedBy  Member      `json:"deleted_by"`
	To         []Recipient `json:"to"`
	Type       string      `json:"type"`
	CreatedAt  string      `json:"created_at"`
	OccurredAt string      `json:"occurred_at"`
}

func (d *MessageDeletedV1) ToDomain() *model.MessageDeleted {
	var createdAt int64
	if d.CreatedAt != "" {
		createdAt = util.SafeParseRFC3339(d.CreatedAt)
	}

	msg := &model.MessageDeleted{
		ID:        util.SafeParseUUID(d.MessageID),
		ThreadID:  util.SafeParseUUID(d.ThreadID),
		DomainID:  int64(d.DomainID),
		CreatedAt: createdAt,
		DeletedAt: util.SafeParseRFC3339(d.OccurredAt),
		DeletedBy: d.DeletedBy.ToPeer(),
	}

	msg.To = make([]model.Peer, 0, len(d.To))
	for _, r := range d.To {
		msg.To = append(msg.To, model.Peer{
			ID:       util.SafeParseUUID(r.ContactID),
			MemberID: r.MemberID,
			Role:     int32(r.Role),
		})
	}

	return msg
}
