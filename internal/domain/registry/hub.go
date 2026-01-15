package registry

import (
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/webitel/im-delivery-service/internal/domain/model"
)

// Hubber defines the external API for the registry system.
type Hubber interface {
	Broadcast(ev model.Eventer) bool
	Register(conn model.Connector)
	Unregister(userID, connID uuid.UUID)
	IsConnected(userID uuid.UUID) bool
	Shutdown()
}

// Hub implements [Hubber] using a Virtual Cell (Actor) architecture.
type Hub struct {
	// cells maintains an active registry of UserID -> Celler.
	cells sync.Map

	// [EVICTION_POLICY]
	evictionInterval time.Duration
	idleTimeout      time.Duration
	stopCh           chan struct{}
}

// NewHub initializes the hub and starts the background eviction (Janitor) process.
func NewHub(evictionInterval, idleTimeout time.Duration) *Hub {
	h := &Hub{
		evictionInterval: evictionInterval,
		idleTimeout:      idleTimeout,
		stopCh:           make(chan struct{}),
	}
	go h.runEvictor()
	return h
}

// IsConnected checks if a user cell exists in the registry.
func (h *Hub) IsConnected(userID uuid.UUID) bool {
	_, ok := h.cells.Load(userID)
	return ok
}

// Broadcast dispatches an event to the specific user's cell mailbox.
func (h *Hub) Broadcast(ev model.Eventer) bool {
	if val, ok := h.cells.Load(ev.GetUserID()); ok {
		if cell, ok := val.(Celler); ok {
			return cell.Push(ev)
		}
	}
	return false
}

// Register performs an [IDEMPOTENT] registration of a new connection.
func (h *Hub) Register(conn model.Connector) {
	uID := conn.GetUserID()
	// [LAZY_INIT] Spawns a new cell only if one doesn't already exist for this user.
	val, _ := h.cells.LoadOrStore(uID, NewCell(uID))

	if cell, ok := val.(Celler); ok {
		cell.Attach(conn)
	}
}

// Unregister removes a connection from a cell.
// Reclamation of the cell itself is handled asynchronously by the Evictor.
func (h *Hub) Unregister(userID, connID uuid.UUID) {
	if val, ok := h.cells.Load(userID); ok {
		if cell, ok := val.(Celler); ok {
			cell.Detach(connID)
		}
	}
}

func (h *Hub) runEvictor() {
	ticker := time.NewTicker(h.evictionInterval)
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

// performEviction executes the [RESOURCE_RECLAMATION] cycle.
func (h *Hub) performEviction() {
	reapedCount := 0
	h.cells.Range(func(key, value any) bool {
		if cell, ok := value.(Celler); ok {
			if cell.IsIdle(h.idleTimeout) {
				cell.Stop()
				h.cells.Delete(key)
				reapedCount++
			}
		}
		return true
	})

	if reapedCount > 0 {
		log.Printf("[Hub] Eviction complete. Reclaimed %d idle user cells.", reapedCount)
	}
}

// Shutdown gracefully stops the hub and all managed cells.
func (h *Hub) Shutdown() {
	close(h.stopCh)
	h.cells.Range(func(key, value any) bool {
		if cell, ok := value.(Celler); ok {
			cell.Stop()
		}
		return true
	})
}
