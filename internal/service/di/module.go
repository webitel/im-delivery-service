package servicedi

import (
	"log/slog"

	"github.com/webitel/im-delivery-service/internal/service"
	"go.uber.org/fx"
)

var Module = fx.Module(
	"service",

	fx.Provide(
		// Domain services
		fx.Annotate(
			service.NewDeliveryService,
			fx.As(new(service.Deliverer)),
		),
		fx.Annotate(
			service.NewPeerEnricherService,
			fx.As(new(service.Enricher)),
		),
		fx.Annotate(
			service.NewAuthService,
			fx.As(new(service.Auther)),
		),
	),

	// [DECORATION_LAYER] Intercept Enricher to add cross-cutting concerns
	fx.Decorate(func(orig service.Enricher, logger *slog.Logger) service.Enricher {
		return &service.EnricherMiddleware{
			Next:   orig,
			Logger: logger,
		}
	}),
)
