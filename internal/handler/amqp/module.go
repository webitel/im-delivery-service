package amqp

import (
	"context"
	"log/slog"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	pubsubadapter "github.com/webitel/im-delivery-service/internal/adapter/pubsub"
	"go.uber.org/fx"
)

const DeliveryExchange = "im_delivery.broadcast"

var Module = fx.Module("amqp-handler",
	fx.Provide(
		pubsubadapter.NewSubscriberProvider,
		pubsubadapter.NewPublisherProvider,

		// [FIX] Building the publisher.
		// If your pp.Build only takes a string, we pass the Exchange name.
		func(pp *pubsubadapter.PublisherProvider) (message.Publisher, error) {
			return pp.Build(DeliveryExchange)
		},

		// [DISPATCHER] Domain-aware wrapper for the publisher
		func(pub message.Publisher) pubsubadapter.EventDispatcher {
			return pubsubadapter.NewEventDispatcher(pub)
		},

		NewMessageHandler,

		func(logger *slog.Logger) (*message.Router, error) {
			return message.NewRouter(message.RouterConfig{}, watermill.NewSlogLogger(logger))
		},
	),

	fx.Invoke(func(
		lc fx.Lifecycle,
		h *MessageHandler,
		router *message.Router,
		subProvider *pubsubadapter.SubscriberProvider,
		logger *slog.Logger,
	) error {
		// [WIRING] Register all defined consumers
		if err := h.RegisterHandlers(router, subProvider); err != nil {
			return err
		}

		lc.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				go func() {
					if err := router.Run(context.Background()); err != nil {
						logger.Error("router runtime error", "err", err)
					}
				}()
				return nil
			},
			OnStop: func(ctx context.Context) error {
				return router.Close()
			},
		})
		return nil
	}),
)
