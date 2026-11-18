package amqp

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/webitel/im-delivery-service/internal/domain/model"
	"github.com/webitel/im-delivery-service/internal/handler/amqp/payload"
)

func TestOnInteractiveCallbackReactedV1_KeepsMembershipAfterEnrichment(t *testing.T) {
	reactor, receiver := uuid.New(), uuid.New()
	reactorMember, receiverMember := uuid.NewString(), uuid.NewString()

	h := newHandler([]model.Peer{
		{ID: reactor, Type: model.PeerUser, Sub: "sub-reactor", Issuer: "https://issuer.example.com", Name: "Reactor", ContactType: "portal"},
		// The contact says "not a bot", but the thread membership does: membership wins.
		{ID: receiver, Type: model.PeerUser, Sub: "sub-bot", Issuer: "bot", Name: "Bot", ContactType: "bot"},
	})

	evs, err := h.OnInteractiveCallbackReactedV1(context.Background(), &payload.InteractiveCallbackV1{
		ReactedBy:       payload.ReactedBy{ID: reactor.String(), Type: 1},
		Receiver:        receiver.String(),
		InReplyTo:       uuid.NewString(),
		ThreadID:        uuid.NewString(),
		ButtonCode:      "btn_code",
		CallbackData:    "btn_data",
		ReactedAt:       "2026-11-17T16:40:55.713468+02:00",
		DomainID:        1,
		ReactedByMember: &payload.Recipient{MemberID: reactorMember, ContactID: reactor.String(), Role: 1},
		ReceiverMember:  &payload.Recipient{MemberID: receiverMember, ContactID: receiver.String(), Role: 2, IsBot: true},
	})
	if err != nil {
		t.Fatalf("handle: %v", err)
	}

	if len(evs) != 1 {
		t.Fatalf("events = %d, want 1", len(evs))
	}

	got, ok := evs[0].GetPayload().(*model.InteractiveCallback)
	if !ok {
		t.Fatalf("payload = %T, want *model.InteractiveCallback", evs[0].GetPayload())
	}

	if got.ReactedBy.MemberID != reactorMember || got.ReactedBy.Role != 1 || got.ReactedBy.Sub != "sub-reactor" {
		t.Errorf("reacted_by = %+v, want member %s role 1 enriched sub", got.ReactedBy, reactorMember)
	}

	if got.Receiver.MemberID != receiverMember || got.Receiver.Role != 2 || !got.Receiver.IsBot || got.Receiver.ContactType != "bot" {
		t.Errorf("receiver = %+v, want member %s role 2 enriched bot", got.Receiver, receiverMember)
	}

	if evs[0].GetUserID() != reactor {
		t.Errorf("target = %s, want reactor %s", evs[0].GetUserID(), reactor)
	}
}

// A legacy event (no membership blocks) must keep the contact-level is_bot.
func TestOnInteractiveCallbackReactedV1_LegacyEventKeepsContactIsBot(t *testing.T) {
	reactor, receiver := uuid.New(), uuid.New()

	h := newHandler([]model.Peer{
		{ID: reactor, Type: model.PeerUser, Sub: "sub-reactor"},
		{ID: receiver, Type: model.PeerUser, Sub: "sub-bot", IsBot: true},
	})

	evs, err := h.OnInteractiveCallbackReactedV1(context.Background(), &payload.InteractiveCallbackV1{
		ReactedBy: payload.ReactedBy{ID: reactor.String(), Type: 1},
		Receiver:  receiver.String(),
	})
	if err != nil {
		t.Fatalf("handle: %v", err)
	}

	got := evs[0].GetPayload().(*model.InteractiveCallback)
	if !got.Receiver.IsBot || got.Receiver.MemberID != "" {
		t.Errorf("receiver = %+v, want contact is_bot kept and no member id", got.Receiver)
	}
}

func TestInteractiveCallbackV1_ToDomainMembership(t *testing.T) {
	reactor, receiver := uuid.New(), uuid.New()

	tests := []struct {
		name           string
		reactedBy      *payload.Recipient
		receiver       *payload.Recipient
		wantReactorMID string
		wantReceiverID string
	}{
		{
			name:           "memberships matched by contact id",
			reactedBy:      &payload.Recipient{MemberID: "member-reactor", ContactID: reactor.String(), Role: 1},
			receiver:       &payload.Recipient{MemberID: "member-receiver", ContactID: receiver.String(), Role: 2},
			wantReactorMID: "member-reactor",
			wantReceiverID: "member-receiver",
		},
		{
			name: "legacy event without memberships",
		},
		{
			name:      "membership of another contact is ignored",
			reactedBy: &payload.Recipient{MemberID: "member-foreign", ContactID: uuid.NewString(), Role: 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := (&payload.InteractiveCallbackV1{
				ReactedBy:       payload.ReactedBy{ID: reactor.String()},
				Receiver:        receiver.String(),
				ReactedByMember: tt.reactedBy,
				ReceiverMember:  tt.receiver,
			}).ToDomain()
			if err != nil {
				t.Fatalf("to domain: %v", err)
			}

			if got.ReactedBy.MemberID != tt.wantReactorMID {
				t.Errorf("reacted_by member = %q, want %q", got.ReactedBy.MemberID, tt.wantReactorMID)
			}

			if got.Receiver.MemberID != tt.wantReceiverID {
				t.Errorf("receiver member = %q, want %q", got.Receiver.MemberID, tt.wantReceiverID)
			}
		})
	}
}
