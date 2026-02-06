package grpcmarshaller

import (
	impb "github.com/webitel/im-delivery-service/gen/go/delivery/v1"
	"github.com/webitel/im-delivery-service/internal/domain/event"
	"github.com/webitel/im-delivery-service/internal/domain/model"
)

func mapPriority(p event.EventPriority) impb.EventPriority {
	switch p {
	case event.PriorityLow:
		return impb.EventPriority_LOW
	case event.PriorityNormal:
		return impb.EventPriority_NORMAL
	case event.PriorityHigh:
		return impb.EventPriority_HIGH
	default:
		return impb.EventPriority_PRIORITY_UNSPECIFIED
	}
}

func marshalPeer(p model.Peer) *impb.Peer {
	res := &impb.Peer{}
	switch p.Type {
	case model.PeerUser:
		res.Kind = &impb.Peer_UserId{UserId: p.Sub}
	case model.PeerGroup:
		res.Kind = &impb.Peer_ChatId{ChatId: p.Sub}
	case model.PeerChannel:
		res.Kind = &impb.Peer_ChannelId{ChannelId: p.Sub}
	}
	if p.IsEnriched() {
		res.Identity = &impb.Identity{Issuer: p.Issuer, Name: p.Name}
	}
	return res
}

func marshalMessagePayload(m *model.Message) *impb.ServerEvent_MessageEvent {
	msg := &impb.ThreadMessage{
		Id: m.ID.String(), ThreadId: m.ThreadID.String(), Text: m.Text,
		CreatedAt: m.CreatedAt, EditedAt: m.EditedAt,
		From: marshalPeer(m.From), To: marshalPeer(m.To),
	}
	if len(m.Images) > 0 {
		msg.Type = impb.MessageType_IMAGE
		msg.Content = &impb.ThreadMessage_Image{Image: &impb.Image{Id: m.Images[0].ID, Url: m.Images[0].URL}}
	} else if len(m.Documents) > 0 {
		msg.Type = impb.MessageType_DOCUMENT
		msg.Content = &impb.ThreadMessage_Document{Document: &impb.Document{Id: m.Documents[0].ID, FileName: m.Documents[0].FileName}}
	} else {
		msg.Type = impb.MessageType_TEXT
	}
	return &impb.ServerEvent_MessageEvent{MessageEvent: &impb.NewMessageEvent{Message: msg}}
}
