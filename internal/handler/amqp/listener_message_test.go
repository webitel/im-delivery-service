package amqp

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/google/uuid"

	"github.com/webitel/im-delivery-service/internal/domain/event"
	"github.com/webitel/im-delivery-service/internal/domain/model"
	"github.com/webitel/im-delivery-service/internal/domain/registry"
	"github.com/webitel/im-delivery-service/internal/handler/amqp/payload"
)

// --- test doubles -----------------------------------------------------------

type fakeEnricher struct{ peers []model.Peer }

func (f fakeEnricher) Resolve(_ context.Context, _ int32, _ ...uuid.UUID) ([]model.Peer, error) {
	return f.peers, nil
}

// leaderStub reports leadership so computeLocalTargets fans out to every
// participant, keeping the target set independent of the local hub.
type leaderStub struct{}

func (leaderStub) IsLeader() bool { return true }

type hubStub struct{}

func (hubStub) Broadcast(event.Eventer)     {}
func (hubStub) Register(registry.Connector) {}
func (hubStub) Unregister(_, _ uuid.UUID)   {}
func (hubStub) Connected(uuid.UUID) bool    { return false }
func (hubStub) Shutdown()                   {}

func strPtr(s string) *string { return &s }

// --- tests ------------------------------------------------------------------

const (
	customerID = "11111111-1111-1111-1111-111111111111"
	ownerBotID = "22222222-2222-2222-2222-222222222222"
	ctrlBotID  = "33333333-3333-3333-3333-333333333333"
)

func newHandler(peers []model.Peer) *MessageHandler {
	return &MessageHandler{
		hub:      hubStub{},
		logger:   slog.New(slog.NewTextHandler(io.Discard, nil)),
		enricher: fakeEnricher{peers: peers},
		leader:   leaderStub{},
	}
}

func targetSet(t *testing.T, events []event.Eventer) map[uuid.UUID]bool {
	t.Helper()

	set := make(map[uuid.UUID]bool, len(events))
	for _, e := range events {
		set[e.GetUserID()] = true
	}

	return set
}

// A customer message on a thread whose control stack has a bot ON TOP of the
// owner bot must reach only the active controller, never the suspended owner.
func TestOnMessageCreatedV1_SkipsNonControllerBot(t *testing.T) {
	peers := []model.Peer{
		{ID: uuid.MustParse(customerID), IsBot: false},
		{ID: uuid.MustParse(ownerBotID), IsBot: true},
		{ID: uuid.MustParse(ctrlBotID), IsBot: true},
	}

	raw := &payload.MessageCreatedV1{
		MessageID: uuid.NewString(),
		ThreadID:  uuid.NewString(),
		DomainID:  1,
		From:      payload.Peer{ContactID: customerID, MemberID: "c-mem"},
		To: []payload.Recipient{
			{ContactID: ownerBotID, MemberID: "ob-mem"},
			{ContactID: ctrlBotID, MemberID: "cb-mem"},
		},
		BotControllerMemberID: strPtr("cb-mem"),
	}

	events, err := newHandler(peers).OnMessageCreatedV1(context.Background(), raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	set := targetSet(t, events)

	if !set[uuid.MustParse(customerID)] {
		t.Errorf("customer (sender echo) must receive the message")
	}
	if !set[uuid.MustParse(ctrlBotID)] {
		t.Errorf("active controller bot must receive the message")
	}
	if set[uuid.MustParse(ownerBotID)] {
		t.Errorf("suspended owner bot must NOT be triggered by an inbound message")
	}
}

// When the owner bot IS the active controller (it is alone at the top of the
// stack), an inbound message must trigger it as before.
func TestOnMessageCreatedV1_OwnerAsControllerReceives(t *testing.T) {
	peers := []model.Peer{
		{ID: uuid.MustParse(customerID), IsBot: false},
		{ID: uuid.MustParse(ownerBotID), IsBot: true},
	}

	raw := &payload.MessageCreatedV1{
		MessageID: uuid.NewString(),
		ThreadID:  uuid.NewString(),
		DomainID:  1,
		From:      payload.Peer{ContactID: customerID, MemberID: "c-mem"},
		To: []payload.Recipient{
			{ContactID: ownerBotID, MemberID: "ob-mem"},
		},
		BotControllerMemberID: strPtr("ob-mem"),
	}

	events, err := newHandler(peers).OnMessageCreatedV1(context.Background(), raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !targetSet(t, events)[uuid.MustParse(ownerBotID)] {
		t.Errorf("owner bot acting as controller must receive the message")
	}
}

// With no controller advertised the fan-out is unchanged: every bot participant
// still receives the message (backwards-compatible behaviour).
func TestOnMessageCreatedV1_NoControllerKeepsAllBots(t *testing.T) {
	peers := []model.Peer{
		{ID: uuid.MustParse(customerID), IsBot: false},
		{ID: uuid.MustParse(ownerBotID), IsBot: true},
		{ID: uuid.MustParse(ctrlBotID), IsBot: true},
	}

	raw := &payload.MessageCreatedV1{
		MessageID: uuid.NewString(),
		ThreadID:  uuid.NewString(),
		DomainID:  1,
		From:      payload.Peer{ContactID: customerID, MemberID: "c-mem"},
		To: []payload.Recipient{
			{ContactID: ownerBotID, MemberID: "ob-mem"},
			{ContactID: ctrlBotID, MemberID: "cb-mem"},
		},
		BotControllerMemberID: nil,
	}

	events, err := newHandler(peers).OnMessageCreatedV1(context.Background(), raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	set := targetSet(t, events)
	if !set[uuid.MustParse(ownerBotID)] || !set[uuid.MustParse(ctrlBotID)] {
		t.Errorf("with no controller every bot must still receive the message")
	}
}
