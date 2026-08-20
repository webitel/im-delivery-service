package grpcmarshaller

import (
	impb "github.com/webitel/im-delivery-service/gen/go/delivery/v1"
	"github.com/webitel/im-delivery-service/internal/domain/event"
	"github.com/webitel/im-delivery-service/internal/domain/model"
)

// mapPriority converts domain priority to Protobuf enum.
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

// marshalPeer converts a single domain Peer to a Protobuf Peer.
func marshalPeer(p *model.Peer) *impb.Peer {
	if p == nil {
		return nil
	}

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

func marshalReplyTo(r *model.ReplyTo) *impb.QuotedMessage {
	if r == nil {
		return nil
	}

	quoted := &impb.QuotedMessage{
		Id:        r.MessageID.String(),
		SenderId:  r.SenderID.String(),
		Type:      r.Type,
		Body:      r.Body,
		CreatedAt: r.CreatedAt,
	}

	if r.AttachmentKind != nil {
		quoted.AttachmentKind = *r.AttachmentKind
	}

	if r.AttachmentName != nil {
		quoted.AttachmentName = *r.AttachmentName
	}

	if r.AttachmentMime != nil {
		quoted.AttachmentMime = *r.AttachmentMime
	}

	return quoted
}

func marshalForwardOrigin(f *model.ForwardOrigin) *impb.ForwardOrigin {
	if f == nil {
		return nil
	}

	origin := &impb.ForwardOrigin{
		Kind:           impb.ForwardOriginKind(f.Kind),
		SenderName:     f.SenderName,
		OriginalSentAt: f.OriginalSentAt,
	}

	if f.SenderID != nil {
		origin.SenderId = f.SenderID.String()
	}

	if f.SourceMessageID != nil {
		origin.SourceMessageId = f.SourceMessageID.String()
	}

	return origin
}

func marshalMessageDeletedPayload(m *model.MessageDeleted) *impb.ServerEvent_MessageDeletedEvent {
	return &impb.ServerEvent_MessageDeletedEvent{
		MessageDeletedEvent: &impb.MessageDeletedEvent{
			Id:        m.ID.String(),
			ThreadId:  m.ThreadID.String(),
			DeletedBy: marshalPeer(&m.DeletedBy),
			DeletedAt: m.DeletedAt,
		},
	}
}

func marshalMessageReactionPayload(m *model.MessageReaction) *impb.ServerEvent_MessageReactionEvent {
	return &impb.ServerEvent_MessageReactionEvent{
		MessageReactionEvent: &impb.MessageReactionEvent{
			MessageId: m.ID.String(),
			ThreadId:  m.ThreadID.String(),
			Reactor:   marshalPeer(&m.Reactor),
			Emoji:     m.Emoji,
			Removed:   m.Removed,
			ReactedAt: m.ReactedAt,
			SendId:    m.SendId,
		},
	}
}

// marshalMessageStatusPayload converts the domain MessageStatusUpdate to a gRPC MessageStatusEvent.
func marshalMessageStatusPayload(m *model.MessageStatusUpdate) *impb.ServerEvent_MessageStatusEvent {
	messageIDs := make([]string, 0, len(m.MessageIDs))
	for _, id := range m.MessageIDs {
		messageIDs = append(messageIDs, id.String())
	}

	statusMap := map[string]impb.MessageDeliveryStatus{
		"delivered": impb.MessageDeliveryStatus_MESSAGE_DELIVERY_STATUS_DELIVERED,
		"read":      impb.MessageDeliveryStatus_MESSAGE_DELIVERY_STATUS_READ,
		"failed":    impb.MessageDeliveryStatus_MESSAGE_DELIVERY_STATUS_FAILED,
	}

	status, ok := statusMap[m.Status]
	if !ok {
		status = impb.MessageDeliveryStatus_MESSAGE_DELIVERY_STATUS_UNSPECIFIED
	}

	return &impb.ServerEvent_MessageStatusEvent{
		MessageStatusEvent: &impb.MessageStatusEvent{
			ThreadId:      m.ThreadID.String(),
			MemberId:      m.MemberID.String(),
			MessageIds:    messageIDs,
			Status:        status,
			Via:           m.Via,
			OccurredAt:    m.OccurredAt,
			UpToMessageId: m.UpToMessageID.String(),
			UpToSeq:       m.UpToSeq,
		},
	}
}

// marshalMessageType maps the domain type name onto the wire enum. The enum
// covers text, document and image only, so a system, interactive, location or
// contact message reports UNSPECIFIED rather than pretending to be text — an
// edit never changes the type, and the client already holds the message.
func marshalMessageType(name string) impb.MessageType {
	switch name {
	case "text":
		return impb.MessageType_TEXT
	case "document":
		return impb.MessageType_DOCUMENT
	case "image":
		return impb.MessageType_IMAGE
	default:
		return impb.MessageType_UNSPECIFIED_MESSAGE_TYPE
	}
}

// marshalMessageEditedPayload converts an edit to a gRPC server event.
func marshalMessageEditedPayload(m *model.MessageEdited) *impb.ServerEvent_MessageEditedEvent {
	return &impb.ServerEvent_MessageEditedEvent{
		MessageEditedEvent: &impb.MessageEditedEvent{
			Id:        m.ID.String(),
			ThreadId:  m.ThreadID.String(),
			EditedBy:  marshalPeer(&m.EditedBy),
			Text:      m.Text,
			Type:      marshalMessageType(m.Type),
			CreatedAt: m.CreatedAt,
			EditedAt:  m.EditedAt,
			Version:   m.Version,
		},
	}
}

// marshalMessagePayload converts the domain Message model to a gRPC server event.
func marshalMessagePayload(m *model.Message) *impb.ServerEvent_MessageEvent {
	// Map the slice of recipients from domain to PB
	recipients := make([]*impb.Peer, 0, len(m.To))
	for i := range m.To {
		recipients = append(recipients, marshalPeer(&m.To[i]))
	}

	msg := &impb.ThreadMessage{
		Id:        m.ID.String(),
		ThreadId:  m.ThreadID.String(),
		Text:      m.Text,
		CreatedAt: m.CreatedAt,
		EditedAt:  m.EditedAt,
		From:      marshalPeer(&m.From),
		ReplyTo:   marshalReplyTo(m.ReplyTo),

		ForwardOrigin: marshalForwardOrigin(m.ForwardOrigin),
	}

	// [CONTENT_TYPE_LOGIC]
	if len(m.Images) > 0 {
		msg.Type = impb.MessageType_IMAGE
		msg.Content = &impb.ThreadMessage_Image{
			Image: &impb.Image{Id: m.Images[0].ID, Url: m.Images[0].URL},
		}
	} else if len(m.Documents) > 0 {
		msg.Type = impb.MessageType_DOCUMENT
		msg.Content = &impb.ThreadMessage_Document{
			Document: &impb.Document{Id: m.Documents[0].ID, FileName: m.Documents[0].Name, Url: m.Documents[0].URL},
		}
	} else {
		msg.Type = impb.MessageType_TEXT
	}

	return &impb.ServerEvent_MessageEvent{
		MessageEvent: &impb.NewMessageEvent{
			Message: msg,
		},
	}
}
