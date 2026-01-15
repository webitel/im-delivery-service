package grpcmarshaller

import (
	impb "github.com/webitel/im-delivery-service/gen/go/api/v1"
	"github.com/webitel/im-delivery-service/internal/domain/model"
)

// marshalMessagePayload maps the business Message entity to PB.
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

// mapThreadMessage converts core message data and handles media/V2 logic.
func mapThreadMessage(m *model.Message) *impb.ThreadMessage {
	msg := &impb.ThreadMessage{
		Id:        m.ID.String(),
		ThreadId:  m.ThreadID.String(),
		Text:      m.Text,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt, // Added to support V2 edits
		From:      &impb.Peer{Kind: &impb.Peer_UserId{UserId: m.From.ID.String()}},
		To:        &impb.Peer{Kind: &impb.Peer_UserId{UserId: m.To.ID.String()}},
	}

	// Mapping media content
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
