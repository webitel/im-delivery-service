package marshaller

import (
	impb "github.com/webitel/im-delivery-service/gen/go/api/v1"
	"github.com/webitel/im-delivery-service/internal/domain/model"
)

// MarshallDeliveryEvent transforms domain Eventer to Protobuf ServerEvent.
// It leverages internal caching to ensure that expensive mapping and string
// conversions happen only once per event, even with multiple subscribers.
func MarshallDeliveryEvent(ev model.Eventer) *impb.ServerEvent {
	//  Return cached Protobuf structure if already computed.
	if cached := ev.GetCached(); cached != nil {
		if pb, ok := cached.(*impb.ServerEvent); ok {
			return pb
		}
	}

	//  Perform mapping.
	res := &impb.ServerEvent{
		Id:        ev.GetID(),
		CreatedAt: ev.GetOccurredAt(),
		Priority:  mapPriority(ev.GetPriority()),
	}

	// Use specialized mappers based on payload type.
	switch p := ev.GetPayload().(type) {
	case *model.Message:
		res.Payload = marshalMessagePayload(p)
	case *model.ConnectedPayload:
		res.Payload = marshalConnectedPayload(p)
	}

	// STORE: Save for subsequent delivery attempts (other devices/retry).
	ev.SetCached(res)

	return res
}

// marshalMessagePayload encapsulates the transformation of a domain message.
func marshalMessagePayload(m *model.Message) *impb.ServerEvent_MessageEvent {
	if m == nil {
		return nil
	}
	return &impb.ServerEvent_MessageEvent{
		MessageEvent: &impb.NewMessageEvent{
			Id:      m.ID.String(),
			Message: mapThreadMessage(m),
		},
	}
}

// marshalConnectedPayload handles the system connection event mapping.
func marshalConnectedPayload(p *model.ConnectedPayload) *impb.ServerEvent_ConnectedEvent {
	return &impb.ServerEvent_ConnectedEvent{
		ConnectedEvent: &impb.ConnectedEvent{
			Ok:            p.Ok,
			ConnectionId:  p.ConnectionID,
			ServerVersion: p.ServerVersion,
		},
	}
}

// mapThreadMessage performs detailed mapping of the message core and its media content.
func mapThreadMessage(m *model.Message) *impb.ThreadMessage {
	msg := &impb.ThreadMessage{
		Id:       m.ID.String(),
		ThreadId: m.ThreadID.String(),
		Text:     m.Text,
		// Using kind-specific peer mapping
		From: &impb.Peer{Kind: &impb.Peer_UserId{UserId: m.From.ID.String()}},
		To:   &impb.Peer{Kind: &impb.Peer_UserId{UserId: m.To.ID.String()}},
	}

	// Logic to discriminate between different message types based on media presence.
	if len(m.Images) > 0 {
		msg.Type = impb.MessageType_IMAGE
		msg.Content = &impb.ThreadMessage_Image{
			Image: &impb.Image{
				Id:       m.Images[0].ID,
				FileName: m.Images[0].FileName,
				MimeType: m.Images[0].MimeType,
			},
		}
	} else if len(m.Documents) > 0 {
		msg.Type = impb.MessageType_DOCUMENT
		msg.Content = &impb.ThreadMessage_Document{
			Document: &impb.Document{
				FileName:  m.Documents[0].FileName,
				MimeType:  m.Documents[0].MimeType,
				SizeBytes: m.Documents[0].Size,
			},
		}
	} else {
		msg.Type = impb.MessageType_TEXT
	}

	return msg
}

// Helper to map domain priorities to Protobuf-specific values
func mapPriority(p model.EventPriority) impb.EventPriority {
	switch p {
	case model.PriorityLow:
		return impb.EventPriority_LOW
	case model.PriorityNormal:
		return impb.EventPriority_NORMAL
	case model.PriorityHigh:
		return impb.EventPriority_HIGH
	default:
		return impb.EventPriority_PRIORITY_UNSPECIFIED
	}
}
