package cmd

import (
	"github.com/webitel/im-delivery-service/config"
	grpcsrv "github.com/webitel/im-delivery-service/infra/server/grpc"
	webiteldi "github.com/webitel/im-delivery-service/infra/webitel/di"
	"github.com/webitel/im-delivery-service/internal/domain/registry"
	amqpdi "github.com/webitel/im-delivery-service/internal/handler/amqp"
	grpchandler "github.com/webitel/im-delivery-service/internal/handler/grpc"
	"github.com/webitel/im-delivery-service/internal/service"
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
		webiteldi.Module,
		service.Module,
		registry.Module,
		grpchandler.Module,
		grpcsrv.Module,
		amqpdi.Module,
	)
}
