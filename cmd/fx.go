package cmd

import (
	"github.com/webitel/im-delivery-service/config"
	webiteldi "github.com/webitel/im-delivery-service/infra/client/di"
	grpcsrv "github.com/webitel/im-delivery-service/infra/server/grpc"
	"github.com/webitel/im-delivery-service/infra/tls"
	"github.com/webitel/im-delivery-service/internal/domain/registry"
	amqpdi "github.com/webitel/im-delivery-service/internal/handler/amqp"
	grpchandler "github.com/webitel/im-delivery-service/internal/handler/grpc"
	servicedi "github.com/webitel/im-delivery-service/internal/service/di"
	"github.com/webitel/webitel-go-kit/infra/discovery"
	"go.uber.org/fx"
)

func NewApp(cfg *config.Config) *fx.App {
	return fx.New(
		fx.Provide(
			func() *config.Config { return cfg },
			ProvideLogger,
			ProvideWatermillLogger,
			ProvideSD,
			ProvidePubSub,
		),
		fx.Invoke(func(discovery discovery.DiscoveryProvider) error { return nil }),
		tls.Module,
		webiteldi.Module,
		servicedi.Module,
		registry.Module,
		grpchandler.Module,
		grpcsrv.Module,
		amqpdi.Module,
	)
}
