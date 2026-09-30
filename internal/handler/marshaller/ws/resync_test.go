package wsmarshaller

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/webitel/im-delivery-service/internal/domain/event"
	"github.com/webitel/im-delivery-service/internal/domain/model"
)

func TestMarshal_Resync(t *testing.T) {
	ev := event.NewSystemEvent(uuid.New(), event.Resync, &model.ResyncPayload{})

	data, err := New().Marshal(ev, uuid.Nil)
	if err != nil {
		t.Fatal(err)
	}

	var out struct {
		Payload map[string]json.RawMessage `json:"payload"`
	}
	if err := json.Unmarshal(data.([]byte), &out); err != nil {
		t.Fatal(err)
	}

	if _, ok := out.Payload["resync_event"]; !ok {
		t.Fatalf("payload = %s, want resync_event", data)
	}
}

// The WS message carries seq and updates_cursor next to the content.
func TestMapMessage_SeqAndCursor(t *testing.T) {
	data, err := json.Marshal(mapMessage(&model.Message{ID: uuid.New(), Seq: 7, UpdatesCursor: "99.1"}))
	if err != nil {
		t.Fatal(err)
	}

	var out struct {
		Seq           int64  `json:"seq"`
		UpdatesCursor string `json:"updates_cursor"`
	}
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatal(err)
	}

	if out.Seq != 7 || out.UpdatesCursor != "99.1" {
		t.Fatalf("json = %s", data)
	}
}

// A status event carries the member as an object and the failure as {code, message}.
func TestMapMessageStatus_MemberAndError(t *testing.T) {
	data, err := json.Marshal(mapMessageStatus(&model.MessageStatusUpdate{
		ThreadID: uuid.New(), Status: "failed", MessageIDs: []uuid.UUID{uuid.New()},
		Member: &model.Peer{MemberID: "m-1", Role: 3, Sub: "3", Name: "Admin"},
		Error:  map[string]any{"code": "recipient_blocked", "message": "Recipient is unavailable"},
	}))
	if err != nil {
		t.Fatal(err)
	}

	var out struct {
		Member struct {
			ID      string `json:"id"`
			Contact struct {
				Name string `json:"name"`
			} `json:"contact"`
		} `json:"member"`
		MemberID *string `json:"member_id"`
		Error    struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatal(err)
	}

	if out.Member.ID != "m-1" || out.Member.Contact.Name != "Admin" || out.MemberID != nil ||
		out.Error.Code != "recipient_blocked" || out.Error.Message != "Recipient is unavailable" {
		t.Fatalf("json = %s", data)
	}
}
