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

func messageEvent(cursor string) event.Eventer {
	return event.NewSystemEvent(uuid.New(), event.MessageCreated, &model.Message{UpdatesCursor: cursor},
		event.WithPriority[*model.Message](event.PriorityHigh))
}

// The resync replays from before the first loss, not from a later one.
func TestSend_ResyncCarriesFirstLostCursor(t *testing.T) {
	c := NewConnector(context.Background(), uuid.New(), 1, nil)
	defer c.Close()

	c.Send(messageEvent("10.1"), time.Millisecond)
	c.Send(messageEvent("20.2"), time.Millisecond)
	c.Send(messageEvent("30.3"), time.Millisecond)

	<-c.Recv()

	c.Send(messageEvent("40.4"), time.Millisecond)

	resync, ok := (<-c.Recv()).GetPayload().(*model.ResyncPayload)
	if !ok || resync.Cursor != "20.2" {
		t.Fatalf("resync = %+v, want cursor of the first dropped event 20.2", resync)
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

func statusEvent(status string) event.Eventer {
	return event.NewSystemEvent(uuid.New(), event.MessageStatusChanged, &model.MessageStatusUpdate{Status: status, UpdatesCursor: "7"},
		event.WithPriority[*model.MessageStatusUpdate](event.PriorityLow))
}

// Every status is journaled, so losing any of them asks for a resync.
func TestSend_LostStatusResyncs(t *testing.T) {
	for status, want := range map[string]bool{"read": true, "delivered": true, "failed": true} {
		c := NewConnector(context.Background(), uuid.New(), 1, nil)

		c.Send(highEvent(event.MessageCreated), time.Millisecond)
		c.Send(statusEvent(status), time.Millisecond)

		if got := c.(*connect).lostUpdates.Load(); got != want {
			t.Errorf("%s dropped: resync = %v, want %v", status, got, want)
		}

		c.Close()
	}
}
