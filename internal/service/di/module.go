package servicedi

import (
	"context"
	"time"

	"github.com/webitel/im-delivery-service/internal/service"
	"go.uber.org/fx"
)

var Module = fx.Module(
	"service",

	fx.Provide(
		// --- Configuration ---

		// [WORKER_CONFIG] Total goroutines for background task consumption.
		fx.Annotate(
			func() int { return 256 },
			fx.ResultTags(`name:"worker_count"`),
		),

		// [TIMEOUT_CONFIG] Grace period for client acknowledgment.
		fx.Annotate(
			func() time.Duration { return 20 * time.Second },
			fx.ResultTags(`name:"ack_timeout"`),
		),

		// --- Core Orchestration ---

		// [EVENT_ORCHESTRATOR] Central engine for Publish/Dismiss/Ack.
		fx.Annotate(
			service.NewEventOrchestrator,
			fx.As(new(service.Orchestrator)),
		),

		// [SESSION_SERVICE] Manages transport-level connectivity and presence.
		fx.Annotate(
			service.NewSessionService,
			fx.As(new(service.SessionManager)),
		),

		service.NewDeviceResolver,

		// --- Handlers Registry & Push ---

		// 1. Provide concrete implementation
		service.NewPushHandler,

		// 2. Register as a single Pusher interface for Invoke
		fx.Annotate(
			func(h *service.PushHandler) service.Pusher { return h },
			fx.As(new(service.Pusher)),
		),

		// 3. Register into "event_handlers" group for Orchestrator
		fx.Annotate(
			func(h *service.PushHandler) service.EventHandler { return h },
			fx.As(new(service.EventHandler)),
			fx.ResultTags(`group:"event_handlers"`),
		),

		// --- Infrastructure & Domain Services ---

		fx.Annotate(service.NewPresenceService, fx.As(new(service.PresenceManager))),
		fx.Annotate(service.NewContactEnricher, fx.As(new(service.Contacter))),
		fx.Annotate(service.NewAuthService, fx.As(new(service.Auther))),
	),

	// [LIFECYCLE_MANAGEMENT]
	fx.Invoke(func(lc fx.Lifecycle, o service.Orchestrator, ph service.Pusher) {
		lc.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				// Start background polling loop
				go ph.Start(context.Background())
				return nil
			},
			OnStop: func(ctx context.Context) error {
				// Graceful shutdown for orchestrator
				if s, ok := o.(interface{ Close() error }); ok {
					return s.Close()
				}
				return nil
			},
		})
	}),
)
