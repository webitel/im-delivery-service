package registry

import (
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/webitel/im-delivery-service/internal/domain/model"
	"github.com/webitel/im-delivery-service/internal/handler/marshaller"
)

// [HUBBER_INTERFACE] DEFINES CORE LOGIC FOR CONNECTION MANAGEMENT
type Hubber interface {
	Register(conn model.Connector)
	Unregister(userID, connID uuid.UUID)
	Broadcast(ev model.Eventer) bool
	IsConnected(userID uuid.UUID) bool // [PERFORMANCE] FAST CHECK FOR ROUTING KEY FILTERING
}

const shardCount = 256

// [HUB_IMPLEMENTATION] SHARDED STORAGE TO MINIMIZE LOCK CONTENTION
type Hub struct {
	shards [shardCount]*shard
}

type shard struct {
	sync.RWMutex
	sessions map[uuid.UUID][]model.Connector
}

func NewHub() *Hub {
	h := &Hub{}
	for i := range shardCount {
		h.shards[i] = &shard{sessions: make(map[uuid.UUID][]model.Connector)}
	}
	return h
}

func (h *Hub) getShard(userID uuid.UUID) *shard {
	return h.shards[userID[0]]
}

// [IS_ACTIVE] CHECKS IF USER HAS AT LEAST ONE ACTIVE STREAM ON THIS NODE
func (h *Hub) IsConnected(userID uuid.UUID) bool {
	s := h.getShard(userID)
	s.RLock()
	defer s.RUnlock()
	_, ok := s.sessions[userID]
	return ok
}

func (h *Hub) Broadcast(ev model.Eventer) bool {
	s := h.getShard(ev.GetUserID())
	s.RLock()
	connectors, ok := s.sessions[ev.GetUserID()]
	s.RUnlock()

	if !ok || len(connectors) == 0 {
		return false
	}

	// [OPTIMIZATION] Marshall once here. All subsequent calls to
	// Marshaller inside gRPC streams will use this cached value.
	_ = marshaller.MarshallDeliveryEvent(ev)

	var delivered bool
	for _, conn := range connectors {
		// All connectors get the same pointer to the event with cached data.
		if conn.Send(ev, time.Second*2) {
			delivered = true
		}
	}

	return delivered
}

func (h *Hub) Register(conn model.Connector) {
	s := h.getShard(conn.GetUserID())
	s.Lock()
	defer s.Unlock()
	s.sessions[conn.GetUserID()] = append(s.sessions[conn.GetUserID()], conn)
}

func (h *Hub) Unregister(userID, connID uuid.UUID) {
	s := h.getShard(userID)
	s.Lock()
	defer s.Unlock()

	conns, ok := s.sessions[userID]
	if !ok {
		return
	}

	for i, c := range conns {
		if c.GetID() == connID {
			c.Close() // [LIFECYCLE] TRIGGER CLEANUP AND POOL RETURN
			conns[i] = conns[len(conns)-1]
			s.sessions[userID] = conns[:len(conns)-1]
			break
		}
	}

	if len(s.sessions[userID]) == 0 {
		delete(s.sessions, userID)
	}
}
