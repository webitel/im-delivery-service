package webiteldi

import (
	"context"

	imauth "github.com/webitel/im-delivery-service/infra/client/im-auth"
	imcontact "github.com/webitel/im-delivery-service/infra/client/im-contact"
	"go.uber.org/fx"
)

var Module = fx.Module(
	"webitel_clients",

	// [CONSTRUCTOR] Provides the resilient contact client
	fx.Provide(imcontact.New),
	fx.Provide(imauth.New),

	// [LIFECYCLE] Ensures the gRPC connection pool is closed gracefully on app shutdown
	fx.Invoke(func(lc fx.Lifecycle, client *imcontact.Client) {
		lc.Append(fx.Hook{
			OnStop: func(ctx context.Context) error {
				return client.Close()
			},
		})
	}),

	fx.Invoke(func(lc fx.Lifecycle, client *imauth.Client) {
		lc.Append(fx.Hook{
			OnStop: func(ctx context.Context) error {
				return client.Close()
			},
		})
	}),
)
