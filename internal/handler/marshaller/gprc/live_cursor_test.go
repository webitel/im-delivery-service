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

func TestMarshal_ConnectedCursor(t *testing.T) {
	ev := event.NewSystemEvent(uuid.New(), event.Connected, &model.ConnectedPayload{Ok: true, UpdatesCursor: "94770179"})

	got, err := New().Marshal(ev, uuid.Nil)
	if err != nil {
		t.Fatal(err)
	}

	if c := got.(*impb.ServerEvent).GetConnectedEvent().GetUpdatesCursor(); c != "94770179" {
		t.Fatalf("connected cursor = %q, want 94770179", c)
	}
}

func TestMarshalMessageStatusPayload_Cursor(t *testing.T) {
	got := marshalMessageStatusPayload(&model.MessageStatusUpdate{Status: "read", UpToSeq: 4, UpdatesCursor: "94770180"}).MessageStatusEvent

	if got.GetUpdatesCursor() != "94770180" || got.GetUpToSeq() != 4 {
		t.Fatalf("status event = %+v", got)
	}
}

func TestMarshalMessageStatusPayload_MemberAndError(t *testing.T) {
	got := marshalMessageStatusPayload(&model.MessageStatusUpdate{
		Status: "failed", Member: &model.Peer{Type: model.PeerUser, Sub: "3", Issuer: "webitel", Name: "Admin"},
		Error: map[string]any{"code": "131047", "message": "re-engagement window expired"},
	}).MessageStatusEvent

	if got.GetMember().GetUserId() != "3" || got.GetError().GetCode() != "131047" || got.GetError().GetMessage() != "re-engagement window expired" {
		t.Fatalf("status event = %+v", got)
	}
}

func TestMarshalMessageStatusPayload_UnreadCount(t *testing.T) {
	two := int64(2)

	read := marshalMessageStatusPayload(&model.MessageStatusUpdate{Status: "read", UnreadCount: &two}).MessageStatusEvent
	if read.UnreadCount == nil || read.GetUnreadCount() != 2 {
		t.Fatalf("read unread_count = %v (set %v), want 2", read.GetUnreadCount(), read.UnreadCount != nil)
	}

	delivered := marshalMessageStatusPayload(&model.MessageStatusUpdate{Status: "delivered"}).MessageStatusEvent
	if delivered.UnreadCount != nil {
		t.Fatal("delivered must not carry unread_count")
	}
}
