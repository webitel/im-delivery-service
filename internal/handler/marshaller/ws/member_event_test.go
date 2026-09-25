package wsmarshaller

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/webitel/im-delivery-service/internal/domain/event"
	"github.com/webitel/im-delivery-service/internal/domain/model"
)

// TestMarshalMemberJoinedWithUpdateSeq verifies that a member joined event
// with UpdateSeq is correctly marshaled to WebSocket JSON.
func TestMarshalMemberJoinedWithUpdateSeq(t *testing.T) {
	threadID := uuid.New()
	contactID := uuid.New()

	member := &model.MemberEvent{
		ThreadID:  threadID,
		ContactID: contactID,
		Metadata:  map[string]any{"reason": "invited"},
		UpdateSeq: 42,
		Action:    "joined",
	}

	ev := event.NewMemberEvent(member, contactID, event.MemberAdded)
	m := New()

	got, err := m.Marshal(ev, contactID)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	// The result should be JSON bytes
	data, ok := got.([]byte)
	if !ok {
		t.Fatalf("expected []byte, got %T", got)
	}

	// Parse it back to verify structure
	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	// Check that member_added_event payload exists
	payload, ok := result["payload"].(map[string]any)
	if !ok {
		t.Fatalf("no payload: %v", result)
	}

	memberAdded, ok := payload["member_added_event"].(map[string]any)
	if !ok {
		t.Fatalf("no member_added_event in payload: %v", payload)
	}

	// Verify UpdateSeq is present and correct
	updateSeq, ok := memberAdded["update_seq"]
	if !ok {
		t.Errorf("no update_seq field in member_added_event: %v", memberAdded)
	}

	if updateSeq != float64(42) { // JSON numbers are float64
		t.Errorf("update_seq = %v, want 42", updateSeq)
	}

	// Verify action field
	action, ok := memberAdded["action"]
	if !ok {
		t.Errorf("no action field in member_added_event: %v", memberAdded)
	}

	if action != "joined" {
		t.Errorf("action = %v, want 'joined'", action)
	}
}

// TestMarshalMemberLeftWithUpdateSeq verifies that a member left event
// with UpdateSeq is correctly marshaled to WebSocket JSON.
func TestMarshalMemberLeftWithUpdateSeq(t *testing.T) {
	threadID := uuid.New()
	contactID := uuid.New()

	member := &model.MemberEvent{
		ThreadID:  threadID,
		ContactID: contactID,
		Metadata:  map[string]any{},
		UpdateSeq: 99,
		Action:    "left",
	}

	ev := event.NewMemberEvent(member, contactID, event.MemberLeft)
	m := New()

	got, err := m.Marshal(ev, contactID)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	data, ok := got.([]byte)
	if !ok {
		t.Fatalf("expected []byte, got %T", got)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	payload, ok := result["payload"].(map[string]any)
	if !ok {
		t.Fatalf("no payload: %v", result)
	}

	memberLeft, ok := payload["member_left_event"].(map[string]any)
	if !ok {
		t.Fatalf("no member_left_event in payload: %v", payload)
	}

	// Verify UpdateSeq is present and correct
	updateSeq, ok := memberLeft["update_seq"]
	if !ok {
		t.Errorf("no update_seq field in member_left_event: %v", memberLeft)
	}

	if updateSeq != float64(99) {
		t.Errorf("update_seq = %v, want 99", updateSeq)
	}

	// Verify action field
	action, ok := memberLeft["action"]
	if !ok {
		t.Errorf("no action field in member_left_event: %v", memberLeft)
	}

	if action != "left" {
		t.Errorf("action = %v, want 'left'", action)
	}
}
