package grpcmarshaller

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/webitel/im-delivery-service/internal/domain/model"
)

// TestMarshalMessagePayloadUpdateSeq verifies that UpdateSeq is carried
// from the domain model to the gRPC NewMessageEvent.
func TestMarshalMessagePayloadUpdateSeq(t *testing.T) {
	msg := &model.Message{
		ID:        uuid.New(),
		ThreadID:  uuid.New(),
		Text:      "test",
		CreatedAt: 1000,
		UpdateSeq: 42,
		From:      model.Peer{MemberID: "m1", Role: 1},
	}

	got := marshalMessagePayload(msg)

	if got.MessageEvent.GetUpdateSeq() != 42 {
		t.Errorf("UpdateSeq = %d, want 42", got.MessageEvent.GetUpdateSeq())
	}
}

// TestMarshalMessageEditedPayloadUpdateSeq verifies that UpdateSeq is carried
// from the domain model to the gRPC MessageEditedEvent.
func TestMarshalMessageEditedPayloadUpdateSeq(t *testing.T) {
	msg := &model.MessageEdited{
		ID:        uuid.New(),
		ThreadID:  uuid.New(),
		Text:      "edited",
		EditedAt:  2000,
		UpdateSeq: 100,
		EditedBy:  model.Peer{MemberID: "m1", Role: 1},
	}

	got := marshalMessageEditedPayload(msg)

	if got.MessageEditedEvent.GetUpdateSeq() != 100 {
		t.Errorf("UpdateSeq = %d, want 100", got.MessageEditedEvent.GetUpdateSeq())
	}
}

// TestMarshalMessageDeletedPayloadUpdateSeq verifies that UpdateSeq is carried
// from the domain model to the gRPC MessageDeletedEvent.
func TestMarshalMessageDeletedPayloadUpdateSeq(t *testing.T) {
	msg := &model.MessageDeleted{
		ID:        uuid.New(),
		ThreadID:  uuid.New(),
		DeletedAt: 3000,
		UpdateSeq: 50,
		DeletedBy: model.Peer{MemberID: "m1", Role: 1},
	}

	got := marshalMessageDeletedPayload(msg)

	if got.MessageDeletedEvent.GetUpdateSeq() != 50 {
		t.Errorf("UpdateSeq = %d, want 50", got.MessageDeletedEvent.GetUpdateSeq())
	}
}

// TestMarshalMessageReactionPayloadUpdateSeq verifies that UpdateSeq is carried
// from the domain model to the gRPC MessageReactionEvent.
func TestMarshalMessageReactionPayloadUpdateSeq(t *testing.T) {
	msg := &model.MessageReaction{
		ID:        uuid.New(),
		ThreadID:  uuid.New(),
		Emoji:     "🔥",
		ReactedAt: 4000,
		UpdateSeq: 75,
		Reactor:   model.Peer{MemberID: "m1", Role: 1},
	}

	got := marshalMessageReactionPayload(msg)

	if got.MessageReactionEvent.GetUpdateSeq() != 75 {
		t.Errorf("UpdateSeq = %d, want 75", got.MessageReactionEvent.GetUpdateSeq())
	}
}

// TestMarshalMemberChangedPayload verifies that MemberEvent including
// UpdateSeq and Action are correctly mapped to gRPC MemberChangedEvent.
func TestMarshalMemberChangedPayload(t *testing.T) {
	threadID := uuid.New()
	contactID := uuid.New()

	tests := []struct {
		name   string
		action string
		seq    int64
	}{
		{"member joined", "joined", 42},
		{"member left", "left", 99},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			member := &model.MemberEvent{
				ThreadID:  threadID,
				ContactID: contactID,
				Action:    tt.action,
				UpdateSeq: tt.seq,
			}

			got := marshalMemberChangedPayload(member)

			if got.MemberChangedEvent.GetThreadId() != threadID.String() {
				t.Errorf("ThreadId = %s, want %s", got.MemberChangedEvent.GetThreadId(), threadID.String())
			}

			if got.MemberChangedEvent.GetContactId() != contactID.String() {
				t.Errorf("ContactId = %s, want %s", got.MemberChangedEvent.GetContactId(), contactID.String())
			}

			if got.MemberChangedEvent.GetAction() != tt.action {
				t.Errorf("Action = %s, want %s", got.MemberChangedEvent.GetAction(), tt.action)
			}

			if got.MemberChangedEvent.GetUpdateSeq() != tt.seq {
				t.Errorf("UpdateSeq = %d, want %d", got.MemberChangedEvent.GetUpdateSeq(), tt.seq)
			}
		})
	}
}

// thread-service sends only up_to_seq; a missing message id must stay empty, not zeros.
func TestMarshalMessageStatusPayload_OmitsMissingUpToMessageID(t *testing.T) {
	got := marshalMessageStatusPayload(&model.MessageStatusUpdate{Status: "read", UpToSeq: 1}).MessageStatusEvent

	if got.GetUpToMessageId() != "" || got.GetUpToSeq() != 1 {
		t.Fatalf("up_to_message_id=%q up_to_seq=%d, want empty and 1", got.GetUpToMessageId(), got.GetUpToSeq())
	}

	raw, err := json.Marshal(&model.MessageStatusUpdate{Status: "read", UpToSeq: 1})
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(string(raw), "up_to_message_id") {
		t.Fatalf("JSON must omit up_to_message_id: %s", raw)
	}
}
