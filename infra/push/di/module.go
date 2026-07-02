package push

import (
	"log/slog"

	"go.uber.org/fx"

	"github.com/webitel/im-delivery-service/infra/push"
	"github.com/webitel/im-delivery-service/infra/push/apns"
	"github.com/webitel/im-delivery-service/infra/push/fcm"
	"github.com/webitel/im-delivery-service/internal/service"
)

// [PUSH_INFRASTRUCTURE_MODULE]
// ---------------------------------------------------------------------------------
// [LOGIC]
// - Collects specific push drivers (FCM, APNS) into a MultiProvider.
// - Webhook is now a generic utility used INTERNALLY by drivers for debugging.
// ---------------------------------------------------------------------------------
var Module = fx.Module("push_infrastructure",
	fx.Provide(
		// [CORE_ORCHESTRATOR]
		// Combines all registered drivers into a single service.PushProvider interface.
		fx.Annotate(
			push.NewMultiProvider,
			fx.ParamTags(``, `group:"push_drivers"`),
			fx.As(new(service.PushProvider)),
		),

		// [FCM_DRIVER]
		// Registered as a member of "push_drivers" group.
		fx.Annotate(
			func(log *slog.Logger) push.Provider {
				return fcm.NewProvider(log)
			},
			fx.ResultTags(`group:"push_drivers"`),
		),

		// [APNS_DRIVER]
		// Registered as a member of "push_drivers" group.
		fx.Annotate(
			func(log *slog.Logger) push.Provider {
				return apns.NewProvider(log)
			},
			fx.ResultTags(`group:"push_drivers"`),
		),
	),
)
