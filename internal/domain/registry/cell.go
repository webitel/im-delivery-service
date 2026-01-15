package registry

import (
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/webitel/im-delivery-service/internal/domain/model"
	"github.com/webitel/im-delivery-service/internal/handler/marshaller"
)

// Celler defines the internal API for user-specific delivery units.
type Celler interface {
	Push(ev model.Eventer) bool
	Attach(conn model.Connector)
	Detach(connID uuid.UUID) bool
	Stop()
}

// Cell implements [ISOLATED_DELIVERY] logic for a single user.
type Cell struct {
	// [IDENTITY]
	// The unique identifier of the user managed by this actor instance.
	userID uuid.UUID

	// [MAILBOX]
	// Buffered channel that decouples the global dispatcher from individual delivery.
	// It acts as a shock absorber, preventing slow consumer latency from
	// propagating back to the Hub or AMQP consumers (Backpressure).
	mailbox chan model.Eventer

	// [SESSIONS]
	// Registry of all active transport channels (gRPC streams) for the user.
	// Allows multiplexing a single event to multiple devices (mobile, web, desktop).
	sessions map[uuid.UUID]model.Connector

	// [CONCURRENCY_CONTROL]
	// Fine-grained lock for managing the sessions map.
	// RWMutex is chosen because read-heavy delivery operations outnumber
	// write-heavy registration events.
	mu sync.RWMutex

	// [LIFECYCLE_CONTROL]
	// Signaling channel used to terminate the background goroutine.
	// Ensures no goroutine leaks occur after the user goes offline.
	doneCh chan struct{}
}

func NewCell(userID uuid.UUID) *Cell {
	c := &Cell{
		userID:   userID,
		mailbox:  make(chan model.Eventer, 1024),
		sessions: make(map[uuid.UUID]model.Connector),
		doneCh:   make(chan struct{}),
	}
	go c.loop()
	return c
}

// Push adds event to [MAILBOX]. Returns false if buffer is saturated (Backpressure).
func (c *Cell) Push(ev model.Eventer) bool {
	select {
	case c.mailbox <- ev:
		return true
	default:
		return false
	}
}

// Attach adds a connection. Logic is [THREAD_SAFE] and scoped to this user only.
func (c *Cell) Attach(conn model.Connector) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.sessions[conn.GetID()] = conn
}

// Detach removes a connection. Returns true if the cell has [ZERO_SESSIONS].
func (c *Cell) Detach(connID uuid.UUID) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.sessions, connID)
	return len(c.sessions) == 0
}

func (c *Cell) loop() {
	for {
		select {
		case <-c.doneCh:
			return
		case ev := <-c.mailbox:
			c.deliver(ev)
		}
	}
}

func (c *Cell) deliver(ev model.Eventer) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if len(c.sessions) == 0 {
		return
	}

	// [WARM_UP] Marshal once per user cell. Subsequent devices hit the cache.
	_ = marshaller.MarshallDeliveryEvent(ev)

	for _, conn := range c.sessions {
		// [NON_BLOCKING] delivery attempt.
		conn.Send(ev, time.Millisecond*500)
	}
}

func (c *Cell) Stop() {
	close(c.doneCh)
}
