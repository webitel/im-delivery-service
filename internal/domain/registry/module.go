package registry

import (
	"context"
	"time"

	"go.uber.org/fx"
)

var Module = fx.Module("registry",
	fx.Provide(
		// [CLEAN_INJECTION] Configure Hub using Functional Options
		func() *Hub {
			return NewHub(
				WithEvictionInterval(15*time.Minute),
				WithIdleTimeout(30*time.Minute),
				WithMailboxSize(2048),
			)
		},
		fx.Annotate(
			func(h *Hub) Hubber { return h },
			fx.As(new(Hubber)),
		),
	),
	fx.Invoke(func(lc fx.Lifecycle, h Hubber) {
		lc.Append(fx.Hook{
			OnStop: func(ctx context.Context) error {
				h.Shutdown() // [GRACEFUL_SHUTDOWN] Stop all Actor goroutines
				return nil
			},
		})
	}),
)
