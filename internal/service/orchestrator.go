// internal/service/event_orchestrator.go
package service

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/fx"

	"github.com/webitel/im-delivery-service/internal/domain/event"
	"github.com/webitel/im-delivery-service/internal/domain/registry"
	"github.com/webitel/im-delivery-service/internal/store"
)

// --- Contracts ---
// [EVENT_HANDLER] Standard interface for background event processing (e.g., Push, DB).
type EventHandler interface {
	Handle(ctx context.Context, ev event.Eventer)
}

// [DISMISS_HANDLER] Optional interface for handlers capable of revoking notifications.
type DismissHandler interface {
	HandleDismiss(ctx context.Context, ev event.Eventer)
}

// [ORCHESTRATOR] The central authority for event life-cycle management.
type Orchestrator interface {
	Notify(ctx context.Context, ev event.Eventer)
	Dismiss(ctx context.Context, ev event.Eventer)
	Ack(ctx context.Context, eid, cid uuid.UUID) error
}

// --- Orchestrator Implementation ---
var _ Orchestrator = (*EventOrchestrator)(nil)

type OrchestratorParams struct {
	fx.In

	Hub        registry.Hubber
	Tracker    store.DeliveryTracker
	Handlers   []EventHandler `group:"event_handlers"` // [AUTO_DISCOVERY]
	Confirmer  DeliveryConfirmer
	Log        *slog.Logger
	Workers    int           `name:"worker_count"`
	AckTimeout time.Duration `name:"ack_timeout"`
}

type EventOrchestrator struct {
	hub        registry.Hubber
	tracker    store.DeliveryTracker
	handlers   []EventHandler
	confirmer  DeliveryConfirmer
	queue      chan task
	log        *slog.Logger
	wg         sync.WaitGroup
	ackTimeout time.Duration
}

type task struct {
	ctx         context.Context
	ev          event.Eventer
	dismissable bool
}

func NewEventOrchestrator(p OrchestratorParams) *EventOrchestrator {
	o := &EventOrchestrator{
		hub:        p.Hub,
		tracker:    p.Tracker,
		handlers:   p.Handlers,
		confirmer:  p.Confirmer,
		log:        p.Log.With("service", "event_orchestrator"),
		queue:      make(chan task, 8192),
		ackTimeout: p.AckTimeout,
	}

	workerCount := p.Workers
	if workerCount <= 0 {
		workerCount = 16
	}

	o.bootstrap(workerCount)

	return o
}

// [PUBLISH] Fan-out to real-time streams and enqueue for multi-channel delivery.
func (o *EventOrchestrator) Notify(ctx context.Context, ev event.Eventer) {
	if ev == nil {
		return
	}

	// [HOT_PATH] Direct WebSocket broadcast.
	o.hub.Broadcast(ev)

	// [ASYNC_PATH] Background processing.
	o.enqueue(ctx, ev, false)
}

// [DISMISS] Signals all handlers to revoke or silence a previously published event.
func (o *EventOrchestrator) Dismiss(ctx context.Context, ev event.Eventer) {
	o.enqueue(ctx, ev, true)
}

// [ACK] Finalizes the delivery process for a specific session.
func (o *EventOrchestrator) Ack(ctx context.Context, eid, cid uuid.UUID) error {
	o.log.Info("ACK_RECEIVED_FROM_WS",
		slog.String("eid", eid.String()),
		slog.String("cid", cid.String()))

	if err := o.tracker.Ack(ctx, eid, cid, o.ackTimeout); err != nil {
		return err
	}

	// [STATUS_REPORT] Resolve the envelope into a per-recipient delivered
	// status for im-thread-service (batched, best-effort).
	if o.confirmer != nil {
		o.confirmer.ConfirmDelivered(ctx, eid, viaWebSocket)
	}

	return nil
}

// [ENQUEUE] Non-blocking hand-off to the internal worker pool.
func (o *EventOrchestrator) enqueue(ctx context.Context, ev event.Eventer, isDismiss bool) {
	select {
	case o.queue <- task{ctx: context.WithoutCancel(ctx), ev: ev, dismissable: isDismiss}:
	default:
		o.log.Warn("ORCHESTRATOR_PIPE_CLOGGED", slog.String("eid", ev.GetID()))
	}
}

// [BOOTSTRAP] Warms up the processing engine.
func (o *EventOrchestrator) bootstrap(n int) {
	for i := range n {
		o.wg.Add(1)
		go func(id int) {
			defer o.wg.Done()

			for t := range o.queue {
				o.consume(t, id)
			}
		}(i)
	}
}

// [CONSUME] Dispatches tasks to registered handlers with individual safety sandboxes.
func (o *EventOrchestrator) consume(t task, workerID int) {
	for _, h := range o.handlers {
		func() {
			defer func() {
				if r := recover(); r != nil {
					o.log.Error("HANDLER_CRASHED", slog.Int("worker", workerID), slog.Any("panic", r))
				}
			}()

			if t.dismissable {
				if rh, ok := h.(DismissHandler); ok {
					rh.HandleDismiss(t.ctx, t.ev)
				}

				return
			}

			h.Handle(t.ctx, t.ev)
		}()
	}
}

// [CLOSE] Implements graceful shutdown logic.
func (o *EventOrchestrator) Close() error {
	o.log.Info("ORCHESTRATOR_DRAINING_QUEUE")
	close(o.queue)
	o.wg.Wait()

	return nil
}
