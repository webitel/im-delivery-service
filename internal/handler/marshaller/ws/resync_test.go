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
