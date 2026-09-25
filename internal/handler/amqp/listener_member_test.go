package amqp

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/webitel/im-delivery-service/internal/domain/model"
	"github.com/webitel/im-delivery-service/internal/handler/amqp/payload"
)

// Every participant (not only the joiner) must get the event, or their update_seq gaps.
func TestOnMemberAdded_FansOutToAllParticipants(t *testing.T) {
	h := newHandler(nil)
	joiner := uuid.MustParse(customerID)
	other := uuid.MustParse(ownerBotID)

	events, err := h.OnMemberAddedV1(context.Background(), &payload.MemberEventV1{
		ThreadID:     uuid.NewString(),
		ContactID:    joiner.String(),
		UpdateSeq:    9,
		Participants: []string{other.String(), joiner.String(), "not-a-uuid"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := targetSet(t, events)
	if len(got) != 2 || !got[joiner] || !got[other] {
		t.Fatalf("targets = %v, want joiner and other once each", got)
	}

	for _, e := range events {
		m, ok := e.GetPayload().(*model.MemberEvent)
		if !ok || m.UpdateSeq != 9 || m.Action != "joined" {
			t.Fatalf("payload = %+v, want update_seq 9 action joined", e.GetPayload())
		}
	}
}

func TestOnMemberLeft_WithoutParticipantsTargetsSubject(t *testing.T) {
	h := newHandler(nil)
	leaver := uuid.MustParse(customerID)

	events, err := h.OnMemberLeftV1(context.Background(), &payload.MemberEventV1{
		ThreadID:  uuid.NewString(),
		ContactID: leaver.String(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := targetSet(t, events); len(got) != 1 || !got[leaver] {
		t.Fatalf("targets = %v, want only the leaver", got)
	}
}
