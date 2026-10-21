package amqp

import (
	"context"
	"encoding/json"
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/webitel/im-delivery-service/internal/domain/event"
	"github.com/webitel/im-delivery-service/internal/domain/model"
	"github.com/webitel/im-delivery-service/internal/handler/amqp/payload"
)

// A tag change reaches only its owner, with the owner's current tags and cursor.
func TestOnThreadTagsChanged_OwnerOnly(t *testing.T) {
	owner, thread := uuid.MustParse(customerID), uuid.New()

	raw := []byte(`{"thread_id":"` + thread.String() + `","domain_id":7,"contact_id":"` + owner.String() +
		`","tags":["snoozed"],"updates_cursor":"42","participants":["` + owner.String() + `"]}`)

	var in payload.ThreadTagsChangedV1
	if err := json.Unmarshal(raw, &in); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	events, err := newHandler(nil).OnThreadTagsChangedV1(context.Background(), &in)
	if err != nil || len(events) != 1 {
		t.Fatalf("events = %d, err = %v", len(events), err)
	}

	ev := events[0]
	if ev.GetUserID() != owner || ev.GetKind() != event.ThreadTagsChanged {
		t.Fatalf("target = %s kind = %v, want owner %s", ev.GetUserID(), ev.GetKind(), owner)
	}

	got := ev.GetPayload().(*model.ThreadTagsChanged)
	if got.ThreadID != thread || !slices.Equal(got.Tags, []string{"snoozed"}) || got.UpdatesCursor != "42" {
		t.Fatalf("payload = %+v", got)
	}
}

// Removing the last tag still sends an empty list, so the client clears its copy.
func TestOnThreadTagsChanged_AllRemoved(t *testing.T) {
	owner := uuid.MustParse(customerID)

	events, err := newHandler(nil).OnThreadTagsChangedV1(context.Background(), &payload.ThreadTagsChangedV1{
		ThreadID: uuid.NewString(), ContactID: owner.String(),
	})
	if err != nil || len(events) != 1 {
		t.Fatalf("events = %d, err = %v", len(events), err)
	}

	tags := events[0].GetPayload().(*model.ThreadTagsChanged).Tags
	if tags == nil || len(tags) != 0 {
		t.Fatalf("tags = %#v, want empty non-nil", tags)
	}
}

func TestOnThreadTagsChanged_InvalidIDsIgnored(t *testing.T) {
	tests := []struct {
		name string
		in   *payload.ThreadTagsChangedV1
	}{
		{name: "no owner", in: &payload.ThreadTagsChangedV1{ThreadID: uuid.NewString()}},
		{name: "no thread", in: &payload.ThreadTagsChangedV1{ContactID: uuid.NewString()}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			events, err := newHandler(nil).OnThreadTagsChangedV1(context.Background(), tt.in)
			if err != nil || len(events) != 0 {
				t.Fatalf("events = %d, err = %v", len(events), err)
			}
		})
	}
}
