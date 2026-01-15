package registry

import (
	"sync"

	"github.com/google/uuid"
	"github.com/webitel/im-delivery-service/internal/domain/model"
)

// Hubber defines the gateway for user session management and event routing.
type Hubber interface {
	Broadcast(ev model.Eventer) bool
	Register(conn model.Connector)
	Unregister(userID, connID uuid.UUID)
	IsConnected(userID uuid.UUID) bool
}

// Hub implements a [SCALABLE_REGISTRY] using Virtual Cell pattern.
type Hub struct {
	// cells stores Map[uuid.UUID]Celler. Optimized for [READ_HEAVY] workloads.
	cells sync.Map
}

func NewHub() *Hub {
	return &Hub{}
}

func (h *Hub) IsConnected(userID uuid.UUID) bool {
	_, ok := h.cells.Load(userID)
	return ok
}

// Broadcast routes event to the specific [USER_CELL]. Returns false on miss or overflow.
func (h *Hub) Broadcast(ev model.Eventer) bool {
	if val, ok := h.cells.Load(ev.GetUserID()); ok {
		if cell, ok := val.(Celler); ok {
			return cell.Push(ev)
		}
	}
	return false
}

// Register ensures [IDEMPOTENT] cell creation and attaches a new transport.
func (h *Hub) Register(conn model.Connector) {
	uID := conn.GetUserID()
	// [LAZY_INIT] Create cell only when first connection arrives.
	val, _ := h.cells.LoadOrStore(uID, NewCell(uID))

	if cell, ok := val.(Celler); ok {
		cell.Attach(conn)
	}
}

// Unregister performs [GRACEFUL_RECLAMATION] of resources when sessions end.
func (h *Hub) Unregister(userID, connID uuid.UUID) {
	if val, ok := h.cells.Load(userID); ok {
		if cell, ok := val.(Celler); ok {
			// If no sessions left, purge the cell from memory.
			if cell.Detach(connID) {
				cell.Stop()
				h.cells.Delete(userID)
			}
		}
	}
}
