package wsmarshaller

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/webitel/im-delivery-service/internal/domain/event"
	"github.com/webitel/im-delivery-service/internal/domain/model"
)

// TestMarshal_InteractiveCallbackEvent pins the client-facing JSON: the event
// key, the field names and the shared WSPeer shape of reacted_by and receiver.
func TestMarshal_InteractiveCallbackEvent(t *testing.T) {
	threadID, messageID := uuid.NewString(), uuid.NewString()
	reactorMember, receiverMember := uuid.NewString(), uuid.NewString()

	cb := &model.InteractiveCallback{
		ReactedBy: model.Peer{
			ID: uuid.New(), Type: model.PeerUser, MemberID: reactorMember, Role: int32(model.RoleMember),
			Sub: "sub-reactor", Issuer: "https://issuer.example.com", Name: "Reactor", ContactType: "portal",
		},
		Receiver: model.Peer{
			ID: uuid.New(), Type: model.PeerUser, MemberID: receiverMember, Role: int32(model.RoleAdmin),
			Sub: "sub-bot", Issuer: "bot", Name: "Bot", Username: "bot_user", ContactType: "bot", IsBot: true,
		},
		InReplyTo:     messageID,
		ThreadID:      threadID,
		ButtonCode:    "btn_code",
		CallbackData:  "btn_data",
		ReactedAt:     "2026-11-17T16:40:55.713468+02:00",
		DomainID:      1,
		UpdatesCursor: "42",
	}

	data, err := New().Marshal(event.NewInteractiveCallbackEvent(cb), uuid.Nil)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var out struct {
		Payload map[string]json.RawMessage `json:"payload"`
	}
	if err := json.Unmarshal(data.([]byte), &out); err != nil {
		t.Fatalf("unmarshal %s: %v", data, err)
	}

	raw, ok := out.Payload["interactive_callback_event"]
	if !ok || len(out.Payload) != 1 {
		t.Fatalf("payload = %s, want only interactive_callback_event", data)
	}

	var got map[string]any
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("unmarshal event: %v", err)
	}

	want := map[string]any{
		"thread_id":      threadID,
		"message_id":     messageID,
		"button_code":    "btn_code",
		"callback_data":  "btn_data",
		"reacted_at":     float64(1794926455713),
		"updates_cursor": "42",
		"reacted_by": map[string]any{
			"id":   reactorMember,
			"role": "ROLE_MEMBER",
			"contact": map[string]any{
				"sub": "sub-reactor", "iss": "https://issuer.example.com", "name": "Reactor",
				"username": "Reactor", "type": "portal", "is_bot": false,
			},
		},
		"receiver": map[string]any{
			"id":   receiverMember,
			"role": "ROLE_ADMIN",
			"contact": map[string]any{
				"sub": "sub-bot", "iss": "bot", "name": "Bot",
				"username": "bot_user", "type": "bot", "is_bot": true,
			},
		},
	}

	gotJSON, _ := json.Marshal(got)
	wantJSON, _ := json.Marshal(want)

	if string(gotJSON) != string(wantJSON) {
		t.Fatalf("interactive_callback_event =\n%s\nwant\n%s", gotJSON, wantJSON)
	}
}
