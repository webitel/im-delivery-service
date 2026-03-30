// internal/service/push_handler.go
package service

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sony/gobreaker"
	"github.com/webitel/im-delivery-service/config"
	leader "github.com/webitel/im-delivery-service/infra/discovery/consul"
	"github.com/webitel/im-delivery-service/internal/domain/event"
	"github.com/webitel/im-delivery-service/internal/domain/model"
	"github.com/webitel/im-delivery-service/internal/store"
)

// [INTERFACE_GUARD]
var _ Pusher = (*PushHandler)(nil)

const (
	pollingInterval = 1 * time.Second
	deliveryTimeout = 15 * time.Second
)

type PushProvider interface {
	Send(ctx context.Context, req *model.PushRequest) error
	Dismiss(ctx context.Context, req *model.PushRequest) error
}

// [PUSHER] Defines the high-level life-cycle and event handling for push notifications.
type Pusher interface {
	Start(ctx context.Context)
	Handle(ctx context.Context, ev event.Eventer)
	HandleDismiss(ctx context.Context, ev event.Eventer)
}

type PushHandler struct {
	tracker        store.DeliveryTracker
	scheduler      store.DeliveryScheduler
	deviceProvider DeviceProvider
	presenceStore  store.PresenceStore
	pusher         PushProvider
	leader         leader.LeaderAwarer
	cb             *gobreaker.CircuitBreaker
	log            *slog.Logger
	timeout        time.Duration
	wg             sync.WaitGroup
}

func NewPushHandler(
	tracker store.DeliveryTracker,
	scheduler store.DeliveryScheduler,
	deviceProvider DeviceProvider,
	presenceStore store.PresenceStore,
	pusher PushProvider,
	leader leader.LeaderAwarer,
	cfg *config.Config,
	log *slog.Logger,
) *PushHandler {
	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        "push-gateway",
		MaxRequests: 5,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.Requests >= 10 && float64(counts.TotalFailures)/float64(counts.Requests) >= 0.5
		},
	})

	return &PushHandler{
		tracker:        tracker,
		scheduler:      scheduler,
		deviceProvider: deviceProvider,
		presenceStore:  presenceStore,
		pusher:         pusher,
		leader:         leader,
		cb:             cb,
		log:            log.With("component", "push_handler"),
		timeout:        cfg.Delivery.AckTimeout,
	}
}

// [START] Runs the polling loop for scheduled tasks.
func (h *PushHandler) Start(ctx context.Context) {
	ticker := time.NewTicker(pollingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			h.wg.Wait()
			return
		case <-ticker.C:
			if h.leader.IsLeader() {
				h.processPendingTasks(ctx)
			}
		}
	}
}

// [PROCESS_PENDING_TASKS] Unified flow: Fetch -> Dispatch.
func (h *PushHandler) processPendingTasks(ctx context.Context) {
	// [UNIFIED_FETCH] PullReady now returns full events and cleans Redis.
	events, err := h.scheduler.PullReady(ctx)
	if err != nil {
		h.log.Error("PULL_FAILED", slog.Any("err", err))
		return
	}

	for _, ev := range events {
		h.wg.Add(1)
		go func(e event.Eventer) {
			defer h.wg.Done()
			h.dispatch(e)
		}(ev)
	}
}

// [HANDLE] Direct schedule with full event persistence.
func (h *PushHandler) Handle(ctx context.Context, ev event.Eventer) {
	if ev.IsEcho() {
		return
	}

	if trackable, ok := ev.(event.IsPushable); ok && trackable.IsPushable() {
		h.log.Debug("SCHEDULING_PUSH_CHECK",
			slog.String("eid", ev.GetID()),
			slog.Duration("delay", h.timeout))
		// Event will live in Redis until PullReady or 24h expiration.
		_ = h.scheduler.Schedule(context.WithoutCancel(ctx), ev, h.timeout)
	}
}

// [DISPATCH] Orchestrates the push delivery logic.
func (h *PushHandler) dispatch(ev event.Eventer) {
	ctx, cancel := context.WithTimeout(context.Background(), deliveryTimeout)
	defer cancel()

	eid, _ := uuid.Parse(ev.GetID())
	uid := ev.GetUserID()

	// [ACK_CHECK] Skip if user already read the message via WebSocket.
	acked, _ := h.tracker.GetAckedSessions(ctx, eid)

	if len(acked) > 0 {
		h.log.Info("PUSH_CANCELLED_BY_ACK",
			slog.String("eid", eid.String()),
			slog.Int("sessions_count", len(acked)))
		_ = h.tracker.Remove(ctx, eid)
		return
	}

	// [RESOLVE] Find all registered push tokens for the user.
	devices, err := h.deviceProvider.GetDevices(ctx, uid)
	if err != nil || len(devices) == 0 {
		_ = h.tracker.Remove(ctx, eid)
		return
	}

	// [FILTER] Remove devices that have active/acked WebSocket sessions.
	targets := h.filter(ctx, uid, devices, acked)
	if len(targets) > 0 {
		h.ship(ctx, ev, targets)
	}

	// [CLEANUP] Finalize task.
	_ = h.tracker.Remove(ctx, eid)
}

// [SHIP] Executes individual push sends through the circuit breaker.
func (h *PushHandler) ship(ctx context.Context, ev event.Eventer, targets []model.Device) {
	for _, target := range targets {
		req := &model.PushRequest{
			Devices: []model.Device{target},
		}

		// [POLYMORPHIC_MAPPING] Automatically extracts Title/Body from the event.
		req.FillFromEvent(ev)

		// [FALLBACK] Ensure visual pushes have a default title.
		if req.Title == "" && !req.IsSilent {
			req.Title = "New Message"
		}

		_, err := h.cb.Execute(func() (any, error) {
			return nil, h.pusher.Send(ctx, req)
		})
		if err != nil {
			h.log.Error("SHIP_FAILED",
				slog.String("device", target.ID),
				slog.String("eid", ev.GetID()),
				slog.Any("err", err),
			)
		}
	}
}

// [FILTER_PUSH_TARGETS] Compares acked session devices with target push devices.
func (h *PushHandler) filter(ctx context.Context, uid uuid.UUID, devices []model.Device, acked []uuid.UUID) []model.Device {
	if len(acked) == 0 {
		return devices
	}

	ackedMap := make(map[string]struct{})
	for _, cid := range acked {
		if devID, err := h.presenceStore.GetSessionDevice(ctx, uid, cid); err == nil && devID != "" {
			ackedMap[devID] = struct{}{}
		}
	}

	filtered := make([]model.Device, 0)
	for _, d := range devices {
		if _, seen := ackedMap[d.ID]; !seen {
			filtered = append(filtered, d)
		}
	}
	return filtered
}

// [HANDLE_DISMISS] Revokes existing notifications across all user devices.
func (h *PushHandler) HandleDismiss(ctx context.Context, ev event.Eventer) {
	if !h.leader.IsLeader() {
		return
	}

	devices, err := h.deviceProvider.GetDevices(ctx, ev.GetUserID())
	if err != nil || len(devices) == 0 {
		return
	}

	req := &model.PushRequest{
		Devices:  devices,
		IsSilent: true,
		Data: map[string]string{
			"action":   "DISMISS",
			"event_id": ev.GetID(),
		},
	}

	// [METADATA] Ensure IDs are populated correctly.
	req.FillFromEvent(ev)

	_, _ = h.cb.Execute(func() (any, error) {
		return nil, h.pusher.Dismiss(ctx, req)
	})
}
