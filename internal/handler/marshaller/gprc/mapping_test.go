package grpcmarshaller

import (
	"testing"

	"github.com/google/uuid"

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
