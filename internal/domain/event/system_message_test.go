package event

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/webitel/im-delivery-service/internal/domain/model"
)

func TestSystemMessageType(t *testing.T) {
	userID := uuid.New()

	t.Run("non-MessageCreated event returns ok=false", func(t *testing.T) {
		ev := &Envelope[any]{
			Kind: Connected,
		}

		systemType, ok := SystemMessageType(ev)
		if ok {
			t.Errorf("SystemMessageType(Connected) returned ok=true, want false")
		}

		if systemType != "" {
			t.Errorf("SystemMessageType(Connected) returned %q, want empty string", systemType)
		}
	})

	t.Run("MessageCreated with nil System returns ok=false", func(t *testing.T) {
		msg := &model.Message{
			ID:     uuid.New(),
			From:   model.Peer{Name: "Alice"},
			System: nil,
		}
		ev := &Envelope[*model.Message]{
			Kind:    MessageCreated,
			Payload: msg,
			UserID:  userID,
		}

		systemType, ok := SystemMessageType(ev)
		if ok {
			t.Errorf("SystemMessageType(MessageCreated with nil System) returned ok=true, want false")
		}

		if systemType != "" {
			t.Errorf("SystemMessageType(MessageCreated with nil System) returned %q, want empty string", systemType)
		}
	})

	t.Run("MessageCreated with empty System.Type returns ok=false", func(t *testing.T) {
		msg := &model.Message{
			ID:   uuid.New(),
			From: model.Peer{Name: "Alice"},
			System: &model.System{
				Type: "",
			},
		}
		ev := &Envelope[*model.Message]{
			Kind:    MessageCreated,
			Payload: msg,
			UserID:  userID,
		}

		systemType, ok := SystemMessageType(ev)
		if ok {
			t.Errorf("SystemMessageType(MessageCreated with empty System.Type) returned ok=true, want false")
		}

		if systemType != "" {
			t.Errorf("SystemMessageType(MessageCreated with empty System.Type) returned %q, want empty string", systemType)
		}
	})

	t.Run("MessageCreated with System.Type returns the type and ok=true", func(t *testing.T) {
		msg := &model.Message{
			ID:        uuid.New(),
			From:      model.Peer{Name: "Alice"},
			CreatedAt: time.Now().UnixMilli(),
			System: &model.System{
				Type: "user_joined",
			},
		}
		ev := &Envelope[*model.Message]{
			Kind:    MessageCreated,
			Payload: msg,
			UserID:  userID,
		}

		systemType, ok := SystemMessageType(ev)
		if !ok {
			t.Errorf("SystemMessageType(MessageCreated with 'user_joined') returned ok=false, want true")
		}

		if systemType != "user_joined" {
			t.Errorf("SystemMessageType(MessageCreated with 'user_joined') returned %q, want 'user_joined'", systemType)
		}
	})
}
