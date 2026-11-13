package grpcmarshaller

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"google.golang.org/protobuf/encoding/protojson"

	impb "github.com/webitel/im-delivery-service/gen/go/delivery/v1"
	"github.com/webitel/im-delivery-service/internal/domain/event"
	"github.com/webitel/im-delivery-service/internal/domain/model"
)

// TestMarshalMemberChangedPayload verifies MemberEvent maps onto gRPC MemberChangedEvent.
func TestMarshalMemberChangedPayload(t *testing.T) {
	threadID := uuid.New()
	contactID := uuid.New()

	tests := []struct {
		name   string
		action string
	}{
		{"member joined", "joined"},
		{"member left", "left"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			member := &model.MemberEvent{
				ThreadID:  threadID,
				ContactID: contactID,
				Action:    tt.action,
			}

			got := marshalMemberChangedPayload(member)

			if got.MemberChangedEvent.GetThreadId() != threadID.String() {
				t.Errorf("ThreadId = %s, want %s", got.MemberChangedEvent.GetThreadId(), threadID.String())
			}

			if got.MemberChangedEvent.GetContactId() != contactID.String() {
				t.Errorf("ContactId = %s, want %s", got.MemberChangedEvent.GetContactId(), contactID.String())
			}

			if got.MemberChangedEvent.GetAction() != tt.action {
				t.Errorf("Action = %s, want %s", got.MemberChangedEvent.GetAction(), tt.action)
			}
		})
	}
}

// Delivered/read carry only the horizon; failed names the message it is about.
func TestMarshalMessageStatusPayload_MessageIDOnlyForFailures(t *testing.T) {
	msg := uuid.New()

	read := marshalMessageStatusPayload(&model.MessageStatusUpdate{Status: "read", UpToSeq: 1, MessageIDs: []uuid.UUID{msg}}).MessageStatusEvent
	if read.GetMessageId() != "" || read.GetUpToSeq() != 1 {
		t.Fatalf("read: message_id=%q up_to_seq=%d, want empty and 1", read.GetMessageId(), read.GetUpToSeq())
	}

	failed := marshalMessageStatusPayload(&model.MessageStatusUpdate{Status: "failed", MessageIDs: []uuid.UUID{msg}}).MessageStatusEvent
	if failed.GetMessageId() != msg.String() {
		t.Fatalf("failed: message_id=%q, want %s", failed.GetMessageId(), msg)
	}
}

func TestMarshalMessageDeletedPayload_ReplacedBy(t *testing.T) {
	replacement := uuid.New()

	tests := []struct {
		name       string
		replacedBy *uuid.UUID
		want       string
	}{
		{"plain delete", nil, ""},
		{"replaced by retry", &replacement, replacement.String()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := marshalMessageDeletedPayload(&model.MessageDeleted{
				ID:         uuid.New(),
				ThreadID:   uuid.New(),
				ReplacedBy: tt.replacedBy,
			}).MessageDeletedEvent.GetReplacedBy()

			if got != tt.want {
				t.Errorf("replaced_by = %q, want %q", got, tt.want)
			}
		})
	}
}

// Through the full Marshal path a retry delete carries replaced_by and an
// ordinary delete leaves it empty, which proto3 never puts on the wire.
func TestMarshal_MessageDeleted_ReplacedBy(t *testing.T) {
	replacement := uuid.New()

	tests := []struct {
		name       string
		replacedBy *uuid.UUID
		want       string
	}{
		{"plain delete", nil, ""},
		{"replaced by retry", &replacement, replacement.String()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ev := event.NewMessageDeletedEvent(&model.MessageDeleted{
				ID: uuid.New(), ThreadID: uuid.New(), ReplacedBy: tt.replacedBy,
			}, uuid.New())

			out, err := New().Marshal(ev, uuid.Nil)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}

			deleted := out.(*impb.ServerEvent).GetMessageDeletedEvent()
			if deleted == nil {
				t.Fatalf("event = %v, want message_deleted_event", out)
			}

			if got := deleted.GetReplacedBy(); got != tt.want {
				t.Errorf("replaced_by = %q, want %q", got, tt.want)
			}

			js, err := protojson.Marshal(deleted)
			if err != nil {
				t.Fatalf("protojson: %v", err)
			}

			if has := strings.Contains(string(js), "replacedBy"); has != (tt.want != "") {
				t.Errorf("json = %s, replacedBy present = %v, want %v", js, has, tt.want != "")
			}
		})
	}
}
