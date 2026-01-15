package registry

import (
	"context"
	"time"

	"go.uber.org/fx"
)

var Module = fx.Module("registry",
	fx.Provide(
		// Provide the concrete implementation with default settings
		func() *Hub {
			return NewHub(
				1*time.Hour,    // evictionInterval: Clean once per hour
				10*time.Minute, // idleTimeout: Wait 10m after last disconnect
			)
		},
		// Annotate to expose Hub as Hubber interface
		fx.Annotate(
			func(h *Hub) Hubber { return h },
			fx.As(new(Hubber)),
		),
	),
	// [LIFECYCLE_MANAGEMENT]
	// Ensure the background routines are stopped when the app shuts down.
	fx.Invoke(func(lc fx.Lifecycle, h Hubber) {
		lc.Append(fx.Hook{
			OnStop: func(ctx context.Context) error {
				h.Shutdown()
				return nil
			},
		})
	}),
)
