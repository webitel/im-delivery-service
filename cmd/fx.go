package cmd

import (
	"go.uber.org/fx"

	"github.com/webitel/webitel-go-kit/infra/discovery"
	"github.com/webitel/webitel-go-kit/infra/profiler"

	"github.com/webitel/im-delivery-service/config"
	webiteldi "github.com/webitel/im-delivery-service/infra/client/di"
	leader "github.com/webitel/im-delivery-service/infra/discovery/consul"
	pushdi "github.com/webitel/im-delivery-service/infra/push/di"
	grpcsrv "github.com/webitel/im-delivery-service/infra/server/grpc"
	httpsrv "github.com/webitel/im-delivery-service/infra/server/http"
	"github.com/webitel/im-delivery-service/infra/tls"
	"github.com/webitel/im-delivery-service/internal/domain/registry"
	amqpdi "github.com/webitel/im-delivery-service/internal/handler/amqp"
	grpchandler "github.com/webitel/im-delivery-service/internal/handler/grpc"
	wshandler "github.com/webitel/im-delivery-service/internal/handler/ws"
	servicedi "github.com/webitel/im-delivery-service/internal/service/di"
	redisdi "github.com/webitel/im-delivery-service/internal/store/redis"
)

// NewApp assembles the application dependency graph and manages its lifecycle.
func NewApp(cfg *config.Config) *fx.App {
	return fx.New(
		// [CORE] Essential application-wide dependencies
		fx.Provide(
			func() *config.Config { return cfg },
			ProvideLogger,
			ProvideWatermillLogger,
			ProvideSD,
			ProvidePubSub,
			ProvideRedis,
			ProvideProfiler,
		),

		// [INIT] Global service discovery orchestration
		fx.Invoke(func(discovery discovery.DiscoveryProvider) error { return nil }),

		// [INFRA] Common infrastructure and security components
		tls.Module,       // TLS/SSL certificates management
		webiteldi.Module, // Internal Webitel clients (storage, contacts, etc.)

		// [DOMAIN] Business logic and service orchestration
		servicedi.Module, // Core service implementations
		registry.Module,  // Real-time connection registry

		// [TRANSPORT] External interfaces and communication protocols
		grpchandler.Module, // gRPC API endpoints
		wshandler.Module,   // WebSocket event delivery system
		amqpdi.Module,      // Messaging handlers and background workers

		// [SERVERS] Protocol-specific server listeners
		grpcsrv.Module, // gRPC server lifecycle (port listening)
		httpsrv.Module, // HTTP/WS server lifecycle (port listening)

		// [DISTRIBUTED_CONSENSUS] Orchestrates leadership state to ensure
		// strictly single-active-node publishing to the message broker.
		leader.Module,

		// [STORES] Data persistence and caching layers
		redisdi.Module,

		// PUSH service module
		pushdi.Module,

		// Fetch metrics from profiler
		profiler.Module,
	)
}
