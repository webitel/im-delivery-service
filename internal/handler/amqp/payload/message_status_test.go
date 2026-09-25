package payload

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestMessageStatusV1_ToDomain(t *testing.T) {
	threadID := uuid.New()
	memberID := uuid.New()
	msg1, msg2 := uuid.New(), uuid.New()

	in := &MessageStatusV1{
		ThreadID:   threadID.String(),
		MemberID:   memberID.String(),
		MessageIDs: []string{msg1.String(), "not-a-uuid", msg2.String()},
		Status:     "delivered",
		Via:        "ws",
		Error:      map[string]any{"code": "470"},
		OccurredAt: "2026-07-15T10:30:00Z",
	}

	got := in.ToDomain()

	if got.ThreadID != threadID || got.MemberID != memberID {
		t.Errorf("ids mismatch: %+v", got)
	}

	if len(got.MessageIDs) != 2 || got.MessageIDs[0] != msg1 || got.MessageIDs[1] != msg2 {
		t.Errorf("invalid message ids must be filtered, got %v", got.MessageIDs)
	}

	if got.Status != "delivered" || got.Via != "ws" || got.Error["code"] != "470" {
		t.Errorf("attrs mismatch: %+v", got)
	}

	want := time.Date(2026, 7, 15, 10, 30, 0, 0, time.UTC).UnixMilli()
	if got.OccurredAt != want {
		t.Errorf("occurred_at = %d, want %d", got.OccurredAt, want)
	}
}

func TestMessageStatusV1_ParticipantIDs(t *testing.T) {
	p1, p2 := uuid.New(), uuid.New()

	t.Run("dedups and filters invalid", func(t *testing.T) {
		in := &MessageStatusV1{
			Participants: []string{p1.String(), "garbage", p2.String(), p1.String()},
		}

		got := in.ParticipantIDs()

		if len(got) != 2 || got[0] != p1 || got[1] != p2 {
			t.Errorf("expected [%s %s], got %v", p1, p2, got)
		}
	})

	t.Run("falls back to the affected member", func(t *testing.T) {
		member := uuid.New()
		in := &MessageStatusV1{MemberID: member.String()}

		got := in.ParticipantIDs()

		if len(got) != 1 || got[0] != member {
			t.Errorf("expected fallback to member, got %v", got)
		}
	})

	t.Run("no participants at all", func(t *testing.T) {
		in := &MessageStatusV1{MemberID: "broken", Participants: []string{"also-broken"}}

		if got := in.ParticipantIDs(); len(got) != 0 {
			t.Errorf("expected empty, got %v", got)
		}
	})
}

func TestMessageCreatedV1UpdateSeq(t *testing.T) {
	in := &MessageCreatedV1{
		MessageID:  uuid.New().String(),
		ThreadID:   uuid.New().String(),
		DomainID:   1,
		From:       Peer{ContactID: uuid.New().String(), MemberID: "m1", Role: 1},
		To:         []Recipient{},
		Body:       "test",
		Type:       "text",
		OccurredAt: "2024-01-01T00:00:00Z",
		SendID:     "s1",
		UpdateSeq:  42,
	}

	got := in.ToDomain()

	if got.UpdateSeq != 42 {
		t.Errorf("UpdateSeq = %d, want 42", got.UpdateSeq)
	}
}

func TestMessageEditedV1UpdateSeq(t *testing.T) {
	in := &MessageEditedV1{
		MessageID:  uuid.New().String(),
		ThreadID:   uuid.New().String(),
		DomainID:   1,
		EditedBy:   Peer{ContactID: uuid.New().String(), MemberID: "m1", Role: 1},
		To:         []Recipient{},
		Body:       "edited",
		Type:       "text",
		Version:    2,
		CreatedAt:  "2024-01-01T00:00:00Z",
		OccurredAt: "2024-01-01T00:00:01Z",
		UpdateSeq:  100,
	}

	got := in.ToDomain()

	if got.UpdateSeq != 100 {
		t.Errorf("UpdateSeq = %d, want 100", got.UpdateSeq)
	}
}

func TestMessageDeletedV1UpdateSeq(t *testing.T) {
	in := &MessageDeletedV1{
		MessageID:  uuid.New().String(),
		ThreadID:   uuid.New().String(),
		DomainID:   1,
		DeletedBy:  Member{ID: "m1", Contact: nil, Role: "admin"},
		To:         []Recipient{},
		Type:       "text",
		CreatedAt:  "2024-01-01T00:00:00Z",
		OccurredAt: "2024-01-01T00:00:01Z",
		UpdateSeq:  50,
	}

	got := in.ToDomain()

	if got.UpdateSeq != 50 {
		t.Errorf("UpdateSeq = %d, want 50", got.UpdateSeq)
	}
}

func TestMessageReactionV1UpdateSeq(t *testing.T) {
	in := &MessageReactionV1{
		MessageID:  uuid.New().String(),
		ThreadID:   uuid.New().String(),
		DomainID:   1,
		Reactor:    Peer{ContactID: uuid.New().String(), MemberID: "m1", Role: 1},
		To:         []Recipient{},
		Emoji:      "🔥",
		Action:     "added",
		OccurredAt: "2024-01-01T00:00:00Z",
		SendId:     "s1",
		UpdateSeq:  75,
	}

	got := in.ToDomain()

	if got.UpdateSeq != 75 {
		t.Errorf("UpdateSeq = %d, want 75", got.UpdateSeq)
	}
}
