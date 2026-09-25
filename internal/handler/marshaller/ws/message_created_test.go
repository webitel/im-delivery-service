package wsmarshaller

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/webitel/im-delivery-service/internal/domain/model"
)

// TestMapMessageUpdateSeq verifies that UpdateSeq is carried from the domain
// model to the WebSocket message DTO and is serialized to JSON.
func TestMapMessageUpdateSeq(t *testing.T) {
	msg := &model.Message{
		ID:        uuid.New(),
		ThreadID:  uuid.New(),
		Text:      "test message",
		CreatedAt: 1000,
		UpdateSeq: 42,
		From:      model.Peer{MemberID: "m1", Role: 1},
	}

	got := mapMessage(msg)

	if got.UpdateSeq != 42 {
		t.Errorf("UpdateSeq = %d, want 42", got.UpdateSeq)
	}

	// Verify it serializes to JSON with the field
	data, err := json.Marshal(got)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	if !contains(string(data), `"update_seq":42`) && !contains(string(data), `"update_seq": 42`) {
		t.Errorf("JSON does not contain update_seq field: %s", data)
	}
}

// TestMapMessageEditedUpdateSeq verifies that UpdateSeq is carried from the
// domain model to the WebSocket edited message DTO.
func TestMapMessageEditedUpdateSeq(t *testing.T) {
	msg := &model.MessageEdited{
		ID:        uuid.New(),
		ThreadID:  uuid.New(),
		Text:      "edited",
		EditedAt:  2000,
		UpdateSeq: 100,
		EditedBy:  model.Peer{MemberID: "m1", Role: 1},
	}

	got := mapMessageEdited(msg)

	if got.UpdateSeq != 100 {
		t.Errorf("UpdateSeq = %d, want 100", got.UpdateSeq)
	}

	data, err := json.Marshal(got)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	if !contains(string(data), `"update_seq":100`) && !contains(string(data), `"update_seq": 100`) {
		t.Errorf("JSON does not contain update_seq field: %s", data)
	}
}

// TestMapMessageDeletedUpdateSeq verifies that UpdateSeq is carried from the
// domain model to the WebSocket deleted message DTO.
func TestMapMessageDeletedUpdateSeq(t *testing.T) {
	msg := &model.MessageDeleted{
		ID:        uuid.New(),
		ThreadID:  uuid.New(),
		DeletedAt: 3000,
		UpdateSeq: 50,
		DeletedBy: model.Peer{MemberID: "m1", Role: 1},
	}

	got := mapMessageDeleted(msg)

	if got.UpdateSeq != 50 {
		t.Errorf("UpdateSeq = %d, want 50", got.UpdateSeq)
	}

	data, err := json.Marshal(got)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	if !contains(string(data), `"update_seq":50`) && !contains(string(data), `"update_seq": 50`) {
		t.Errorf("JSON does not contain update_seq field: %s", data)
	}
}

// TestMapReactionUpdateSeq verifies that UpdateSeq is carried from the domain
// model to the WebSocket reaction DTO.
func TestMapReactionUpdateSeq(t *testing.T) {
	reaction := &model.MessageReaction{
		ID:        uuid.New(),
		ThreadID:  uuid.New(),
		Emoji:     "🔥",
		ReactedAt: 4000,
		UpdateSeq: 75,
		Reactor:   model.Peer{MemberID: "m1", Role: 1},
	}

	got := mapReaction(reaction, uuid.New())

	if got.UpdateSeq != 75 {
		t.Errorf("UpdateSeq = %d, want 75", got.UpdateSeq)
	}

	data, err := json.Marshal(got)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	if !contains(string(data), `"update_seq":75`) && !contains(string(data), `"update_seq": 75`) {
		t.Errorf("JSON does not contain update_seq field: %s", data)
	}
}

func contains(s, substr string) bool {
	for i := range len(s) - len(substr) + 1 {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}

	return false
}
