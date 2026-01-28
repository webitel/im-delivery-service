package registry

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/webitel/im-delivery-service/internal/domain/event"
)

// Interface guard
var _ Connector = (*connect)(nil)

// [CONNECTOR] THE INTERFACE FOR EXTERNAL LAYERS (REGISTRY/HUB)
// This allows mocking and decoupling from the concrete implementation
type Connector interface {
	GetID() uuid.UUID
	GetUserID() uuid.UUID
	Send(ev event.Eventer, timeout time.Duration) bool // Thread-safe send with backpressure handling
	Recv() <-chan event.Eventer
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
	id             uuid.UUID
	userID         uuid.UUID
	metadata       ConnectMetadata
	createdAt      time.Time
	ctx            context.Context
	cancelFn       context.CancelFunc
	sendCh         chan event.Eventer
	closeOnce      sync.Once // [PROTECTION]
	lastActivityAt int64     // [ATOMIC_FIELD]
	droppedCount   uint64    // [ATOMIC_FIELD]
}

// [POOL] SYNC.POOL FOR OBJECT REUSE (REDUCES GC PRESSURE)
var connectPool = sync.Pool{
	New: func() any {
		return &connect{}
	},
}

// [NEW_CONNECTOR] FACTORY FUNCTION USING POOLING
func NewConnector(ctx context.Context, userID uuid.UUID, bufferSize int) Connector {
	c := connectPool.Get().(*connect)

	// [INITIALIZATION]
	// Delegate state setup to the reset method to ensure a clean slate.
	c.reset(ctx, userID, bufferSize)

	return c
}

// reset re-initializes the connector's internal state using a struct literal.
// This is the cleanest way to wipe 'stale' data from pooled objects and reset the sync.Once guard.
func (c *connect) reset(ctx context.Context, userID uuid.UUID, bufferSize int) {
	childCtx, cancel := context.WithCancel(ctx)

	// [BLANK_SLATE_ASSIGNMENT]
	// By reassigning the pointer's value to a new literal, we ensure all fields,
	// including metadata and counters, are reset to their zero-values or defaults.
	*c = connect{
		id:             uuid.New(),
		userID:         userID,
		createdAt:      time.Now(),
		ctx:            childCtx,
		cancelFn:       cancel,
		sendCh:         make(chan event.Eventer, bufferSize),
		lastActivityAt: time.Now().UnixNano(),
	}
}

// --- IMPLEMENTATION OF CONNECTOR INTERFACE ---

func (c *connect) GetID() uuid.UUID     { return c.id }
func (c *connect) GetUserID() uuid.UUID { return c.userID }

// Send attempts to push an event into the channel.
// If the channel is full, it tries to evict lower priority events to make room.
func (c *connect) Send(ev event.Eventer, timeout time.Duration) bool {
	// [RESOURCE_MANAGEMENT] Create a localized context to enforce a strict delivery window.
	// This ensures that the User Cell is not held hostage by a single stalled session.
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	select {
	// 1. [LIFECYCLE_GATE] Immediately abort if the underlying transport is already dead.
	case <-c.ctx.Done():
		return false

	// 2. [PRIMARY_DELIVERY] Attempt to enqueue the event into the session's mailbox.
	// Unlike a 'default' block, this will wait up to 'timeout' for space to become available,
	// which smooths out transient network jitter.
	case c.sendCh <- ev:
		return true

	// 3. [BACKPRESSURE_THRESHOLD] Triggered if the buffer remains saturated for the entire duration.
	// This indicates a persistent slow consumer or network congestion.
	case <-ctx.Done():
		// Initiate smart eviction or shedding logic to preserve system throughput.
		return c.handleBackpressure(ev, timeout)
	}
}

// handleBackpressure manages full buffers by dropping low-priority events.
func (c *connect) handleBackpressure(ev event.Eventer, timeout time.Duration) bool {
	// If the incoming event is low priority, drop it immediately to save buffer for high priority
	if ev.GetPriority() <= event.PriorityLow {
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

func (c *connect) Recv() <-chan event.Eventer { return c.sendCh }

// Close terminates the session, triggers cleanup, and recycles the object.
func (c *connect) Close() {
	// [IDEMPOTENCY_SHIELD]
	// Ensures the teardown logic runs exactly once. This prevents "panic: close of closed channel"
	// and double-entry corruption of the sync.Pool when called concurrently
	// by the Hub (shutdown), Cell (eviction), or gRPC handler (defer).
	c.closeOnce.Do(func() {
		// 1. [SIGNAL_ABORT] Immediately cancel the context to stop any pending Send operations.
		c.cancelFn()

		// 2. [UPSTREAM_NOTIFY] Closing the channel signals the gRPC stream handler (via !ok)
		// to send a final 'Disconnected' event and exit the loop gracefully.
		if c.sendCh != nil {
			close(c.sendCh)
		}

		// 3. [MEMORY_SANITIZATION]
		// Zero out references to prevent memory leaks while the object is idle in the pool.
		// This ensures the next user of this pooled object starts with a clean slate.
		c.sendCh = nil
		c.metadata = ConnectMetadata{}

		// 4. [RESOURCE_RECYCLING] Return the sanitized structure to reduce GC allocation pressure.
		connectPool.Put(c)
	})
}
