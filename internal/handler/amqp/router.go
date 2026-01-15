package amqp

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	pubsubadapter "github.com/webitel/im-delivery-service/internal/adapter/pubsub"
	"github.com/webitel/im-delivery-service/internal/domain/registry"
	"go.uber.org/fx"
)

const (
	WebitelExchange = "im_message.events"
)

type MessageHandler struct {
	hub    registry.Hubber
	logger *slog.Logger
}

func NewMessageHandler(hub registry.Hubber, logger *slog.Logger) *MessageHandler {
	return &MessageHandler{hub: hub, logger: logger}
}

// RegisterHandlers configures AMQP subscriptions for the service node.
func RegisterHandlers(
	router *message.Router,
	subProvider *pubsubadapter.SubscriberProvider,
	h *MessageHandler,
	hub registry.Hubber,
) error {
	// Identify the node to create a unique temporary queue for fan-out messaging
	nodeID, err := os.Hostname()
	if err != nil {
		nodeID = watermill.NewShortUUID()
	}

	routes := []struct {
		topic   string
		queue   string
		handler message.NoPublishHandlerFunc
	}{
		{
			topic:   MessageTopicV1,
			queue:   MessageQueueV1,
			handler: bind(hub, h.OnMessageCreatedV1), // bind now returns NoPublishHandlerFunc
		},
	}

	for _, r := range routes {
		// Each node gets its own queue to ensure every instance receives the event
		uniqueQueue := fmt.Sprintf("%s.%s", r.queue, nodeID)

		sub, err := subProvider.Build(uniqueQueue, WebitelExchange, r.topic)
		if err != nil {
			return fmt.Errorf("failed to build subscriber for %s: %w", uniqueQueue, err)
		}

		// Registering handler without an output topic (pure consumption)
		router.AddNoPublisherHandler(
			uniqueQueue+"_executor",
			r.topic,
			sub,
			r.handler,
		)
	}
	return nil
}

// NewWatermillRouter initializes the router and manages its lifecycle via Uber Fx
func NewWatermillRouter(lc fx.Lifecycle, logger *slog.Logger) (*message.Router, error) {
	router, err := message.NewRouter(message.RouterConfig{}, watermill.NewSlogLogger(logger))
	if err != nil {
		return nil, err
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				// Running the router in a background goroutine
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

	return router, nil
}
