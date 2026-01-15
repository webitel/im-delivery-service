package registry

import "go.uber.org/fx"

var Module = fx.Module("registry",
	fx.Provide(
		NewHub,
		fx.Annotate(
			func(h *Hub) Hubber { return h },
			fx.As(new(Hubber)),
		),
	),
)
