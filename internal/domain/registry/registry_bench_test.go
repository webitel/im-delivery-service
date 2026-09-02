package registry

import (
	"context"
	"runtime"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/webitel/im-delivery-service/internal/domain/event"
	"go.uber.org/goleak"
)

// newEvent builds a minimal, valid Eventer for benchmarking the delivery path.
func newEvent(userID uuid.UUID) event.Eventer {
	return &event.Envelope[string]{
		ID:         uuid.New(),
		UserID:     userID,
		Kind:       0,
		Priority:   event.PriorityNormal,
		Payload:    "benchmark-payload",
		OccurredAt: 0,
	}
}

// -----------------------------------------------------------------------------
// 1. MEMORY PER CONNECTED USER
//
// Registers N users (each with one attached connector) and reports the resident
// heap growth per user. This quantifies the "always-on" cost of the sharded
// actor model: cell mailbox channel + connector sendCh + cell goroutine stack.
// -----------------------------------------------------------------------------

func benchMemoryPerUser(b *testing.B, users, mailbox, connBuf int) {
	b.Helper()
	b.ReportAllocs()

	runtime.GC()
	var before runtime.MemStats
	runtime.ReadMemStats(&before)

	h := NewHub(
		WithMailboxSize(mailbox),
		WithEvictionInterval(time.Hour), // keep evictor idle during measurement
		WithIdleTimeout(time.Hour),
	)

	conns := make([]Connector, users)
	for i := range users {
		uid := uuid.New()
		c := NewConnector(context.Background(), uid, connBuf, nil)
		h.Register(c)
		conns[i] = c
	}

	runtime.GC()
	var after runtime.MemStats
	runtime.ReadMemStats(&after)

	heapPerUser := float64(after.HeapAlloc-before.HeapAlloc) / float64(users)
	b.ReportMetric(heapPerUser, "heapBytes/user")
	b.ReportMetric(float64(after.HeapAlloc-before.HeapAlloc)/(1024*1024), "totalHeapMB")

	// keep references alive until after measurement
	runtime.KeepAlive(conns)
	h.Shutdown()
}

func BenchmarkMemoryPerUser_Prod(b *testing.B) {
	// actual production wiring: hub mailbox 2048 (module.go), connector buffer 1024 (session.go)
	benchMemoryPerUser(b, 50_000, 2048, 1024)
}

func BenchmarkMemoryPerUser_Tuned(b *testing.B) {
	// proposed right-sized buffers: mailbox 256, connector 128
	benchMemoryPerUser(b, 50_000, 256, 128)
}

func BenchmarkMemoryPerUser_SmallMailbox(b *testing.B) {
	// lower bound: shows how much RAM the channel buffers alone cost
	benchMemoryPerUser(b, 50_000, 32, 32)
}

// -----------------------------------------------------------------------------
// 2. BROADCAST THROUGHPUT + ALLOCATIONS
//
// Measures the CPU cost and per-event allocations of the hot delivery path:
// Hub.Broadcast -> shard lookup -> Cell.Push -> mailbox -> deliver -> conn.Send.
// -----------------------------------------------------------------------------

func BenchmarkBroadcast_SingleSession(b *testing.B) {
	h := NewHub(WithMailboxSize(1024), WithEvictionInterval(time.Hour), WithIdleTimeout(time.Hour))
	defer h.Shutdown()

	uid := uuid.New()
	c := NewConnector(context.Background(), uid, 4096, nil)
	h.Register(c)

	// drain the connector so the mailbox never blocks
	go func() {
		for range c.Recv() {
		}
	}()

	ev := newEvent(uid)
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		h.Broadcast(ev)
	}
}

func BenchmarkBroadcast_FanoutManyUsers(b *testing.B) {
	const users = 10_000
	h := NewHub(WithMailboxSize(1024), WithEvictionInterval(time.Hour), WithIdleTimeout(time.Hour))
	defer h.Shutdown()

	ids := make([]uuid.UUID, users)
	for i := range users {
		uid := uuid.New()
		ids[i] = uid
		c := NewConnector(context.Background(), uid, 4096, nil)
		h.Register(c)
		go func() {
			for range c.Recv() {
			}
		}()
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := range b.N {
		h.Broadcast(newEvent(ids[i%users]))
	}
}

// -----------------------------------------------------------------------------
// 3. GOROUTINE LEAK CHECK
//
// Registers and unregisters users in cycles, then shuts the hub down and asserts
// no goroutines leaked. Directly tests the "goroutines/descriptors grow and are
// never released" hypothesis for the cell lifecycle.
// -----------------------------------------------------------------------------

func TestNoGoroutineLeak_RegisterUnregister(t *testing.T) {
	defer goleak.VerifyNone(t)

	h := NewHub(WithEvictionInterval(time.Hour), WithIdleTimeout(time.Hour))

	for cycle := range 5 {
		_ = cycle
		conns := make([]Connector, 0, 1000)
		for range 1000 {
			uid := uuid.New()
			c := NewConnector(context.Background(), uid, 32, nil)
			h.Register(c)
			conns = append(conns, c)
		}
		for _, c := range conns {
			h.Unregister(c.GetUserID(), c.GetID())
			c.Close()
		}
	}

	// Evict idle (session-less) cells so their loop() goroutines exit.
	h.performEviction()
	h.Shutdown()

	// give the runtime a moment to reap stopped goroutines
	time.Sleep(200 * time.Millisecond)
}
