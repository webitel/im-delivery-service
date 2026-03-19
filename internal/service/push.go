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

// [INTERFACE GUARD]
var _ Pusher = (*PushHandler)(nil)

const (
	pollingInterval = 1 * time.Second
	deliveryTimeout = 15 * time.Second
)

type PushProvider interface {
	Send(ctx context.Context, req *model.PushRequest) error
	Dismiss(ctx context.Context, req *model.PushRequest) error
}

// Pusher defines the high-level life-cycle and event handling for push notifications.
type Pusher interface {
	Start(ctx context.Context)
	Handle(ctx context.Context, ev event.Eventer)
	HandleDismiss(ctx context.Context, ev event.Eventer)
}

type PushHandler struct {
	tracker   store.DeliveryTracker
	scheduler store.DeliveryScheduler
	resolver  *DeviceResolver
	pusher    PushProvider
	leader    leader.LeaderAwarer
	cb        *gobreaker.CircuitBreaker
	log       *slog.Logger
	timeout   time.Duration
	wg        sync.WaitGroup
}

func NewPushHandler(
	tracker store.DeliveryTracker,
	scheduler store.DeliveryScheduler,
	resolver *DeviceResolver,
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
		tracker:   tracker,
		scheduler: scheduler,
		resolver:  resolver,
		pusher:    pusher,
		leader:    leader,
		cb:        cb,
		log:       log.With("component", "push_handler"),
		timeout:   cfg.Delivery.AckTimeout,
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

func (h *PushHandler) processPendingTasks(ctx context.Context) {
	tasks, err := h.scheduler.PullReady(ctx)
	if err != nil {
		h.log.Error("PULL_FAILED", slog.Any("err", err))
		return
	}

	for _, task := range tasks {
		h.wg.Add(1)
		go func(t store.ScheduledTask) {
			defer h.wg.Done()
			h.dispatch(t.EventID, t.UserID)
		}(task)
	}
}

// [HANDLE] Schedules a new push event if the event is pushable.
func (h *PushHandler) Handle(ctx context.Context, ev event.Eventer) {
	if ev.IsEcho() {
		return
	}

	if trackable, ok := ev.(event.IsPushable); ok && trackable.CanPush() {
		eid, _ := uuid.Parse(ev.GetID())
		_ = h.scheduler.Schedule(context.WithoutCancel(ctx), eid, ev.GetUserID(), h.timeout)
	}
}

func (h *PushHandler) dispatch(eid, uid uuid.UUID) {
	ctx, cancel := context.WithTimeout(context.Background(), deliveryTimeout)
	defer cancel()

	// 1. Check who already received message via WebSocket
	acked, _ := h.tracker.GetAckedSessions(ctx, eid)

	// 2. Resolve target devices using our new service
	devices, err := h.resolver.GetDevices(ctx, uid)
	if err != nil || len(devices) == 0 {
		_ = h.tracker.Remove(ctx, eid)
		return
	}

	// 3. Filter out devices that already have active/acked sessions
	targets := h.filterPushTargets(ctx, uid, devices, acked)
	if len(targets) > 0 {
		h.ship(ctx, eid, uid, targets)
	}

	_ = h.tracker.Remove(ctx, eid)
}

func (h *PushHandler) ship(ctx context.Context, eid, uid uuid.UUID, targets []model.Device) {
	for _, target := range targets {
		req := &model.PushRequest{
			UserID:     uid.String(),
			Devices:    []model.Device{target},
			Title:      "New Message",
			CollapseID: eid.String(),
		}

		_, err := h.cb.Execute(func() (any, error) {
			return nil, h.pusher.Send(ctx, req)
		})
		if err != nil {
			h.log.Error("SHIP_FAILED", slog.String("device", target.ID), slog.Any("err", err))
		}
	}
}

func (h *PushHandler) filterPushTargets(ctx context.Context, uid uuid.UUID, devices []model.Device, acked []uuid.UUID) []model.Device {
	if len(acked) == 0 {
		return devices
	}

	ackedMap := make(map[string]struct{})
	for _, cid := range acked {
		if devID, err := h.resolver.presence.GetSessionDevice(ctx, uid, cid); err == nil && devID != "" {
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

// [DISMISS] Sends silent push to clear notifications on other devices.
func (h *PushHandler) HandleDismiss(ctx context.Context, ev event.Eventer) {
	if !h.leader.IsLeader() {
		return
	}

	uid, eid := ev.GetUserID(), ev.GetID()
	devices, err := h.resolver.GetDevices(ctx, uid)
	if err != nil || len(devices) == 0 {
		return
	}

	req := &model.PushRequest{
		UserID:     uid.String(),
		Devices:    devices,
		CollapseID: eid,
		IsSilent:   true,
		Data:       map[string]string{"action": "DISMISS", "event_id": eid},
	}
	_, _ = h.cb.Execute(func() (any, error) {
		return nil, h.pusher.Dismiss(ctx, req)
	})
}

// // internal/service/push_handler.go
// package service

// import (
// 	"context"
// 	"encoding/json"
// 	"log/slog"
// 	"sync"
// 	"time"

// 	"github.com/google/uuid"
// 	"github.com/sony/gobreaker"
// 	"github.com/webitel/im-delivery-service/config"
// 	adminv1 "github.com/webitel/im-delivery-service/gen/go/admin/v1"
// 	authv1 "github.com/webitel/im-delivery-service/gen/go/auth/v1"
// 	imadmin "github.com/webitel/im-delivery-service/infra/client/im-admin"
// 	imauth "github.com/webitel/im-delivery-service/infra/client/im-auth"
// 	leader "github.com/webitel/im-delivery-service/infra/discovery/consul"
// 	"github.com/webitel/im-delivery-service/internal/domain/event"
// 	"github.com/webitel/im-delivery-service/internal/domain/model"
// 	"github.com/webitel/im-delivery-service/internal/store"
// 	"google.golang.org/protobuf/types/known/wrapperspb"
// )

// const (
// 	ProviderFCM  = "fcm"
// 	ProviderAPNS = "apn"
// 	ProviderWeb  = "web"

// 	// [INTERVAL] Frequency of pulling due tasks from the persistent scheduler.
// 	pollingInterval = 1 * time.Second
// 	// [TIMEOUT] Maximum duration for a single delivery attempt (external network call).
// 	deliveryTimeout = 15 * time.Second
// )

// // [GATEWAY] Interface for interacting with external notification providers.
// type PushProvider interface {
// 	Send(ctx context.Context, req *model.PushRequest) error
// 	Dismiss(ctx context.Context, req *model.PushRequest) error
// }

// // [INTERFACE] High-level contract for the push notification sub-system.
// type Pusher interface {
// 	Start(ctx context.Context)
// 	Handle(ctx context.Context, ev event.Eventer)
// }

// // [PUSH_HANDLER] Core service for deferred and reliable push notification delivery.
// type PushHandler struct {
// 	tracker   store.DeliveryTracker
// 	presence  store.PresenceStore
// 	scheduler store.DeliveryScheduler
// 	pusher    PushProvider
// 	leader    leader.LeaderAwarer
// 	imauth    *imauth.Client
// 	imadmin   *imadmin.Client
// 	cb        *gobreaker.CircuitBreaker
// 	log       *slog.Logger
// 	timeout   time.Duration
// 	wg        sync.WaitGroup
// }

// // [INTERFACE_GUARDS] Verification of interface compliance.
// var (
// 	_ EventHandler   = (*PushHandler)(nil)
// 	_ DismissHandler = (*PushHandler)(nil)
// 	_ Pusher         = (*PushHandler)(nil)
// )

// func NewPushHandler(
// 	tracker store.DeliveryTracker,
// 	presence store.PresenceStore,
// 	scheduler store.DeliveryScheduler,
// 	pusher PushProvider,
// 	leader leader.LeaderAwarer,
// 	imauth *imauth.Client,
// 	imadmin *imadmin.Client,
// 	cfg *config.Config,
// 	log *slog.Logger,
// ) *PushHandler {
// 	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
// 		Name:        "push-gateway",
// 		MaxRequests: 5,
// 		Interval:    10 * time.Second,
// 		Timeout:     30 * time.Second,
// 		ReadyToTrip: func(counts gobreaker.Counts) bool {
// 			return counts.Requests >= 10 && float64(counts.TotalFailures)/float64(counts.Requests) >= 0.5
// 		},
// 	})

// 	return &PushHandler{
// 		tracker:   tracker,
// 		presence:  presence,
// 		scheduler: scheduler,
// 		pusher:    pusher,
// 		leader:    leader,
// 		imauth:    imauth,
// 		imadmin:   imadmin,
// 		cb:        cb,
// 		log:       log.With("component", "push_handler"),
// 		timeout:   cfg.Delivery.AckTimeout,
// 	}
// }

// // [LIFECYCLE] Initiates the polling loop.
// func (h *PushHandler) Start(ctx context.Context) {
// 	h.log.Info("PUSH_WORKER_STARTED", slog.Duration("interval", pollingInterval))
// 	ticker := time.NewTicker(pollingInterval)
// 	defer ticker.Stop()

// 	for {
// 		select {
// 		case <-ctx.Done():
// 			h.log.Info("PUSH_WORKER_SHUTTING_DOWN")
// 			h.wg.Wait()
// 			return
// 		case <-ticker.C:
// 			if !h.leader.IsLeader() {
// 				continue
// 			}
// 			h.processPendingTasks(ctx)
// 		}
// 	}
// }

// func (h *PushHandler) processPendingTasks(ctx context.Context) {
// 	tasks, err := h.scheduler.PullReady(ctx)
// 	if err != nil {
// 		h.log.Error("SCHEDULER_PULL_FAILED", slog.Any("err", err))
// 		return
// 	}

// 	for _, task := range tasks {
// 		h.wg.Add(1)
// 		go func(t store.ScheduledTask) {
// 			defer h.wg.Done()
// 			h.dispatch(t.EventID, t.UserID)
// 		}(task)
// 	}
// }

// func (h *PushHandler) Handle(ctx context.Context, ev event.Eventer) {
// 	if ev.IsEcho() {
// 		return
// 	}

// 	trackable, ok := ev.(event.IsPushable)
// 	if !ok || !trackable.CanPush() {
// 		return
// 	}

// 	eid, _ := uuid.Parse(ev.GetID())
// 	if err := h.scheduler.Schedule(context.WithoutCancel(ctx), eid, ev.GetUserID(), h.timeout); err != nil {
// 		h.log.Error("SCHEDULING_FAILED", slog.Any("err", err))
// 	}
// }

// func (h *PushHandler) dispatch(eid, uid uuid.UUID) {
// 	ctx, cancel := context.WithTimeout(context.Background(), deliveryTimeout)
// 	defer cancel()

// 	ackedSessions, _ := h.tracker.GetAckedSessions(ctx, eid)

// 	devices, err := h.resolveDevices(ctx, uid)
// 	if err != nil || len(devices) == 0 {
// 		_ = h.tracker.Remove(ctx, eid)
// 		return
// 	}

// 	targets := h.filterPushTargets(ctx, uid, devices, ackedSessions)
// 	if len(targets) == 0 {
// 		_ = h.tracker.Remove(ctx, eid)
// 		return
// 	}

// 	h.ship(ctx, eid, uid, targets)
// 	_ = h.tracker.Remove(ctx, eid)
// }

// // [RESOLVE] Entry point for device lookup with cache-aside pattern.
// func (h *PushHandler) resolveDevices(ctx context.Context, uid uuid.UUID) ([]model.Device, error) {
// 	if cached, err := h.presence.UserDevices(ctx, uid); err == nil && cached != nil {
// 		return *cached, nil
// 	}

// 	resp, err := h.imauth.GetAuthorizations(ctx, &authv1.GetAuthorizationRequest{
// 		Contact: &authv1.InputContact{Input: &authv1.InputContact_Id{Id: uid.String()}},
// 		Push:    &wrapperspb.BoolValue{Value: true},
// 	})
// 	if err != nil {
// 		return nil, err
// 	}

// 	devices := h.mapAuthToDevices(resp)
// 	if len(devices) == 0 {
// 		return devices, nil
// 	}

// 	// [ENRICHMENT] Fetch app configurations in a loop as API only supports single ID.
// 	enriched, err := h.enrichWithAppConfigs(ctx, devices)
// 	if err != nil {
// 		h.log.Warn("APP_ENRICHMENT_FAILED", slog.Any("err", err))
// 	}

// 	_ = h.presence.SyncDevices(ctx, uid, enriched)
// 	return enriched, nil
// }

// // [ENRICH] Loops through unique app IDs to fetch configurations one by one.
// func (h *PushHandler) enrichWithAppConfigs(ctx context.Context, devices []model.Device) ([]model.Device, error) {
// 	// [COLLECT] Identify unique application IDs from the device list.
// 	appConfigs := make(map[string]*adminv1.Application)
// 	for _, d := range devices {
// 		if d.AppName == "" {
// 			continue
// 		}
// 		if _, seen := appConfigs[d.AppName]; seen {
// 			continue
// 		}

// 		// [REMOTE_CALL] Single ID lookup as per API constraints.
// 		res, err := h.imadmin.SearchApps(ctx, &adminv1.SearchAppRequest{
// 			Id: d.AppName,
// 		})
// 		if err != nil {
// 			h.log.Warn("SEARCH_APP_FAILED", slog.String("app_id", d.AppName), slog.Any("err", err))
// 			continue
// 		}

// 		// [MATCH] Check if we got a valid response for the requested ID.
// 		if res != nil && len(res.Data) > 0 {
// 			appConfigs[d.AppName] = res.Data[0]
// 		}
// 	}

// 	// [MAP_DEVICES] Apply fetched configurations to each device.
// 	for i := range devices {
// 		app, ok := appConfigs[devices[i].AppName]
// 		if !ok || app.Service == nil || app.Service.PushService == nil {
// 			continue
// 		}

// 		ps := app.Service.PushService

// 		switch devices[i].PushType {
// 		case ProviderFCM:
// 			if ps.Fcm != nil {
// 				devices[i].PushConfig.Proxy = ps.Fcm.Proxy
// 				devices[i].PushConfig.Credentials = ps.Fcm.Account
// 			}
// 		case ProviderAPNS:
// 			if ps.Apn != nil {
// 				devices[i].PushConfig.Proxy = ps.Apn.Proxy
// 				devices[i].PushConfig.Topic = ps.Apn.Topic
// 				if ps.Apn.Token != nil {
// 					devices[i].PushConfig.Credentials = ps.Apn.Token.AuthKey
// 				}
// 			}
// 		case ProviderWeb:
// 			if ps.Web != nil {
// 				devices[i].PushConfig.Proxy = ps.Web.Proxy
// 				devices[i].PushConfig.Credentials = ps.Web.Token
// 			}
// 		}
// 	}
// 	return devices, nil
// }

// // [SHIP] Executes actual network calls via circuit breaker.
// func (h *PushHandler) ship(ctx context.Context, eid, uid uuid.UUID, targets []model.Device) {
// 	for _, target := range targets {
// 		req := &model.PushRequest{
// 			UserID:     uid.String(),
// 			Devices:    []model.Device{target},
// 			Title:      "New Message",
// 			Body:       "You have a new notification",
// 			CollapseID: eid.String(),
// 		}

// 		_, err := h.cb.Execute(func() (any, error) {
// 			return nil, h.pusher.Send(ctx, req)
// 		})
// 		if err != nil {
// 			h.log.Error("PUSH_DISPATCH_FAILED", slog.String("app", target.AppName), slog.Any("err", err))
// 		}
// 	}
// }

// func (h *PushHandler) mapAuthToDevices(resp *authv1.AuthorizationList) []model.Device {
// 	res := make([]model.Device, 0)
// 	if resp == nil {
// 		return res
// 	}

// 	for _, a := range resp.Data {
// 		if a.Device == nil || a.Device.Push == nil {
// 			continue
// 		}

// 		d := model.Device{ID: a.Device.Id, AppName: a.AppId}
// 		switch t := a.Device.Push.Token.(type) {
// 		case *authv1.PUSHSubscription_Fcm:
// 			d.PushType, d.PushToken, d.Platform = ProviderFCM, t.Fcm, model.PlatformAndroid
// 		case *authv1.PUSHSubscription_Apn:
// 			d.PushType, d.PushToken, d.Platform = ProviderAPNS, t.Apn, model.PlatformIOS
// 		case *authv1.PUSHSubscription_Web:
// 			d.PushType, d.Platform = ProviderWeb, model.PlatformWeb
// 			if b, err := json.Marshal(t.Web); err == nil {
// 				d.PushToken = string(b)
// 			}
// 		}

// 		if d.PushToken != "" {
// 			res = append(res, d)
// 		}
// 	}
// 	return res
// }

// func (h *PushHandler) filterPushTargets(ctx context.Context, uid uuid.UUID, devices []model.Device, acked []uuid.UUID) []model.Device {
// 	if len(acked) == 0 {
// 		return devices
// 	}
// 	ackedMap := make(map[string]struct{})
// 	for _, cid := range acked {
// 		if devID, err := h.presence.GetSessionDevice(ctx, uid, cid); err == nil && devID != "" {
// 			ackedMap[devID] = struct{}{}
// 		}
// 	}
// 	filtered := make([]model.Device, 0)
// 	for _, d := range devices {
// 		if _, seen := ackedMap[d.ID]; !seen {
// 			filtered = append(filtered, d)
// 		}
// 	}
// 	return filtered
// }

// func (h *PushHandler) HandleDismiss(ctx context.Context, ev event.Eventer) {
// 	if !h.leader.IsLeader() {
// 		return
// 	}
// 	uid, eid := ev.GetUserID(), ev.GetID()
// 	devices, err := h.presence.UserDevices(ctx, uid)
// 	if err != nil || devices == nil || len(*devices) == 0 {
// 		return
// 	}
// 	req := &model.PushRequest{
// 		UserID:     uid.String(),
// 		Devices:    *devices,
// 		CollapseID: eid,
// 		IsSilent:   true,
// 		Data:       map[string]string{"action": "DISMISS", "event_id": eid},
// 	}
// 	_, _ = h.cb.Execute(func() (any, error) {
// 		return nil, h.pusher.Dismiss(ctx, req)
// 	})
// }

// // internal/service/push_handler.go
// package service

// import (
// 	"context"
// 	"encoding/json"
// 	"fmt"
// 	"log/slog"
// 	"sync"
// 	"time"

// 	"github.com/google/uuid"
// 	"github.com/sony/gobreaker"
// 	"github.com/webitel/im-delivery-service/config"
// 	authv1 "github.com/webitel/im-delivery-service/gen/go/auth/v1"
// 	imauth "github.com/webitel/im-delivery-service/infra/client/im-auth"
// 	leader "github.com/webitel/im-delivery-service/infra/discovery/consul"
// 	"github.com/webitel/im-delivery-service/internal/domain/event"
// 	"github.com/webitel/im-delivery-service/internal/domain/model"
// 	"github.com/webitel/im-delivery-service/internal/store"
// 	"google.golang.org/protobuf/types/known/wrapperspb"
// )

// const (
// 	ProviderFCM  = "fcm"
// 	ProviderAPNS = "apn"
// 	ProviderWeb  = "web"

// 	// [INTERVAL] Frequency of pulling due tasks from the persistent scheduler.
// 	pollingInterval = 1 * time.Second
// 	// [TIMEOUT] Maximum duration for a single delivery attempt (external network call).
// 	deliveryTimeout = 15 * time.Second
// )

// // [GATEWAY] Interface for interacting with external notification providers (FCM, APNs).
// type PushProvider interface {
// 	Send(ctx context.Context, req *model.PushRequest) error
// 	Dismiss(ctx context.Context, req *model.PushRequest) error
// }

// // [INTERFACE] High-level contract for the push notification sub-system.
// type Pusher interface {
// 	Start(ctx context.Context)
// 	Handle(ctx context.Context, ev event.Eventer)
// }

// // [PUSH_HANDLER] Core service for deferred and reliable push notification delivery.
// type PushHandler struct {
// 	tracker   store.DeliveryTracker
// 	presence  store.PresenceStore
// 	scheduler store.DeliveryScheduler
// 	pusher    PushProvider
// 	leader    leader.LeaderAwarer
// 	auth      *imauth.Client
// 	cb        *gobreaker.CircuitBreaker
// 	log       *slog.Logger
// 	timeout   time.Duration
// 	wg        sync.WaitGroup
// }

// // [INTERFACE_GUARDS] Verification of interface compliance.
// var (
// 	_ EventHandler   = (*PushHandler)(nil)
// 	_ DismissHandler = (*PushHandler)(nil)
// 	_ Pusher         = (*PushHandler)(nil)
// )

// func NewPushHandler(
// 	tracker store.DeliveryTracker,
// 	presence store.PresenceStore,
// 	scheduler store.DeliveryScheduler,
// 	pusher PushProvider,
// 	leader leader.LeaderAwarer,
// 	auth *imauth.Client,
// 	cfg *config.Config,
// 	log *slog.Logger,
// ) *PushHandler {
// 	// [CIRCUIT_BREAKER] Fail-fast mechanism to protect the system from gateway degradation.
// 	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
// 		Name:        "push-gateway",
// 		MaxRequests: 5,
// 		Interval:    10 * time.Second,
// 		Timeout:     30 * time.Second,
// 		ReadyToTrip: func(counts gobreaker.Counts) bool {
// 			return counts.Requests >= 10 && float64(counts.TotalFailures)/float64(counts.Requests) >= 0.5
// 		},
// 	})

// 	return &PushHandler{
// 		tracker:   tracker,
// 		presence:  presence,
// 		scheduler: scheduler,
// 		pusher:    pusher,
// 		leader:    leader,
// 		auth:      auth,
// 		cb:        cb,
// 		log:       log.With("component", "push_handler"),
// 		timeout:   cfg.Delivery.AckTimeout,
// 	}
// }

// // [LIFECYCLE] Initiates the polling loop. Only the Leader node executes tasks.
// func (h *PushHandler) Start(ctx context.Context) {
// 	h.log.Info("PUSH_WORKER_STARTED", slog.Duration("interval", pollingInterval))
// 	ticker := time.NewTicker(pollingInterval)
// 	defer ticker.Stop()

// 	for {
// 		select {
// 		case <-ctx.Done():
// 			h.log.Info("PUSH_WORKER_SHUTTING_DOWN")
// 			h.wg.Wait() // [GRACEFUL] Ensures mid-flight requests finish correctly.
// 			return
// 		case <-ticker.C:
// 			// [LEADER_ONLY] Restricts processing to a single node to avoid cluster contention.
// 			if !h.leader.IsLeader() {
// 				continue
// 			}
// 			h.processPendingTasks(ctx)
// 		}
// 	}
// }

// // [POLLING] Fetches ready-to-deliver tasks and executes them in parallel.
// func (h *PushHandler) processPendingTasks(ctx context.Context) {
// 	tasks, err := h.scheduler.PullReady(ctx)
// 	if err != nil {
// 		h.log.Error("SCHEDULER_PULL_FAILED", slog.Any("err", err))
// 		return
// 	}

// 	if len(tasks) > 0 {
// 		h.log.Info("TASKS_FETCHED", slog.Int("count", len(tasks)))
// 	}

// 	for _, task := range tasks {
// 		h.wg.Add(1)
// 		go func(t store.ScheduledTask) {
// 			defer h.wg.Done()
// 			h.log.Debug("DISPATCHING_TASK", slog.String("eid", t.EventID.String()), slog.String("uid", t.UserID.String()))
// 			h.dispatch(t.EventID, t.UserID)
// 		}(task)
// 	}
// }

// // [HANDLE] Processes event for push delivery.
// func (h *PushHandler) Handle(ctx context.Context, ev event.Eventer) {
// 	// [ECHO_CHECK] Skip push notifications if the event is just an echo sync for the sender.
// 	if ev.IsEcho() {
// 		h.log.Debug("SKIP_PUSH_FOR_ECHO", slog.String("eid", ev.GetID()))
// 		return
// 	}

// 	// [TRACKABLE_CHECK] Ensure event requires persistent delivery.
// 	trackable, ok := ev.(event.IsPushable)
// 	if !ok || !trackable.CanPush() {
// 		return
// 	}

// 	eid, _ := uuid.Parse(ev.GetID())

// 	// [SCHEDULING] Hand off to background delivery loop.
// 	if err := h.scheduler.Schedule(context.WithoutCancel(ctx), eid, ev.GetUserID(), h.timeout); err != nil {
// 		h.log.Error("SCHEDULING_FAILED", slog.Any("err", err), slog.String("eid", eid.String()))
// 	}
// }

// // [DISPATCH] The core execution logic: Late-ACK check and transmission.
// func (h *PushHandler) dispatch(eid, uid uuid.UUID) {
// 	ctx, cancel := context.WithTimeout(context.Background(), deliveryTimeout)
// 	defer cancel()

// 	// 1. [LATE_ACK_CHECK] Skip push if user has already acknowledged the event.
// 	ackedSessions, err := h.tracker.GetAckedSessions(ctx, eid)
// 	if err != nil {
// 		h.log.Warn("ACK_CHECK_FAILED", slog.Any("err", err), slog.String("eid", eid.String()))
// 	}

// 	// 2. [RESOLVE] Get push-enabled targets.
// 	devices, err := h.resolveDevices(ctx, uid)
// 	if err != nil || len(devices) == 0 {
// 		if err != nil {
// 			h.log.Debug("RESOLVE_DEVICES_FAILED", slog.Any("err", err), slog.String("uid", uid.String()))
// 		}
// 		_ = h.tracker.Remove(ctx, eid)
// 		return
// 	}

// 	// 3. [FILTER] Remove devices that belong to already acknowledged sessions.
// 	targets := h.filterPushTargets(ctx, uid, devices, ackedSessions)
// 	if len(targets) == 0 {
// 		h.log.Debug("NO_TARGETS_AFTER_FILTER", slog.String("eid", eid.String()))
// 		_ = h.tracker.Remove(ctx, eid)
// 		return
// 	}

// 	// 4. [EXECUTE] Ship the payload through the circuit breaker.
// 	h.ship(ctx, eid, uid, targets)

// 	// 5. [CLEANUP] Remove tracking state after final delivery attempt.
// 	_ = h.tracker.Remove(ctx, eid)
// }

// // [SHIP] Encapsulates the actual network call to the push gateway.
// func (h *PushHandler) ship(ctx context.Context, eid, uid uuid.UUID, targets []model.Device) {
// 	req := &model.PushRequest{
// 		UserID:     uid.String(),
// 		Devices:    targets,
// 		Title:      "New Message",
// 		Body:       "You have a new notification",
// 		CollapseID: eid.String(), // [IDEMPOTENCY] Prevents duplicate notifications on handset.
// 	}

// 	_, err := h.cb.Execute(func() (any, error) {
// 		return nil, h.pusher.Send(ctx, req)
// 	})
// 	if err != nil {
// 		h.log.Error("GATEWAY_DELIVERY_FAILED", slog.Any("err", err), slog.String("eid", eid.String()))
// 	} else {
// 		h.log.Info("PUSH_DELIVERED", slog.String("eid", eid.String()), slog.Int("targets", len(targets)))
// 	}
// }

// // [HANDLE_DISMISS] Immediate silent push to revoke notification UI on all devices.
// func (h *PushHandler) HandleDismiss(ctx context.Context, ev event.Eventer) {
// 	// [LEADER_ONLY] Only the active leader handles global revocation.
// 	if !h.leader.IsLeader() {
// 		return
// 	}

// 	uid := ev.GetUserID()
// 	eid := ev.GetID()

// 	devices, err := h.presence.UserDevices(ctx, uid)
// 	if err != nil || devices == nil || len(*devices) == 0 {
// 		return
// 	}

// 	req := &model.PushRequest{
// 		UserID:     uid.String(),
// 		Devices:    *devices,
// 		CollapseID: eid,
// 		IsSilent:   true,
// 		Data:       map[string]string{"action": "DISMISS", "event_id": eid},
// 	}

// 	_, err = h.cb.Execute(func() (any, error) {
// 		return nil, h.pusher.Dismiss(ctx, req)
// 	})
// 	if err != nil {
// 		h.log.Warn("DISMISS_FAILED", slog.Any("err", err), slog.String("eid", eid))
// 	}
// }

// // [FILTER] Logic to exclude already active/acked device IDs.
// func (h *PushHandler) filterPushTargets(ctx context.Context, uid uuid.UUID, devices []model.Device, acked []uuid.UUID) []model.Device {
// 	if len(acked) == 0 {
// 		return devices
// 	}

// 	ackedMap := make(map[string]struct{})
// 	for _, cid := range acked {
// 		if devID, err := h.presence.GetSessionDevice(ctx, uid, cid); err == nil && devID != "" {
// 			ackedMap[devID] = struct{}{}
// 		}
// 	}

// 	filtered := make([]model.Device, 0)
// 	for _, d := range devices {
// 		if _, seen := ackedMap[d.ID]; !seen {
// 			filtered = append(filtered, d)
// 		}
// 	}
// 	return filtered
// }

// // [RESOLVE] Optimized device lookup with local presence cache fallback to auth service.
// func (h *PushHandler) resolveDevices(ctx context.Context, uid uuid.UUID) ([]model.Device, error) {
// 	if cached, err := h.presence.UserDevices(ctx, uid); err == nil && cached != nil {
// 		return *cached, nil
// 	}

// 	userID := uid.String()
// 	fmt.Println(userID)

// 	resp, err := h.auth.GetAuthorizations(ctx, &authv1.GetAuthorizationRequest{
// 		Contact: &authv1.InputContact{Input: &authv1.InputContact_Id{Id: uid.String()}},
// 		Push:    &wrapperspb.BoolValue{Value: true},
// 	})
// 	if err != nil {
// 		return nil, err
// 	}

// 	resolved := h.mapAuthToDevices(resp)
// 	_ = h.presence.SyncDevices(ctx, uid, resolved)
// 	return resolved, nil
// }

// func (h *PushHandler) mapAuthToDevices(resp *authv1.AuthorizationList) []model.Device {
// 	res := make([]model.Device, 0)
// 	if resp == nil {
// 		return res
// 	}

// 	for _, a := range resp.Data {
// 		if a.Device == nil || a.Device.Push == nil {
// 			continue
// 		}

// 		d := model.Device{ID: a.Device.Id, AppName: a.AppId}
// 		switch t := a.Device.Push.Token.(type) {
// 		case *authv1.PUSHSubscription_Fcm:
// 			d.PushType, d.PushToken, d.Platform = ProviderFCM, t.Fcm, model.PlatformAndroid
// 		case *authv1.PUSHSubscription_Apn:
// 			d.PushType, d.PushToken, d.Platform = ProviderAPNS, t.Apn, model.PlatformIOS
// 		case *authv1.PUSHSubscription_Web:
// 			d.PushType, d.Platform = ProviderWeb, model.PlatformWeb
// 			if b, err := json.Marshal(t.Web); err == nil {
// 				d.PushToken = string(b)
// 			}
// 		}

// 		if d.PushToken != "" {
// 			res = append(res, d)
// 		}
// 	}
// 	return res
// }
