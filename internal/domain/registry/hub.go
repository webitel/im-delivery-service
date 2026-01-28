package registry

import (
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/webitel/im-delivery-service/internal/domain/event"
	"golang.org/x/sys/cpu"
)

// Interface guard
var _ Hubber = (*Hub)(nil)

// Hubber defines the external API for the high-concurrency registry.
// It acts as the entry point for both incoming events (Broadcast) and
// transport lifecycle management (Register/Unregister).
type Hubber interface {
	Broadcast(ev event.Eventer) bool
	Register(conn Connector)
	Unregister(userID, connID uuid.UUID)
	IsConnected(userID uuid.UUID) bool
	Shutdown()
}

const shardCount = 256

// Hub implements [Hubber] using a SHARDED_ACTOR architecture.
// This design eliminates global lock contention by partitioning the workload.
type Hub struct {
	// [CONCURRENCY_STRATEGY] Array of independent shards.
	// Each shard handles a subset of users based on their UUID.
	shards    []*shard
	config    hubConfig
	stopCh    chan struct{}
	closeOnce sync.Once
}

type hubConfig struct {
	evictionInterval time.Duration
	idleTimeout      time.Duration
	mailboxSize      int
}

// shard represents a logical partition of the user registry.
type shard struct {
	sync.RWMutex
	// [REGISTRY] Map of UserID to their dedicated delivery Cell (Actor).
	cells map[uuid.UUID]*Cell
	// Modern CPUs load data into L1/L2 caches in fixed-size blocks (Cache Lines),
	// typically 64 bytes. Without padding, multiple 'shard' instances would
	// sit on the same line.
	//
	// When Core-1 acquires a Mutex on Shard-A, it marks the entire Cache Line
	// as "modified". If Core-2 attempts to access Shard-B (even though it's
	// a different object), its cache is invalidated, forcing a slow reload
	// from L3 or RAM. This "False Sharing" can degrade performance by 10-50x
	// on high-core systems.
	//
	// cpu.CacheLinePad ensures each shard has its own exclusive line, enabling
	// true parallel execution without cross-core cache contention.
	_ cpu.CacheLinePad
}

// NewHub initializes the registry with [SHARDED_LOCKING] and starts the evictor.
func NewHub(opts ...Option) *Hub {
	h := &Hub{
		shards: make([]*shard, shardCount),
		config: hubConfig{
			evictionInterval: 1 * time.Minute,
			idleTimeout:      10 * time.Minute,
			mailboxSize:      1024,
		},
		stopCh: make(chan struct{}),
	}

	// [MEMORY_ALLOCATION] Pre-allocate all shards to prevent runtime pointer nil-checks.
	for i := range shardCount {
		h.shards[i] = &shard{cells: make(map[uuid.UUID]*Cell)}
	}

	for _, opt := range opts {
		opt(h)
	}

	// [BACKGROUND_PROCESS] Start the resource reclamation routine.
	go h.runEvictor()
	return h
}

// getShard maps a UserID to a specific shard using the first byte of the UUID.
// [LOCK_FREE_ROUTING] This operation requires no locks.
func (h *Hub) getShard(userID uuid.UUID) *shard {
	return h.shards[userID[0]]
}

// IsConnected checks if a user has an active [CELL] in the registry.
func (h *Hub) IsConnected(userID uuid.UUID) bool {
	s := h.getShard(userID)
	s.RLock()
	defer s.RUnlock()
	_, ok := s.cells[userID]
	return ok
}

// Broadcast dispatches an event to the specific user's [MAILBOX].
func (h *Hub) Broadcast(ev event.Eventer) bool {
	userID := ev.GetUserID()
	s := h.getShard(userID)

	// [READ_OPTIMIZATION] Use RLock for fast path event distribution.
	s.RLock()
	cell, ok := s.cells[userID]
	s.RUnlock()

	if ok {
		return cell.Push(ev)
	}
	return false
}

// Register performs an [IDEMPOTENT] registration of a new connection.
// It creates a new Cell (Actor) if the user is connecting for the first time.
func (h *Hub) Register(conn Connector) {
	userID := conn.GetUserID()
	s := h.getShard(userID)

	s.Lock()
	cell, ok := s.cells[userID]
	if !ok {
		// [ACTOR_CREATION] Initialize a new isolated delivery unit for the user.
		cell = NewCell(userID, h.config.mailboxSize)
		s.cells[userID] = cell
	}
	s.Unlock()

	// [SESSION_ATTACH] Delegate session management to the Cell.
	cell.Attach(conn)
}

// Unregister removes a specific connection from the user's [CELL].
func (h *Hub) Unregister(userID, connID uuid.UUID) {
	s := h.getShard(userID)
	s.RLock()
	cell, ok := s.cells[userID]
	s.RUnlock()

	if ok {
		cell.Detach(connID)
	}
}

// runEvictor is a long-running routine that triggers [CLEANUP] cycles.
func (h *Hub) runEvictor() {
	ticker := time.NewTicker(h.config.evictionInterval)
	defer ticker.Stop()

	for {
		select {
		case <-h.stopCh:
			return
		case <-ticker.C:
			h.performEviction()
		}
	}
}

// performEviction executes the [RECLAMATION] logic shard-by-shard.
func (h *Hub) performEviction() {
	reaped := 0
	for i := range shardCount {
		s := h.shards[i]

		// [GRANULAR_LOCKING] Lock only one shard at a time to keep others responsive.
		s.Lock()
		for id, cell := range s.cells {
			if cell.IsIdle(h.config.idleTimeout) {
				cell.Stop() // Terminate Actor goroutine
				delete(s.cells, id)
				reaped++
			}
		}
		s.Unlock()
	}

	if reaped > 0 {
		slog.Info("RESOURCE_RECLAIMED", "count", reaped, "shard_total", shardCount)
	}
}

// Shutdown ensures a [GRACEFUL_EXIT] by stopping all background actors exactly once.
func (h *Hub) Shutdown() {
	// [IDEMPOTENCY_GUARANTEE]
	// Use sync.Once to prevent 'panic: close of closed channel' if Shutdown is
	// triggered multiple times (e.g., by both SIGTERM and Uber Fx lifecycle).
	h.closeOnce.Do(func() {
		// 1. [SIGNAL_TERMINATION]
		// Broadcast closure to the background evictor goroutine.
		close(h.stopCh)

		// 2. [SHARD_DRAINING]
		// Iterate through all shards to stop individual User Cells.
		for i := range shardCount {
			s := h.shards[i]

			s.Lock()
			for _, cell := range s.cells {
				// [CASCADE_STOP]
				// Each Cell will stop its event loop and close its connectors,
				// triggering final delivery events to the clients.
				cell.Stop()
			}

			// 3. [MEMORY_MANAGEMENT]
			// Explicitly clear the map to release references and assist the
			// Garbage Collector in reclaiming memory for high-density sessions.
			s.cells = nil
			s.Unlock()
		}

		slog.Info("HUB_SHUTDOWN_COMPLETE",
			slog.Int("shards_processed", shardCount),
			slog.String("status", "graceful_drain_finished"),
		)
	})
}
