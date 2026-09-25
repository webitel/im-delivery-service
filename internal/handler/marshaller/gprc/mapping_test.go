package grpcmarshaller

import (
	"encoding/json"
	"strings"
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

// thread-service sends only up_to_seq; a missing message id must stay empty, not zeros.
func TestMarshalMessageStatusPayload_OmitsMissingUpToMessageID(t *testing.T) {
	got := marshalMessageStatusPayload(&model.MessageStatusUpdate{Status: "read", UpToSeq: 1}).MessageStatusEvent

	if got.GetUpToMessageId() != "" || got.GetUpToSeq() != 1 {
		t.Fatalf("up_to_message_id=%q up_to_seq=%d, want empty and 1", got.GetUpToMessageId(), got.GetUpToSeq())
	}

	raw, err := json.Marshal(&model.MessageStatusUpdate{Status: "read", UpToSeq: 1})
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(string(raw), "up_to_message_id") {
		t.Fatalf("JSON must omit up_to_message_id: %s", raw)
	}
}
