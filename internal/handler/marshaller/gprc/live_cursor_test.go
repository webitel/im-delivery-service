package grpcmarshaller

import (
	"testing"

	"github.com/google/uuid"

	impb "github.com/webitel/im-delivery-service/gen/go/delivery/v1"
	"github.com/webitel/im-delivery-service/internal/domain/event"
	"github.com/webitel/im-delivery-service/internal/domain/model"
)

// A new message carries its seq (for status ticks) and the GetUpdates cursor.
func TestMarshalMessagePayload_SeqAndCursor(t *testing.T) {
	got := marshalMessagePayload(&model.Message{ID: uuid.New(), Seq: 7, UpdatesCursor: "99.1790000000000"}).MessageEvent

	if got.GetMessage().GetSeq() != 7 || got.GetUpdatesCursor() != "99.1790000000000" {
		t.Fatalf("event = %+v, want seq 7 and the cursor", got)
	}
}

func TestMarshal_ResyncCursor(t *testing.T) {
	ev := event.NewSystemEvent(uuid.New(), event.Resync, &model.ResyncPayload{Cursor: "5.6"})

	got, err := New().Marshal(ev, uuid.Nil)
	if err != nil {
		t.Fatal(err)
	}

	if c := got.(*impb.ServerEvent).GetResyncEvent().GetCursor(); c != "5.6" {
		t.Fatalf("resync cursor = %q, want 5.6", c)
	}
}
