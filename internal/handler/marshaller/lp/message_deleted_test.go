package lpmarshaller

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/webitel/im-delivery-service/internal/domain/event"
	"github.com/webitel/im-delivery-service/internal/domain/model"
)

// deletedEventJSON marshals a deletion through the long-poll batch and returns
// its payload object exactly as a client receives it.
func deletedEventJSON(t *testing.T, m *model.MessageDeleted) map[string]json.RawMessage {
	t.Helper()

	data, err := New().MarshalBatch([]event.Eventer{event.NewMessageDeletedEvent(m, uuid.New())}, uuid.Nil)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var out struct {
		Events []struct {
			Type    string                     `json:"type"`
			Payload map[string]json.RawMessage `json:"payload"`
		} `json:"events"`
	}
	if err := json.Unmarshal(data.([]byte), &out); err != nil {
		t.Fatalf("unmarshal %s: %v", data, err)
	}

	if len(out.Events) != 1 || out.Events[0].Type != "message_deleted" {
		t.Fatalf("batch = %s, want one message_deleted event", data)
	}

	return out.Events[0].Payload
}

func TestMarshalBatch_MessageDeleted_ReplacedByInJSON(t *testing.T) {
	replacement := uuid.New()

	deleted := deletedEventJSON(t, &model.MessageDeleted{ID: uuid.New(), ThreadID: uuid.New(), ReplacedBy: &replacement})

	var got string
	if err := json.Unmarshal(deleted["replaced_by"], &got); err != nil || got != replacement.String() {
		t.Fatalf("replaced_by = %s (err %v), want %q", deleted["replaced_by"], err, replacement)
	}
}

func TestMarshalBatch_MessageDeleted_PlainDeleteOmitsReplacedBy(t *testing.T) {
	deleted := deletedEventJSON(t, &model.MessageDeleted{ID: uuid.New(), ThreadID: uuid.New()})

	if raw, ok := deleted["replaced_by"]; ok {
		t.Fatalf("replaced_by = %s, want key absent", raw)
	}

	if _, ok := deleted["id"]; !ok {
		t.Fatalf("payload = %v, want id present", deleted)
	}
}
