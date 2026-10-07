package amqp

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/webitel/im-delivery-service/internal/domain/model"
	"github.com/webitel/im-delivery-service/internal/handler/amqp/payload"
)

// The status member is enriched like a message sender: identity from contacts, membership from the event.
func TestOnMessageStatus_EnrichesMember(t *testing.T) {
	reader := uuid.MustParse(customerID)
	h := newHandler([]model.Peer{{ID: reader, Sub: "3", Issuer: "webitel", Name: "Admin"}})

	events, err := h.OnMessageStatusV1(context.Background(), &payload.MessageStatusV1{
		ThreadID:     uuid.NewString(),
		MemberID:     reader.String(),
		Member:       &payload.Peer{ContactID: reader.String(), MemberID: "m-1", Role: 3},
		Status:       "read",
		Participants: []string{reader.String()},
	})
	if err != nil || len(events) == 0 {
		t.Fatalf("events = %d, err = %v", len(events), err)
	}

	got := events[0].GetPayload().(*model.MessageStatusUpdate).Member
	if got == nil || got.Name != "Admin" || got.MemberID != "m-1" || got.Role != 3 {
		t.Fatalf("member = %+v, want enriched contact with membership", got)
	}
}

// The reader's unread count reaches only the reader; other participants get the status without it.
func TestOnMessageStatus_UnreadCountOnlyForReader(t *testing.T) {
	reader := uuid.MustParse(customerID)
	other := uuid.New()
	h := newHandler(nil)

	unread := int64(6)
	events, err := h.OnMessageStatusV1(context.Background(), &payload.MessageStatusV1{
		ThreadID:     uuid.NewString(),
		MemberID:     reader.String(),
		Status:       "read",
		UnreadCount:  &unread,
		Participants: []string{reader.String(), other.String()},
	})
	if err != nil || len(events) != 2 {
		t.Fatalf("events = %d, err = %v", len(events), err)
	}

	for _, ev := range events {
		got := ev.GetPayload().(*model.MessageStatusUpdate).UnreadCount
		switch ev.GetUserID() {
		case reader:
			if got == nil || *got != unread {
				t.Fatalf("reader unread_count = %v, want %d", got, unread)
			}
		case other:
			if got != nil {
				t.Fatalf("participant unread_count = %d, want nil", *got)
			}
		default:
			t.Fatalf("unexpected target %s", ev.GetUserID())
		}
	}
}
