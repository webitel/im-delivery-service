package leader

import (
	"go.uber.org/fx"
)

var Module = fx.Module(
	"leader",
	fx.Provide(
		// [STATE_HOLDER] Provide the concrete type to allow status updates
		func() *Status {
			return &Status{}
		},

		// [INTERFACE_CAST] Provide the same instance as a read-only LeaderAwarer
		fx.Annotate(
			func(ls *Status) LeaderAwarer { return ls },
			fx.As(new(LeaderAwarer)),
		),

		fx.Annotate(
			ProvideLeaderElector,
			fx.As(new(Elector)),
		),
	),
)
