package mapper

import (
	"time"

	"github.com/google/uuid"
	"github.com/webitel/im-delivery-service/internal/domain/model"
	"github.com/webitel/im-delivery-service/internal/service/dto"
)

type MessageV1Adapter struct {
	raw    *dto.MessageV1
	userID uuid.UUID
	cached any
}

func NewMessageV1Adapter(raw *dto.MessageV1, userID uuid.UUID) model.Eventer {
	return &MessageV1Adapter{raw: raw, userID: userID}
}

func (a *MessageV1Adapter) GetID() string                    { return a.raw.MessageID }
func (a *MessageV1Adapter) GetKind() model.EventKind         { return model.MessageCreated }
func (a *MessageV1Adapter) GetPriority() model.EventPriority { return model.PriorityHigh }
func (a *MessageV1Adapter) GetTraceID() string               { return a.raw.MessageID }
func (a *MessageV1Adapter) GetUserID() uuid.UUID             { return a.userID }
func (a *MessageV1Adapter) GetCached() any                   { return a.cached }
func (a *MessageV1Adapter) SetCached(v any)                  { a.cached = v }

func (a *MessageV1Adapter) GetOccurredAt() int64 {
	t, err := time.Parse(time.RFC3339, a.raw.OccurredAt)
	if err != nil {
		return time.Now().UnixMilli()
	}
	return t.UnixMilli()
}

func (a *MessageV1Adapter) GetPayload() any {
	// Map FromID and ToID from raw DTO to domain model.
	// Without this, the marshaller sees empty UUIDs (00000000-...).

	// Pre-parse to handle potential malformed UUIDs safely
	fromID, _ := uuid.Parse(a.raw.FromID)
	threadID, _ := uuid.Parse(a.raw.ThreadID)
	msgID, _ := uuid.Parse(a.raw.MessageID)

	return &model.Message{
		ID:        msgID,
		ThreadID:  threadID,
		Text:      a.raw.Body,
		CreatedAt: a.GetOccurredAt(),
		// [IMPORTANT] Fill the Peer information
		From: model.Peer{
			ID: fromID,
		},
		To: model.Peer{
			ID: a.userID, // The recipient is the user this adapter was created for
		},
	}
}
