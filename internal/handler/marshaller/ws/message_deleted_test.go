package wsmarshaller

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/webitel/im-delivery-service/internal/domain/event"
	"github.com/webitel/im-delivery-service/internal/domain/model"
)

func TestMapMessageDeleted_ReplacedBy(t *testing.T) {
	replacement := uuid.New()

	if got := mapMessageDeleted(&model.MessageDeleted{ID: uuid.New()}).ReplacedBy; got != "" {
		t.Errorf("plain delete: replaced_by = %q, want empty", got)
	}

	if got := mapMessageDeleted(&model.MessageDeleted{ID: uuid.New(), ReplacedBy: &replacement}).ReplacedBy; got != replacement.String() {
		t.Errorf("retry: replaced_by = %q, want %q", got, replacement.String())
	}
}

// deletedEventJSON marshals a deletion through the WS marshaller and returns
// the message_deleted_event object exactly as a client receives it.
func deletedEventJSON(t *testing.T, m *model.MessageDeleted) map[string]json.RawMessage {
	t.Helper()

	data, err := New().Marshal(event.NewMessageDeletedEvent(m, uuid.New()), uuid.Nil)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var out struct {
		Payload map[string]map[string]json.RawMessage `json:"payload"`
	}
	if err := json.Unmarshal(data.([]byte), &out); err != nil {
		t.Fatalf("unmarshal %s: %v", data, err)
	}

	deleted, ok := out.Payload[EventMessageDeleted.String()]
	if !ok {
		t.Fatalf("payload = %s, want %s", data, EventMessageDeleted)
	}

	return deleted
}

func TestMarshal_MessageDeleted_ReplacedByInJSON(t *testing.T) {
	replacement := uuid.New()

	deleted := deletedEventJSON(t, &model.MessageDeleted{ID: uuid.New(), ThreadID: uuid.New(), ReplacedBy: &replacement})

	var got string
	if err := json.Unmarshal(deleted["replaced_by"], &got); err != nil || got != replacement.String() {
		t.Fatalf("replaced_by = %s (err %v), want %q", deleted["replaced_by"], err, replacement)
	}
}

// An ordinary delete must not carry the key at all: clients treat its presence
// as "drop the bubble" instead of showing a tombstone.
func TestMarshal_MessageDeleted_PlainDeleteOmitsReplacedBy(t *testing.T) {
	deleted := deletedEventJSON(t, &model.MessageDeleted{ID: uuid.New(), ThreadID: uuid.New()})

	if raw, ok := deleted["replaced_by"]; ok {
		t.Fatalf("replaced_by = %s, want key absent", raw)
	}

	if _, ok := deleted["id"]; !ok {
		t.Fatalf("event = %v, want id present", deleted)
	}
}
