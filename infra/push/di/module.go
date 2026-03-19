package push

import (
	"log/slog"

	"github.com/webitel/im-delivery-service/infra/push"
	"github.com/webitel/im-delivery-service/infra/push/apns"
	"github.com/webitel/im-delivery-service/infra/push/fcm"
	"github.com/webitel/im-delivery-service/infra/push/webhook"
	"github.com/webitel/im-delivery-service/internal/service"
	"go.uber.org/fx"
)

// Module provides the infrastructure components for push notification delivery.
// It uses Uber Fx value groups to collect multiple drivers into a single orchestrator.
var Module = fx.Module("push_infrastructure",
	fx.Provide(
		// -------------------------------------------------------------------------
		// [CORE ORCHESTRATOR]
		// -------------------------------------------------------------------------

		// MultiProvider manages multiple push drivers (FCM, APNs, Webhook) simultaneously.
		fx.Annotate(
			push.NewMultiProvider,
			// ParamTags match NewMultiProvider(log, drivers...):
			// 1. "" - Default logger dependency
			// 2. "group:\"push_drivers\"" - Collects all providers registered below
			fx.ParamTags(``, `group:"push_drivers"`),
			fx.As(new(service.PushProvider)),
		),

		// -------------------------------------------------------------------------
		// [FCM DRIVER]
		// -------------------------------------------------------------------------

		// Registers Google Firebase Cloud Messaging provider.
		// Now stateless: credentials will be extracted from device config at runtime.
		fx.Annotate(
			func(log *slog.Logger) push.Provider {
				return fcm.NewFCMProvider(log)
			},
			fx.ResultTags(`group:"push_drivers"`),
		),

		// -------------------------------------------------------------------------
		// [APNS DRIVER]
		// -------------------------------------------------------------------------

		// Registers Apple Push Notification service provider.
		// Now stateless: p8 tokens and topics are resolved per-request.
		fx.Annotate(
			func(log *slog.Logger) push.Provider {
				return apns.NewAPNSProvider(log)
			},
			fx.ResultTags(`group:"push_drivers"`),
		),

		// -------------------------------------------------------------------------
		// [WEBHOOK DRIVER] (Optional / Default)
		// -------------------------------------------------------------------------

		// Webhook provider can still take a default URL from config if needed,
		// but DeviceResolver can override it via dev.PushConfig.Proxy.
		fx.Annotate(
			func(log *slog.Logger) push.Provider {
				// We return a provider that can handle both static and dynamic webhooks.
				return webhook.NewWebhookProvider("")
			},
			fx.ResultTags(`group:"push_drivers"`),
		),
	),
)
