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
	"sync/atomic"
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

	// [OPTIMIZATION] Atomic timestamp to avoid mutex contention during activity checks
	lastActivityUnix int64
}

func NewCell(userID uuid.UUID, bufferSize int) *Cell {
	c := &Cell{
		userID:           userID,
		mailbox:          make(chan model.Eventer, bufferSize),
		sessions:         make(map[uuid.UUID]model.Connector),
		doneCh:           make(chan struct{}),
		lastActivityUnix: time.Now().Unix(),
	}
	go c.loop()
	return c
}

// touch updates the last activity timestamp using atomic store
func (c *Cell) touch() {
	atomic.StoreInt64(&c.lastActivityUnix, time.Now().Unix())
}

// IsIdle checks if the cell can be reclaimed based on session count and inactivity
func (c *Cell) IsIdle(timeout time.Duration) bool {
	c.mu.RLock()
	hasSessions := len(c.sessions) > 0
	c.mu.RUnlock()

	if hasSessions {
		return false
	}

	lastActivity := time.Unix(atomic.LoadInt64(&c.lastActivityUnix), 0)
	return time.Since(lastActivity) > timeout
}

func (c *Cell) Push(ev model.Eventer) bool {
	c.touch()
	select {
	case c.mailbox <- ev:
		return true
	default:
		// [BACKPRESSURE] Drop event if mailbox is full to protect system stability
		return false
	}
}

func (c *Cell) Attach(conn model.Connector) {
	c.mu.Lock()
	c.sessions[conn.GetID()] = conn
	c.mu.Unlock()
	c.touch()
}

func (c *Cell) Detach(connID uuid.UUID) bool {
	c.mu.Lock()
	delete(c.sessions, connID)
	isEmpty := len(c.sessions) == 0
	c.mu.Unlock()
	c.touch()
	return isEmpty
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

// deliver broadcasts events to all active sessions of the user
func (c *Cell) deliver(ev model.Eventer) {
	// [STRATEGY] Snapshot sessions under RLock to minimize lock holding time
	// during potentially slow network I/O
	c.mu.RLock()
	if len(c.sessions) == 0 {
		c.mu.RUnlock()
		return
	}
	conns := make([]model.Connector, 0, len(c.sessions))
	for _, conn := range c.sessions {
		conns = append(conns, conn)
	}
	c.mu.RUnlock()

	for _, conn := range conns {
		// [RELIABILITY] Use a strict timeout to prevent slow streams from hanging the Actor
		conn.Send(ev, time.Millisecond*250)
	}
}

func (c *Cell) Stop() {
	close(c.doneCh)
}
