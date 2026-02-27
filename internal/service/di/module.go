package servicedi

import (
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
		// Now directly providing ContactEnricher without decorator
		fx.Annotate(
			service.NewContactEnricher,
			fx.As(new(service.Contacter)),
		),
		fx.Annotate(
			service.NewAuthService,
			fx.As(new(service.Auther)),
		),
	),
)
