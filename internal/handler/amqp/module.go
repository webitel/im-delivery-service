package amqp

import (
	"context"
	"log/slog"
	"sync/atomic"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	leader "github.com/webitel/im-delivery-service/infra/discovery/consul"
	pubsubadapter "github.com/webitel/im-delivery-service/internal/adapter/pubsub"
	"github.com/webitel/im-delivery-service/internal/domain/event"
	"go.uber.org/fx"
)

// [INFRASTRUCTURE] Global delivery constants
const DeliveryExchange = "im_delivery.broadcast"

// [INTERFACE_ADAPTER] atomicDispatcher gates all side-effects based on leadership state
type atomicDispatcher struct {
	active atomic.Pointer[pubsubadapter.EventDispatcher]
}

func (a *atomicDispatcher) Publish(ctx context.Context, ev event.Eventer) error {
	if d := a.active.Load(); d != nil {
		return (*d).Publish(ctx, ev)
	}
	return nil // [NO-OP] Followers silently drop events to prevent duplicates
}

func (a *atomicDispatcher) Publisher() message.Publisher {
	if d := a.active.Load(); d != nil {
		return (*d).Publisher()
	}
	return nil
}

func (a *atomicDispatcher) IsLeader() bool {
	return a.active.Load() != nil
}

// [DEPENDENCIES] Explicit parameter structure for the Invoke stage
type invokeParams struct {
	fx.In

	Lifecycle fx.Lifecycle
	Handler   *MessageHandler
	Router    *message.Router
	SubProv   *pubsubadapter.SubscriberProvider
	Elector   leader.LeadershipElector
	Logger    *slog.Logger

	// [RESOLVE] Get the raw implementation by name to avoid the Proxy-to-Proxy loop
	BaseDispatcher pubsubadapter.EventDispatcher `name:"base_dispatcher"`
	// [RESOLVE] Get the proxy pointer to manage its state
	Proxy *atomicDispatcher
}

var Module = fx.Module("amqp-handler",
	fx.Provide(
		pubsubadapter.NewSubscriberProvider,
		pubsubadapter.NewPublisherProvider,

		// [BROKER] RabbitMQ publisher factory
		func(pp *pubsubadapter.PublisherProvider) (message.Publisher, error) {
			return pp.Build(DeliveryExchange)
		},

		// [BASE_DISPATCHER] Register the real logic with a NAME
		fx.Annotate(
			pubsubadapter.NewEventDispatcher,
			fx.ResultTags(`name:"base_dispatcher"`),
		),

		// [PROXY_DISPATCHER] Register the atomic proxy as the PRIMARY interface for the app
		// This is what NewMessageHandler will receive.
		fx.Annotate(
			func() *atomicDispatcher {
				return &atomicDispatcher{}
			},
			fx.As(new(pubsubadapter.EventDispatcher)),
		),

		// [STATE_ACCESS] Also provide the raw pointer so we can call Store() in Invoke
		func(d pubsubadapter.EventDispatcher) *atomicDispatcher {
			return d.(*atomicDispatcher)
		},

		NewMessageHandler,

		func(logger *slog.Logger) (*message.Router, error) {
			return message.NewRouter(message.RouterConfig{}, watermill.NewSlogLogger(logger))
		},
	),

	fx.Invoke(func(p invokeParams) error {
		// [WIRING] Map AMQP topics to domain handlers
		if err := p.Handler.RegisterHandlers(p.Router, p.SubProv); err != nil {
			return err
		}

		mainCtx, cancelMain := context.WithCancel(context.Background())

		p.Lifecycle.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				// [WORKER] Standard consumer routine
				go func() {
					if err := p.Router.Run(mainCtx); err != nil {
						p.Logger.Error("AMQP_ROUTER_FAILURE", "err", err)
					}
				}()

				// [LEADERSHIP] Dynamic implementation swapping
				go p.Elector.Run(mainCtx,
					func(leaderCtx context.Context) error {
						p.Logger.Info("LEADER_PROMOTED: enabling global event dispatcher")
						// [HOT_SWAP] Map the proxy to the real named dispatcher
						p.Proxy.active.Store(&p.BaseDispatcher)
						return nil
					},
					func() {
						p.Logger.Warn("LEADER_DEMOTED: deactivating global event dispatcher")
						// [SHUTDOWN] Disable the gate
						p.Proxy.active.Store(nil)
					},
				)
				return nil
			},
			OnStop: func(ctx context.Context) error {
				cancelMain()
				return p.Router.Close()
			},
		})
		return nil
	}),
)
