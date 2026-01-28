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
	"github.com/webitel/im-delivery-service/internal/domain/event"
)

// Celler defines the internal API for user-specific delivery units.
type Celler interface {
	Push(ev event.Eventer) bool
	Attach(conn Connector)
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
	mailbox chan event.Eventer

	// [SESSIONS]
	// Registry of all active transport channels (gRPC streams) for the user.
	// Allows multiplexing a single event to multiple devices (mobile, web, desktop).
	sessions map[uuid.UUID]Connector

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
		mailbox:          make(chan event.Eventer, bufferSize),
		sessions:         make(map[uuid.UUID]Connector),
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

func (c *Cell) Push(ev event.Eventer) bool {
	c.touch()
	select {
	case c.mailbox <- ev:
		return true
	default:
		// [BACKPRESSURE] Drop event if mailbox is full to protect system stability
		return false
	}
}

func (c *Cell) Attach(conn Connector) {
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
			// [STRATEGY: BATCH_DRAINING]
			// Once awakened, don't return to the expensive 'select' immediately.
			// Tight loop to drain pending events reduces scheduler overhead.
			c.deliver(ev)

			// Attempt to drain up to 64 events in one go to smooth out bursts.
			// This number is a sweet spot between latency and CPU fairness.
			for range 64 {
				select {
				case nextEv := <-c.mailbox:
					c.deliver(nextEv)
				default:
					// Mailbox empty, go back to wait
					goto wait
				}
			}
		wait:
		}
	}
}

// deliver broadcasts events to all active sessions of the user.
func (c *Cell) deliver(ev event.Eventer) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if len(c.sessions) == 0 {
		return
	}

	for _, conn := range c.sessions {
		// Strict 250ms window. If a connection is slow, it won't kill the Actor loop.
		conn.Send(ev, time.Millisecond*250)
	}
}

func (c *Cell) Stop() {
	close(c.doneCh)

	c.mu.Lock()
	defer c.mu.Unlock()
	for id, conn := range c.sessions {
		conn.Close()
		delete(c.sessions, id)
	}
}
