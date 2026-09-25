package registry

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/webitel/im-delivery-service/internal/domain/event"
	"github.com/webitel/im-delivery-service/internal/domain/model"
)

func highEvent(kind event.EventKind) event.Eventer {
	return event.NewSystemEvent(uuid.New(), kind, "x", event.WithPriority[string](event.PriorityHigh))
}

// A dropped replayable event makes the next delivery start with a resync hint.
func TestSend_DroppedMessageQueuesResync(t *testing.T) {
	c := NewConnector(context.Background(), uuid.New(), 2, nil)
	defer c.Close()

	c.Send(highEvent(event.MessageCreated), time.Millisecond)
	c.Send(highEvent(event.MessageCreated), time.Millisecond)

	if c.Send(highEvent(event.MessageCreated), time.Millisecond) {
		t.Fatal("third event fit into a full buffer")
	}

	<-c.Recv()
	<-c.Recv()

	next := highEvent(event.MessageEdited)
	if !c.Send(next, time.Millisecond) {
		t.Fatal("send after drain failed")
	}

	if _, ok := (<-c.Recv()).GetPayload().(*model.ResyncPayload); !ok {
		t.Fatal("first event after a loss must be the resync hint")
	}

	if got := <-c.Recv(); got.GetID() != next.GetID() {
		t.Fatalf("got %s, want the queued event after resync", got.GetKind())
	}

	if c.(*connect).lostUpdates.Load() {
		t.Fatal("resync sent, loss flag must clear")
	}
}

// Losing an ephemeral event (typing) is not worth a catch-up.
func TestSend_DroppedTypingDoesNotResync(t *testing.T) {
	c := NewConnector(context.Background(), uuid.New(), 1, nil)
	defer c.Close()

	c.Send(highEvent(event.Typing), time.Millisecond)
	c.Send(highEvent(event.Typing), time.Millisecond)

	if c.(*connect).lostUpdates.Load() {
		t.Fatal("dropped typing must not ask for resync")
	}
}
