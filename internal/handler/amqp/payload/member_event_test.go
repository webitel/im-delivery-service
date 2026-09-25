package payload

import (
	"testing"

	"github.com/google/uuid"
)

// TestMemberEventV1ToDomain verifies the AMQP member payload maps onto the domain model.
func TestMemberEventV1ToDomain(t *testing.T) {
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
	}

	got := in.ToDomain()

	if got.ThreadID != threadID {
		t.Errorf("ThreadID = %v, want %v", got.ThreadID, threadID)
	}

	if got.ContactID != contactID {
		t.Errorf("ContactID = %v, want %v", got.ContactID, contactID)
	}

	if got.Metadata["reason"] != "invited" {
		t.Errorf("Metadata not preserved")
	}
}
