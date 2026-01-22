package grpcmarshaller

import (
	impb "github.com/webitel/im-delivery-service/gen/go/delivery/v1"
	"github.com/webitel/im-delivery-service/internal/domain/event"
	"github.com/webitel/im-delivery-service/internal/domain/model"
)

// MarshallDeliveryEvent transforms domain Eventer to Protobuf ServerEvent.
// It acts as a gateway and uses type-specific marshallers.
func MarshallDeliveryEvent(ev event.Eventer) *impb.ServerEvent {
	// 1. [PERFORMANCE] Check cache first.
	if cached := ev.GetCached(); cached != nil {
		if pb, ok := cached.(*impb.ServerEvent); ok {
			return pb
		}
	}

	// 2. Base event mapping.
	res := &impb.ServerEvent{
		Id:        ev.GetID(),
		CreatedAt: ev.GetOccurredAt(),
		Priority:  mapPriority(ev.GetPriority()),
	}

	// 3. [STRATEGY] Route to specific logic based on payload type.
	switch p := ev.GetPayload().(type) {
	case *model.Message:
		res.Payload = marshalMessagePayload(p)
	case *model.ConnectedPayload:
		res.Payload = marshalConnectedPayload(p)
	}

	// 4. [CACHE] Save the result back.
	ev.SetCached(res)
	return res
}
