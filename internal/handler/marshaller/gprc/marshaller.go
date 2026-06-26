package grpcmarshaller

import (
	impb "github.com/webitel/im-delivery-service/gen/go/delivery/v1"
	"github.com/webitel/im-delivery-service/internal/domain/event"
	"github.com/webitel/im-delivery-service/internal/domain/model"
	"github.com/webitel/im-delivery-service/internal/handler/marshaller"
)

// INTERFACE GUARD
var _ marshaller.EventMarshaller = (*Marshaller)(nil)

type Marshaller struct{}

func New() *Marshaller { return &Marshaller{} }

func (m *Marshaller) Marshal(ev event.Eventer) (any, error) {
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
	case *model.ConnectedPayload:
		res.Payload = &impb.ServerEvent_ConnectedEvent{ConnectedEvent: &impb.ConnectedEvent{
			Ok:            p.Ok,
			ConnectionId:  p.ConnectionID,
			ServerVersion: model.ServerVersion,
		}}
	case *model.DisconnectedPayload:
		res.Payload = &impb.ServerEvent_DisconnectedEvent{DisconnectedEvent: &impb.DisconnectedEvent{Reason: p.Reason}}
	}

	ev.SetCached(res)

	return res, nil
}
