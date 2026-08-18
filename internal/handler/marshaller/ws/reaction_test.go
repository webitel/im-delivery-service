package wsmarshaller

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/webitel/im-delivery-service/internal/domain/model"
)

// TestMapReactionReactedByMePerViewer guards that reacted_by_me is resolved from
// the viewer's own contact id and that the aggregate mirrors the im-thread
// history shape (nested reaction object).
func TestMapReactionReactedByMePerViewer(t *testing.T) {
	me := uuid.New()
	other := uuid.New()

	reaction := &model.MessageReaction{
		ID:       uuid.New(),
		ThreadID: uuid.New(),
		Reactor:  model.Peer{ID: other, MemberID: uuid.NewString(), Role: 1},
		Emoji:    "🔥",
		Reactions: []model.ReactionAggregate{
			{Emoji: "🔥", Count: 1, ReactorIDs: []string{other.String()}, LastReactedAt: 1},
			{Emoji: "😭", Count: 1, ReactorIDs: []string{me.String()}, LastReactedAt: 2},
		},
	}

	got := mapReaction(reaction, me)

	if got.MessageID != reaction.ID.String() {
		t.Errorf("message_id = %q, want %q", got.MessageID, reaction.ID.String())
	}

	if len(got.Reactions) != 2 {
		t.Fatalf("want 2 aggregates, got %d", len(got.Reactions))
	}

	if got.Reactions[0].Emoji != "🔥" {
		t.Errorf("reaction[0].emoji = %q, want 🔥", got.Reactions[0].Emoji)
	}

	// 🔥 was added by the other member, not the viewer.
	if got.Reactions[0].ReactedByMe {
		t.Errorf("🔥: viewer did not react, want reacted_by_me=false")
	}

	// 😭 is held by the viewer.
	if !got.Reactions[1].ReactedByMe {
		t.Errorf("😭: viewer reacted, want reacted_by_me=true")
	}
}

// TestMapReactionAlwaysEmitsReactionsArray guards that clearing the last
// reaction still sends an empty (non-null) reactions array so the client can
// empty the bar.
func TestMapReactionAlwaysEmitsReactionsArray(t *testing.T) {
	reaction := &model.MessageReaction{
		ID:       uuid.New(),
		ThreadID: uuid.New(),
		Reactor:  model.Peer{ID: uuid.New(), MemberID: uuid.NewString(), Role: 1},
		Emoji:    "🔥",
		Removed:  true,
	}

	raw, err := json.Marshal(mapReaction(reaction, uuid.New()))
	if err != nil {
		t.Fatalf("marshal reaction: %v", err)
	}

	if !strings.Contains(string(raw), `"reactions":[]`) {
		t.Errorf("want empty reactions array, got: %s", raw)
	}
}
