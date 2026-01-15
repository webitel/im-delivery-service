package model

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

// Interface guard
var _ Connector = (*connect)(nil)

// [CONNECTOR] THE INTERFACE FOR EXTERNAL LAYERS (REGISTRY/HUB)
// This allows mocking and decoupling from the concrete implementation
type Connector interface {
	GetID() uuid.UUID
	GetUserID() uuid.UUID
	Send(ev InboundEventer, timeout time.Duration) bool // Thread-safe send with backpressure handling
	Recv() <-chan InboundEventer
	Close() // Terminate connection and release resources
}

// [METADATA] EXPORTED FOR TRANSPORT AND ANALYTICS LAYERS
type ConnectMetadata struct {
	Platform  string
	Version   string
	RemoteIP  string
	UserAgent string
}

// [CONNECT] CONCRETE IMPLEMENTATION (UNEXPORTED TO FORCE INTERFACE USAGE)
type connect struct {
	id        uuid.UUID
	userID    uuid.UUID
	metadata  ConnectMetadata
	createdAt time.Time

	ctx      context.Context
	cancelFn context.CancelFunc

	sendCh chan InboundEventer

	// [ATOMIC_FIELDS] Optimized for lock-free performance
	lastActivityAt int64
	droppedCount   uint64
}

// [POOL] SYNC.POOL FOR OBJECT REUSE (REDUCES GC PRESSURE)
var connectPool = sync.Pool{
	New: func() any {
		return &connect{}
	},
}

// [NEW_CONNECTOR] FACTORY FUNCTION USING POOLING
func NewConnector(ctx context.Context, userID uuid.UUID, bufferSize int) Connector {
	// Acquire from pool
	c := connectPool.Get().(*connect)

	childCtx, cancel := context.WithCancel(ctx)

	// Re-initialize fields
	c.id = uuid.New()
	c.userID = userID
	c.createdAt = time.Now()
	c.ctx = childCtx
	c.cancelFn = cancel
	c.sendCh = make(chan InboundEventer, bufferSize)

	atomic.StoreInt64(&c.lastActivityAt, time.Now().UnixNano())
	atomic.StoreUint64(&c.droppedCount, 0)

	return c
}

// --- IMPLEMENTATION OF CONNECTOR INTERFACE ---

func (c *connect) GetID() uuid.UUID     { return c.id }
func (c *connect) GetUserID() uuid.UUID { return c.userID }

// Send attempts to push an event into the channel.
// If the channel is full, it tries to evict lower priority events to make room.
func (c *connect) Send(ev InboundEventer, timeout time.Duration) bool {
	select {
	case <-c.ctx.Done():
		return false
	case c.sendCh <- ev:
		return true
	default:
		// Channel is full, initiate smart eviction strategy
		return c.handleBackpressure(ev, timeout)
	}
}

// handleBackpressure manages full buffers by dropping low-priority events.
func (c *connect) handleBackpressure(ev InboundEventer, timeout time.Duration) bool {
	// If the incoming event is low priority, drop it immediately to save buffer for high priority
	if ev.GetPriority() <= PriorityLow {
		atomic.AddUint64(&c.droppedCount, 1)
		return false
	}

	// Try to evict one existing low-priority event from the channel to make room
	// This is a simplified LIFO eviction for high-priority messages
	select {
	case oldEv := <-c.sendCh:
		if oldEv.GetPriority() < ev.GetPriority() {
			// Successfully replaced lower priority event with a higher one
			c.sendCh <- ev
			return true
		}
		// If the existing event was also high priority, put it back (best effort)
		select {
		case c.sendCh <- oldEv:
		default:
			// If we can't even put it back, it's lost
		}
	case <-time.After(timeout):
		// Hard timeout reached
	}

	atomic.AddUint64(&c.droppedCount, 1)
	return false
}

func (c *connect) Recv() <-chan InboundEventer { return c.sendCh }

// [CLOSE] RELEASES THE OBJECT BACK TO THE POOL
func (c *connect) Close() {
	c.cancelFn()

	// Reset pointers to prevent memory leaks while in pool
	c.sendCh = nil
	c.metadata = ConnectMetadata{}

	connectPool.Put(c)
}
