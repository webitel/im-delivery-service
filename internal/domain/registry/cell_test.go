package registry

import (
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/webitel/im-delivery-service/internal/domain/event"
	"github.com/webitel/im-delivery-service/internal/domain/model"
)

// fakeConnector implements Connector for testing.
type fakeConnector struct {
	id     uuid.UUID
	userID uuid.UUID
	mu     sync.Mutex
	events []event.Eventer
	filter func(string) bool // nil = always allow
}

func (f *fakeConnector) GetID() uuid.UUID     { return f.id }
func (f *fakeConnector) GetUserID() uuid.UUID { return f.userID }
func (f *fakeConnector) Recv() <-chan event.Eventer {
	return nil // not used in tests
}

func (f *fakeConnector) Send(ev event.Eventer, timeout time.Duration) bool {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.events = append(f.events, ev)

	return true
}

func (f *fakeConnector) SystemMessageAllowed(systemType string) bool {
	if f.filter == nil {
		return true
	}

	return f.filter(systemType)
}
func (f *fakeConnector) Close() {}

func (f *fakeConnector) getEvents() []event.Eventer {
	f.mu.Lock()
	defer f.mu.Unlock()

	events := make([]event.Eventer, len(f.events))
	copy(events, f.events)

	return events
}

func TestCellDeliver(t *testing.T) {
	userID := uuid.New()

	t.Run("system message with matching filter is delivered", func(t *testing.T) {
		conn := &fakeConnector{
			id:     uuid.New(),
			userID: userID,
			filter: func(st string) bool { return st == "user_joined" },
		}

		cell := &Cell{
			userID:   userID,
			sessions: map[uuid.UUID]Connector{conn.id: conn},
		}

		msg := &model.Message{
			ID:        uuid.New(),
			From:      model.Peer{Name: "System"},
			CreatedAt: time.Now().UnixMilli(),
			System: &model.System{
				Type: "user_joined",
			},
		}
		ev := &event.Envelope[*model.Message]{
			ID:       uuid.New(),
			Kind:     event.MessageCreated,
			Payload:  msg,
			UserID:   userID,
			Priority: event.PriorityHigh,
		}

		cell.deliver(ev)

		events := conn.getEvents()
		if len(events) != 1 {
			t.Errorf("expected 1 event, got %d", len(events))
		}
	})

	t.Run("system message with non-matching filter is not delivered", func(t *testing.T) {
		conn := &fakeConnector{
			id:     uuid.New(),
			userID: userID,
			filter: func(st string) bool { return st == "user_left" },
		}

		cell := &Cell{
			userID:   userID,
			sessions: map[uuid.UUID]Connector{conn.id: conn},
		}

		msg := &model.Message{
			ID:        uuid.New(),
			From:      model.Peer{Name: "System"},
			CreatedAt: time.Now().UnixMilli(),
			System: &model.System{
				Type: "user_joined",
			},
		}
		ev := &event.Envelope[*model.Message]{
			ID:       uuid.New(),
			Kind:     event.MessageCreated,
			Payload:  msg,
			UserID:   userID,
			Priority: event.PriorityHigh,
		}

		cell.deliver(ev)

		events := conn.getEvents()
		if len(events) != 0 {
			t.Errorf("expected 0 events, got %d", len(events))
		}
	})

	t.Run("message with nil System is delivered even with always-false filter", func(t *testing.T) {
		conn := &fakeConnector{
			id:     uuid.New(),
			userID: userID,
			filter: func(st string) bool { return false },
		}

		cell := &Cell{
			userID:   userID,
			sessions: map[uuid.UUID]Connector{conn.id: conn},
		}

		msg := &model.Message{
			ID:        uuid.New(),
			From:      model.Peer{Name: "Alice"},
			CreatedAt: time.Now().UnixMilli(),
			System:    nil,
		}
		ev := &event.Envelope[*model.Message]{
			ID:       uuid.New(),
			Kind:     event.MessageCreated,
			Payload:  msg,
			UserID:   userID,
			Priority: event.PriorityHigh,
		}

		cell.deliver(ev)

		events := conn.getEvents()
		if len(events) != 1 {
			t.Errorf("expected 1 event, got %d", len(events))
		}
	})

	t.Run("non-MessageCreated event is delivered even with always-false filter", func(t *testing.T) {
		conn := &fakeConnector{
			id:     uuid.New(),
			userID: userID,
			filter: func(st string) bool { return false },
		}

		cell := &Cell{
			userID:   userID,
			sessions: map[uuid.UUID]Connector{conn.id: conn},
		}

		ev := &event.Envelope[any]{
			ID:       uuid.New(),
			Kind:     event.Connected,
			UserID:   userID,
			Priority: event.PriorityHigh,
		}

		cell.deliver(ev)

		events := conn.getEvents()
		if len(events) != 1 {
			t.Errorf("expected 1 event, got %d", len(events))
		}
	})

	t.Run("event is delivered when filter field is nil", func(t *testing.T) {
		conn := &fakeConnector{
			id:     uuid.New(),
			userID: userID,
			filter: nil,
		}

		cell := &Cell{
			userID:   userID,
			sessions: map[uuid.UUID]Connector{conn.id: conn},
		}

		msg := &model.Message{
			ID:        uuid.New(),
			From:      model.Peer{Name: "System"},
			CreatedAt: time.Now().UnixMilli(),
			System: &model.System{
				Type: "user_joined",
			},
		}
		ev := &event.Envelope[*model.Message]{
			ID:       uuid.New(),
			Kind:     event.MessageCreated,
			Payload:  msg,
			UserID:   userID,
			Priority: event.PriorityHigh,
		}

		cell.deliver(ev)

		events := conn.getEvents()
		if len(events) != 1 {
			t.Errorf("expected 1 event, got %d", len(events))
		}
	})

	t.Run("panic in filter does not stop delivery to other connectors", func(t *testing.T) {
		panicConn := &fakeConnector{
			id:     uuid.New(),
			userID: userID,
			filter: func(st string) bool { panic("test panic") },
		}

		goodConn := &fakeConnector{
			id:     uuid.New(),
			userID: userID,
			filter: func(st string) bool { return true },
		}

		cell := &Cell{
			userID: userID,
			sessions: map[uuid.UUID]Connector{
				panicConn.id: panicConn,
				goodConn.id:  goodConn,
			},
		}

		msg := &model.Message{
			ID:        uuid.New(),
			From:      model.Peer{Name: "System"},
			CreatedAt: time.Now().UnixMilli(),
			System: &model.System{
				Type: "user_joined",
			},
		}
		ev := &event.Envelope[*model.Message]{
			ID:       uuid.New(),
			Kind:     event.MessageCreated,
			Payload:  msg,
			UserID:   userID,
			Priority: event.PriorityHigh,
		}

		cell.deliver(ev)

		goodEvents := goodConn.getEvents()
		if len(goodEvents) != 1 {
			t.Errorf("expected good connector to receive 1 event, got %d", len(goodEvents))
		}

		panicEvents := panicConn.getEvents()
		if len(panicEvents) != 1 {
			t.Errorf("expected panic connector to receive 1 event (fail-open), got %d", len(panicEvents))
		}
	})
}
