package model

import (
	"strconv"

	"github.com/google/uuid"
	"github.com/webitel/im-delivery-service/internal/service/dto"
)

// Interface guard
var _ InboundEventer = (*MessageV1Event)(nil)

// MessageV1Event adapts the current RabbitMQ DTO to the domain InboundEventer.
type MessageV1Event struct {
	raw    *dto.MessageV1
	userID uuid.UUID
	cached any
}

func NewMessageV1Event(raw *dto.MessageV1, userID uuid.UUID) InboundEventer {
	return &MessageV1Event{raw: raw, userID: userID}
}

// GetPayload performs the transformation from V1 DTO to the unified Message entity.
func (e *MessageV1Event) GetPayload() any {
	return NewMessage(
		e.raw.MessageID,
		e.raw.ThreadID,
		e.raw.FromID,
		e.raw.Body,
		e.raw.OccurredAt,
		e.userID,
		e.mapImages(),
		e.mapDocs(),
	)
}

func (e *MessageV1Event) mapImages() (res []*Image) {
	for _, img := range e.raw.Images {
		res = append(res, &Image{
			ID:       strconv.FormatInt(img.FileID, 10),
			FileName: img.Name,
			MimeType: img.Mime,
		})
	}
	return
}

func (e *MessageV1Event) mapDocs() (res []*Document) {
	for _, doc := range e.raw.Documents {
		res = append(res, &Document{
			ID:       strconv.FormatInt(doc.FileID, 10),
			FileName: doc.Name,
			MimeType: doc.Mime,
			Size:     doc.Size,
		})
	}
	return
}

// InboundEventer interface implementation
func (e *MessageV1Event) GetID() string                     { return e.raw.MessageID }
func (e *MessageV1Event) GetTraceID() string                { return e.raw.MessageID }
func (e *MessageV1Event) GetKind() InboundEventKind         { return MessageCreated }
func (e *MessageV1Event) GetPriority() InboundEventPriority { return PriorityHigh }
func (e *MessageV1Event) GetUserID() uuid.UUID              { return e.userID }
func (e *MessageV1Event) GetOccurredAt() int64              { return safeParseRFC3339(e.raw.OccurredAt) }
func (e *MessageV1Event) GetCached() any                    { return e.cached }
func (e *MessageV1Event) SetCached(v any)                   { e.cached = v }
