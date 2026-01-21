package grpcmarshaller

import (
	impb "github.com/webitel/im-delivery-service/gen/go/delivery/v1"
	"github.com/webitel/im-delivery-service/internal/domain/model"
)

// marshalMessagePayload maps domain Message to Protobuf payload wrapper.
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

// mapThreadMessage performs detailed mapping of the message body and metadata.
func mapThreadMessage(m *model.Message) *impb.ThreadMessage {
	msg := &impb.ThreadMessage{
		Id:        m.ID.String(),
		ThreadId:  m.ThreadID.String(),
		Text:      m.Text,
		CreatedAt: m.CreatedAt,
		EditedAt:  m.EditedAt,
		From:      marshalPeer(m.From),
		To:        marshalPeer(m.To),
	}

	// [CONTENT_SELECTION] Map the primary attachment based on domain availability.
	switch {
	case len(m.Images) > 0:
		img := m.Images[0]
		msg.Type = impb.MessageType_IMAGE
		msg.Content = &impb.ThreadMessage_Image{
			Image: &impb.Image{
				Id:       img.ID,
				FileName: img.FileName,
				MimeType: img.MimeType,
			},
		}
	case len(m.Documents) > 0:
		doc := m.Documents[0]
		msg.Type = impb.MessageType_DOCUMENT
		msg.Content = &impb.ThreadMessage_Document{
			Document: &impb.Document{
				Id:        doc.ID,
				FileName:  doc.FileName,
				MimeType:  doc.MimeType,
				SizeBytes: doc.Size,
			},
		}
	default:
		msg.Type = impb.MessageType_TEXT
	}

	return msg
}

// marshalPeer maps participant information to Protobuf Peer structure.
func marshalPeer(p model.Peer) *impb.Peer {
	res := &impb.Peer{}

	switch p.Type {
	case model.PeerUser:
		res.Kind = &impb.Peer_UserId{UserId: p.Sub}
	case model.PeerBot:
		res.Kind = &impb.Peer_BotId{BotId: p.Sub}
	case model.PeerChat:
		res.Kind = &impb.Peer_ChatId{ChatId: p.Sub}
	case model.PeerChannel:
		res.Kind = &impb.Peer_ChannelId{ChannelId: p.Sub}
	}

	if p.IsEnriched() {
		res.Identity = &impb.Identity{
			Issuer: p.Issuer,
			Name:   p.Name,
		}
	}

	return res
}
