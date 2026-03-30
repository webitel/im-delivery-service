package servicedi

import (
	"context"
	"time"

	"github.com/webitel/im-delivery-service/internal/service"
	"go.uber.org/fx"
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

		// --- Domain & Helper Services ---

		// [PRESENCE] Tracks user online/offline status and session heat-maps.
		fx.Annotate(service.NewPresenceService, fx.As(new(service.PresenceManager))),

		// [ENRICHER] Fetches additional contact metadata from external account services.
		fx.Annotate(service.NewContactEnricher, fx.As(new(service.Contacter))),

		// [AUTH] Validates JWT/OAuth tokens and builds the initial security context.
		fx.Annotate(service.NewAuthService, fx.As(new(service.Auther))),
	),

	// [LIFECYCLE_INVOCATION]
	// Bootstraps background workers and ensures clean resource disposal on shutdown.
	fx.Invoke(func(lc fx.Lifecycle, o service.Orchestrator, ph service.Pusher) {
		lc.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				// Launch the push notification polling loop in a dedicated goroutine.
				go ph.Start(context.Background())
				return nil
			},
			OnStop: func(ctx context.Context) error {
				// Trigger graceful shutdown for the orchestrator if supported.
				if s, ok := o.(interface{ Close() error }); ok {
					return s.Close()
				}
				return nil
			},
		})
	}),
)
