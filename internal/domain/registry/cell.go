/*
Package registry provides a high-performance event distribution system based on the Actor Model.

Key Architectural Concepts:
  - Virtual Cells: Every active user is represented by an isolated 'Cell' (Actor) that
    encapsulates all concurrent gRPC streams (sessions) for that specific identity.
  - Decoupling & Backpressure: Through the use of internal per-user mailboxes, the
    package ensures that slow network consumers do not block global system throughput.
  - Computational Efficiency: Events are marshaled into the wire format exactly once
    per user group, leveraging internal caching to minimize CPU and GC overhead.
  - Concurrency Management: Utilizes lock-free lookups via sync.Map and fine-grained
    sharded locking within individual cells to eliminate global mutex contention.
*/
package registry

import (
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/webitel/im-delivery-service/internal/domain/model"
)

// Celler defines the internal API for user-specific delivery units.
type Celler interface {
	Push(ev model.Eventer) bool
	Attach(conn model.Connector)
	Detach(connID uuid.UUID) bool
	IsIdle(timeout time.Duration) bool
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

	// lastActivityAt records the last time an event was processed for this cell.
	lastActivityAt time.Time
}

// internal/domain/registry/cell.go

func NewCell(userID uuid.UUID, bufferSize int) *Cell {
	c := &Cell{
		userID:         userID,
		mailbox:        make(chan model.Eventer, bufferSize), // [DYNAMIC_BUFFER]
		sessions:       make(map[uuid.UUID]model.Connector),
		doneCh:         make(chan struct{}),
		lastActivityAt: time.Now(),
	}
	go c.loop()
	return c
}

// IsIdle returns true if the user has no active sessions and hasn't received events lately.
func (c *Cell) IsIdle(timeout time.Duration) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// A cell is considered idle only if it has NO active connections
	// and the quiet period has exceeded the threshold.
	return len(c.sessions) == 0 && time.Since(c.lastActivityAt) > timeout
}

func (c *Cell) touch() {
	c.mu.Lock()
	c.lastActivityAt = time.Now()
	c.mu.Unlock()
}

func (c *Cell) Push(ev model.Eventer) bool {
	c.touch() // Keep alive on incoming events
	select {
	case c.mailbox <- ev:
		return true
	default:
		return false
	}
}

func (c *Cell) Attach(conn model.Connector) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.lastActivityAt = time.Now()
	c.sessions[conn.GetID()] = conn
}

func (c *Cell) Detach(connID uuid.UUID) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.sessions, connID)
	c.lastActivityAt = time.Now()
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

	for _, conn := range c.sessions {
		conn.Send(ev, time.Millisecond*500)
	}
}

func (c *Cell) Stop() {
	close(c.doneCh)
}
