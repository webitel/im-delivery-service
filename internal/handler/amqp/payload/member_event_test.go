package payload

import (
	"testing"

	"github.com/google/uuid"
)

// TestMemberEventV1UpdateSeq verifies that update_seq from the AMQP payload
// is correctly mapped to the domain model.
func TestMemberEventV1UpdateSeq(t *testing.T) {
	threadID := uuid.New()
	contactID := uuid.New()

	in := &MemberEventV1{
		ThreadID:   threadID.String(),
		DomainID:   1,
		ContactID:  contactID.String(),
		OccurredAt: "2024-01-01T00:00:00Z",
		System: MemberEventSystem{
			Type:     "member_joined",
			Metadata: map[string]any{"reason": "invited"},
		},
		UpdateSeq: 42,
	}

	got := in.ToDomain()

	if got.ThreadID != threadID {
		t.Errorf("ThreadID = %v, want %v", got.ThreadID, threadID)
	}

	if got.ContactID != contactID {
		t.Errorf("ContactID = %v, want %v", got.ContactID, contactID)
	}

	if got.UpdateSeq != 42 {
		t.Errorf("UpdateSeq = %d, want 42", got.UpdateSeq)
	}

	if got.Metadata["reason"] != "invited" {
		t.Errorf("Metadata not preserved")
	}
}
