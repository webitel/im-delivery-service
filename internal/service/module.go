package service

import (
	"log/slog"

	"go.uber.org/fx"
)

var Module = fx.Module(
	"service",

	fx.Provide(
		// Domain services
		fx.Annotate(
			NewDeliveryService,
			fx.As(new(Deliverer)),
		),
		fx.Annotate(
			NewPeerEnricherService,
			fx.As(new(Enricher)),
		),
	),

	// [DECORATION_LAYER] Intercept Enricher to add cross-cutting concerns
	fx.Decorate(func(orig Enricher, logger *slog.Logger) Enricher {
		return &enricherMiddleware{
			next:   orig,
			logger: logger,
		}
	}),
)
