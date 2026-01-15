package model

import (
	"strconv"

	"github.com/google/uuid"
	"github.com/webitel/im-delivery-service/internal/service/dto"
)

// MessageV2Event implements InboundEventer for the new V2 payload.
type MessageV2Event struct {
	raw    *dto.MessageV2
	userID uuid.UUID
	cached any
}

func NewMessageV2Event(raw *dto.MessageV2, userID uuid.UUID) InboundEventer {
	return &MessageV2Event{raw: raw, userID: userID}
}

// GetPayload performs the transformation.
// It handles new V2 fields: EditedAt, Metadata, and Forwarding.
func (e *MessageV2Event) GetPayload() any {
	// Initialize core message with standard fields
	msg := &Message{
		ID:        safeParseUUID(e.raw.MessageID),
		ThreadID:  safeParseUUID(e.raw.ThreadID),
		Text:      e.raw.Body,
		CreatedAt: e.raw.OccurredAt, // Already int64 in V2
		UpdatedAt: e.raw.EditedAt,   // Mapping the new field
		From:      Peer{ID: safeParseUUID(e.raw.FromID), Type: PeerUser},
		To:        Peer{ID: e.userID, Type: PeerUser},
		Images:    e.mapImages(),
		Documents: e.mapDocs(),
	}

	// Mapping new V2-specific metadata
	if len(e.raw.Metadata) > 0 {
		msg.Metadata = make(map[string]any)
		for k, v := range e.raw.Metadata {
			msg.Metadata[k] = v
		}
		// Example of internal business logic: add forward flag to metadata
		msg.Metadata["internal_forwarded"] = e.raw.IsForward
	}

	return msg
}

// mapImages converts V2 image DTOs to Domain Image entities.
func (e *MessageV2Event) mapImages() (res []*Image) {
	for _, img := range e.raw.Images {
		res = append(res, &Image{
			ID:       strconv.FormatInt(img.FileID, 10),
			FileName: img.Name,
			MimeType: img.Mime,
		})
	}
	return
}

// mapDocs converts V2 document DTOs to Domain Document entities.
func (e *MessageV2Event) mapDocs() (res []*Document) {
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

// --- InboundEventer interface implementation ---

func (e *MessageV2Event) GetID() string                     { return e.raw.MessageID }
func (e *MessageV2Event) GetTraceID() string                { return e.raw.MessageID }
func (e *MessageV2Event) GetKind() InboundEventKind         { return MessageCreated }
func (e *MessageV2Event) GetPriority() InboundEventPriority { return PriorityHigh }
func (e *MessageV2Event) GetUserID() uuid.UUID              { return e.userID }
func (e *MessageV2Event) GetOccurredAt() int64              { return e.raw.OccurredAt }
func (e *MessageV2Event) GetCached() any                    { return e.cached }
func (e *MessageV2Event) SetCached(v any)                   { e.cached = v }
