package grpcmarshaller

import (
	"github.com/google/uuid"

	impb "github.com/webitel/im-delivery-service/gen/go/delivery/v1"
	"github.com/webitel/im-delivery-service/internal/domain/event"
	"github.com/webitel/im-delivery-service/internal/domain/model"
	"github.com/webitel/im-delivery-service/internal/handler/marshaller"
)

// INTERFACE GUARD
var _ marshaller.EventMarshaller = (*Marshaller)(nil)

type Marshaller struct{}

func New() *Marshaller { return &Marshaller{} }

// Marshal ignores viewer: the delivery proto's MessageReactionEvent carries no
// per-emoji aggregate, so there is no per-viewer field to resolve here.
func (m *Marshaller) Marshal(ev event.Eventer, _ uuid.UUID) (any, error) {
	// [CACHE] Retrieve if already marshaled for this event
	if cached := ev.GetCached(); cached != nil {
		if pb, ok := cached.(*impb.ServerEvent); ok {
			return pb, nil
		}
	}

	res := &impb.ServerEvent{
		Id: ev.GetID(), CreatedAt: ev.GetOccurredAt(),
		Priority: mapPriority(ev.GetPriority()),
	}

	switch p := ev.GetPayload().(type) {
	case *model.Message:
		res.Payload = marshalMessagePayload(p)
	case *model.MessageDeleted:
		res.Payload = marshalMessageDeletedPayload(p)
	case *model.MessageReaction:
		res.Payload = marshalMessageReactionPayload(p)
	case *model.MessageStatusUpdate:
		res.Payload = marshalMessageStatusPayload(p)
	case *model.ConnectedPayload:
		res.Payload = &impb.ServerEvent_ConnectedEvent{ConnectedEvent: &impb.ConnectedEvent{
			Ok:            p.Ok,
			ConnectionId:  p.ConnectionID,
			ServerVersion: model.ServerVersion,
		}}
	case *model.DisconnectedPayload:
		res.Payload = &impb.ServerEvent_DisconnectedEvent{DisconnectedEvent: &impb.DisconnectedEvent{Reason: p.Reason}}
	case *model.Typing:
		// member is the enriched typing participant, marshalled with the SAME
		// helper as a message sender (marshalPeer) — identical shape to a
		// NewMessageEvent's `from`.
		te := &impb.TypingEvent{
			ThreadId:  p.ThreadID,
			TimeoutMs: p.TimeoutMs,
			Member:    marshalPeer(&p.From),
		}

		// preview_text is optional; attach only when this session is authorized.
		if p.PreviewText != "" {
			preview := p.PreviewText
			te.PreviewText = &preview
		}

		res.Payload = &impb.ServerEvent_TypingEvent{TypingEvent: te}
	}

	ev.SetCached(res)

	return res, nil
}
