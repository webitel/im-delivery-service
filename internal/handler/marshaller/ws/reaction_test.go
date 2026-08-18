package wsmarshaller

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/webitel/im-delivery-service/internal/domain/model"
)

// TestMapReactionReactedByMePerViewer guards that reacted_by_me is resolved from
// the viewer's own contact id and that the raw reactor_ids never reach the wire.
func TestMapReactionReactedByMePerViewer(t *testing.T) {
	me := uuid.New()
	other := uuid.New()

	reaction := &model.MessageReaction{
		ID:       uuid.New(),
		ThreadID: uuid.New(),
		Reactor:  model.Peer{ID: other, MemberID: uuid.NewString(), Role: 1},
		Emoji:    "👀",
		Reactions: []model.ReactionAggregate{
			{Emoji: "👀", Count: 2, ReactorIDs: []string{me.String(), other.String()}, LastReactedAt: 1},
			{Emoji: "😭", Count: 1, ReactorIDs: []string{other.String()}, LastReactedAt: 2},
		},
	}

	got := mapReaction(reaction, me)

	if len(got.Reactions) != 2 {
		t.Fatalf("want 2 aggregates, got %d", len(got.Reactions))
	}

	if !got.Reactions[0].ReactedByMe {
		t.Errorf("👀: viewer reacted, want reacted_by_me=true")
	}

	if got.Reactions[1].ReactedByMe {
		t.Errorf("😭: viewer did not react, want reacted_by_me=false")
	}

	// reactor_ids must not leak into the client payload.
	raw, err := json.Marshal(got)
	if err != nil {
		t.Fatalf("marshal reaction: %v", err)
	}

	if strings.Contains(string(raw), "reactor_ids") {
		t.Errorf("reactor_ids leaked into payload: %s", raw)
	}
}
