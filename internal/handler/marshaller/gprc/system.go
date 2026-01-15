package grpcmarshaller

import (
	impb "github.com/webitel/im-delivery-service/gen/go/api/v1"
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
