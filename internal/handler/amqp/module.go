package amqp

import (
	"context"
	"log/slog"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	pubsubadapter "github.com/webitel/im-delivery-service/internal/adapter/pubsub"
	"go.uber.org/fx"
)

var Module = fx.Module("amqp-handler",
	fx.Provide(
		pubsubadapter.NewPublisherProvider,
		pubsubadapter.NewSubscriberProvider,
		func(pp *pubsubadapter.PublisherProvider) (pubsubadapter.EventDispatcher, error) {
			pub, err := pp.Build("im.delivery.events")
			if err != nil {
				return nil, err
			}

			return pubsubadapter.NewEventDispatcher(pub), nil
		},
		NewMessageHandler,

		// [INFRASTRUCTURE] Simple factory for the router
		func(logger *slog.Logger) (*message.Router, error) {
			return message.NewRouter(message.RouterConfig{}, watermill.NewSlogLogger(logger))
		},
	),

	// [LIFECYCLE] Centralized management of router and handlers
	fx.Invoke(func(
		lc fx.Lifecycle,
		h *MessageHandler,
		router *message.Router,
		sub *pubsubadapter.SubscriberProvider,
		logger *slog.Logger,
	) error {
		// 1. Register domain handlers
		if err := h.RegisterHandlers(router, sub); err != nil {
			return err
		}

		// 2. Manage router lifecycle
		lc.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				go func() {
					if err := router.Run(context.Background()); err != nil {
						logger.Error("watermill router run error", "err", err)
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
