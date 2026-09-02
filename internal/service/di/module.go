package servicedi

import (
	"context"
	"time"

	"go.uber.org/fx"

	imadmin "github.com/webitel/im-delivery-service/infra/client/im-admin"
	imthread "github.com/webitel/im-delivery-service/infra/client/im-thread"
	"github.com/webitel/im-delivery-service/internal/service"
)

// Module defines the service-layer dependency injection tree.
// It orchestrates core business logic, background workers, and push notification life-cycles.
var Module = fx.Module(
	"service",

	fx.Provide(
		// --- Global Configurations ---

		// [WORKER_POOL_SIZE] Total concurrent goroutines for message processing.
		fx.Annotate(
			func() int { return 256 },
			fx.ResultTags(`name:"worker_count"`),
		),

		// [ACK_TIMEOUT] Duration to wait for client delivery confirmation before fallback.
		fx.Annotate(
			func() time.Duration { return 20 * time.Second },
			fx.ResultTags(`name:"ack_timeout"`),
		),

		// --- Core Orchestration & Session Management ---

		// [ORCHESTRATOR] Primary engine handling event routing, acks, and state transitions.
		fx.Annotate(
			service.NewEventOrchestrator,
			fx.As(new(service.Orchestrator)),
		),

		// [SESSION_MANAGER] Tracks real-time transport connectivity (WebSocket/gRPC).
		fx.Annotate(
			service.NewSessionService,
			fx.As(new(service.SessionManager)),
		),

		// [DEVICE_PROVIDER] Resolves user push-tokens and platform-specific device metadata.
		fx.Annotate(
			service.NewDeviceService,
			fx.As(new(service.DeviceProvider)),
		),

		// --- Background Push Infrastructure ---

		// 1. Concrete Implementation: Responsible for the actual delivery logic.
		service.NewPushHandler,

		// 2. Lifecycle Interface: Exposes Start/Stop methods for background polling.
		fx.Annotate(
			func(h *service.PushHandler) service.Pusher { return h },
			fx.As(new(service.Pusher)),
		),

		// 3. Event Handling: Plugs the PushHandler into the Orchestrator's event loop via Group Tags.
		fx.Annotate(
			func(h *service.PushHandler) service.EventHandler { return h },
			fx.As(new(service.EventHandler)),
			fx.ResultTags(`group:"event_handlers"`),
		),

		// --- Delivery Status Reporting ---

		// 1. Concrete Implementation: resolves ACKs/pushes into MarkDelivered
		//    reports for im-thread-service (batched, best-effort).
		service.NewMessageStatusReporter,

		// 2. Thread client surface consumed by the reporter.
		fx.Annotate(
			func(c *imthread.Client) service.ThreadStatusClient { return c },
			fx.As(new(service.ThreadStatusClient)),
		),

		// 3. Delivery confirmations funnel (WS ACKs via Orchestrator, pushes via PushHandler).
		fx.Annotate(
			func(r *service.MessageStatusReporter) service.DeliveryConfirmer { return r },
			fx.As(new(service.DeliveryConfirmer)),
		),

		// 4. Event Handling: observes message fan-out envelopes to remember
		//    their message context for later ACK resolution.
		fx.Annotate(
			func(r *service.MessageStatusReporter) service.EventHandler { return r },
			fx.As(new(service.EventHandler)),
			fx.ResultTags(`group:"event_handlers"`),
		),

		// --- Domain & Helper Services ---

		// [PRESENCE] Tracks user online/offline status and session heat-maps.
		fx.Annotate(service.NewPresenceService, fx.As(new(service.PresenceManager))),

		// [ENRICHER] Fetches additional contact metadata from external account services.
		fx.Annotate(service.NewContactEnricher, fx.As(new(service.Contacter))),

		// [AUTH] Validates JWT/OAuth tokens and builds the initial security context.
		fx.Annotate(service.NewAuthService, fx.As(new(service.Auther))),

		// [APP_CONFIG] Resolves per-application delivery policy for system messages.
		fx.Annotate(service.NewAppConfigService, fx.As(new(service.AppConfigProvider))),

		// [ADMIN_APP_SEARCHER] Adapts *imadmin.Client to the narrow AdminAppSearcher interface (mirrors the ThreadStatusClient adapter above).
		fx.Annotate(
			func(c *imadmin.Client) service.AdminAppSearcher { return c },
			fx.As(new(service.AdminAppSearcher)),
		),
	),

	// [LIFECYCLE_INVOCATION]
	// Bootstraps background workers and ensures clean resource disposal on shutdown.
	fx.Invoke(func(lc fx.Lifecycle, o service.Orchestrator, ph service.Pusher, reporter *service.MessageStatusReporter) {
		lc.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				// Launch the push notification polling loop in a dedicated goroutine.
				go ph.Start(context.Background())

				return nil
			},
			OnStop: func(ctx context.Context) error {
				// Stop the orchestrator first so its worker pool stops
				// producing new status confirmations, then drain the reporter.
				// Late confirmations from WS pumps / push goroutines after this
				// point are safely dropped by the reporter's shutdown guard
				// rather than panicking on a closed channel.
				if s, ok := o.(interface{ Close() error }); ok {
					if err := s.Close(); err != nil {
						return err
					}
				}

				// Drain pending delivery/read receipts before shutdown.
				return reporter.Close()
			},
		})
	}),
)
