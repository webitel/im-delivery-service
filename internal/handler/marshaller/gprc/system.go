package grpcmarshaller

import (
	impb "github.com/webitel/im-delivery-service/gen/go/delivery/v1"
	"github.com/webitel/im-delivery-service/internal/domain/model"
)

// marshalConnectedPayload maps system connection data to PB.
func marshalConnectedPayload(p *model.ConnectedPayload) *impb.ServerEvent_ConnectedEvent {
	if p == nil {
		return nil
	}
	return &impb.ServerEvent_ConnectedEvent{
		ConnectedEvent: &impb.ConnectedEvent{
			Ok:            p.Ok,
			ConnectionId:  p.ConnectionID,
			ServerVersion: p.ServerVersion,
		},
	}
}

// marshalDisconnectedPayload maps system closure notification to PB.
func marshalDisconnectedPayload(p *model.DisconnectedPayload) *impb.ServerEvent_DisconnectedEvent {
	if p == nil {
		return nil
	}
	return &impb.ServerEvent_DisconnectedEvent{
		DisconnectedEvent: &impb.DisconnectedEvent{
			Reason: p.Reason,
			Code:   p.Code, // ensure 'code' field exists in your .proto file
		},
	}
}
