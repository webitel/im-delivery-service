package amqp

import (
	"context"
	"log/slog"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	leader "github.com/webitel/im-delivery-service/infra/discovery/consul"
	pubsubadapter "github.com/webitel/im-delivery-service/internal/adapter/pubsub"
	"github.com/webitel/im-delivery-service/internal/domain/event"
	"github.com/webitel/webitel-go-kit/pkg/semconv"
	"go.uber.org/fx"
)

// [CONSTANTS] Global infrastructure settings
const DeliveryExchange = "im_delivery.broadcast"

// [DISPATCHER_PROXY] Gates execution based on the shared leadership state from leader package.
type atomicDispatcher struct {
	awarer leader.LeaderAwarer
	base   pubsubadapter.EventDispatcher
}

// Publish only executes if the current node is the leader.
func (a *atomicDispatcher) Publish(ctx context.Context, ev event.Eventer) error {
	if !a.awarer.IsLeader() {
		// [FOLLOWER] Standby nodes silently ignore outgoing events to prevent duplicates.
		return nil
	}
	return a.base.Publish(ctx, ev)
}

// Publisher returns the underlying watermill publisher if leading.
func (a *atomicDispatcher) Publisher() message.Publisher {
	if !a.awarer.IsLeader() {
		return nil
	}
	return a.base.Publisher()
}

// [DEPENDENCIES] Parameter structure for the Fx Invoke stage.
type invokeParams struct {
	fx.In

	Lifecycle fx.Lifecycle
	Handler   *MessageHandler
	Router    *message.Router
	SubProv   *pubsubadapter.SubscriberProvider
	Elector   leader.Elector
	Logger    *slog.Logger
	// [STATE] We inject the concrete Status to update leadership during transitions.
	Status *leader.Status
}

var Module = fx.Module("amqp-handler",
	fx.Provide(
		pubsubadapter.NewSubscriberProvider,
		pubsubadapter.NewPublisherProvider,

		// [BROKER] Primary RabbitMQ publisher factory
		func(pp *pubsubadapter.PublisherProvider) (message.Publisher, error) {
			return pp.Build(DeliveryExchange)
		},

		// [BASE_DISPATCHER] The actual domain logic registered with a name to avoid cycles.
		fx.Annotate(
			pubsubadapter.NewEventDispatcher,
			fx.ResultTags(`name:"base_dispatcher"`),
		),

		// [PROXY_DISPATCHER] We use fx.ParamTags to tell Fx which dependency is the "base_dispatcher"
		fx.Annotate(
			func(aw leader.LeaderAwarer, base pubsubadapter.EventDispatcher) pubsubadapter.EventDispatcher {
				return &atomicDispatcher{
					awarer: aw,
					base:   base,
				}
			},
			fx.As(new(pubsubadapter.EventDispatcher)),
			// [TAGS] The first nil is for LeaderAwarer (no tag), the second is for our base dispatcher
			fx.ParamTags(``, `name:"base_dispatcher"`),
		),

		NewMessageHandler,

		// [ROUTER] Watermill router for handling incoming AMQP messages
		func(logger *slog.Logger) (*message.Router, error) {
			return message.NewRouter(message.RouterConfig{}, watermill.NewSlogLogger(logger))
		},
	),

	fx.Invoke(func(p invokeParams) error {
		// [WIRING] Map all AMQP topics to domain message handlers
		if err := p.Handler.RegisterHandlers(p.Router, p.SubProv); err != nil {
			return err
		}

		mainCtx, cancelMain := context.WithCancel(context.Background())

		p.Lifecycle.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				// [WORKER] Start the background message consumer
				go func() {
					if err := p.Router.Run(mainCtx); err != nil {
						p.Logger.Error("AMQP_ROUTER_STOPPED", slog.Any(semconv.ErrorKey, err))
					}
				}()

				// [ORCHESTRATION] Start the Consul leader election loop
				go p.Elector.Run(mainCtx,
					func(leaderCtx context.Context) error {
						p.Logger.Info("LEADER_PROMOTED: enabling event delivery")
						// [SYNC] Set global status to true, enabling the atomicDispatcher
						p.Status.Set(true)
						return nil
					},
					func() {
						p.Logger.Warn("LEADER_DEMOTED: disabling event delivery")
						// [SYNC] Reset global status, putting atomicDispatcher into standby mode
						p.Status.Set(false)
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
