package amqp

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	pubsubadapter "github.com/webitel/im-delivery-service/internal/adapter/pubsub"
	"github.com/webitel/im-delivery-service/internal/domain/registry"
	"github.com/webitel/im-delivery-service/internal/service"
)

const (
	// Listening to  (Consumed)
	WebitelExchange = "im_message.events"
	MessageTopicV1  = "im_message.#.message.created.v1"

	// Publishing to  (Produced)
	DeliveryExchange = "im_delivery.events"
	MessageQueueV1   = "im_delivery.message_v1"
)

type MessageHandler struct {
	hub       registry.Hubber
	logger    *slog.Logger
	enricher  service.Enricher
	publisher pubsubadapter.EventDispatcher
}

func NewMessageHandler(
	hub registry.Hubber,
	logger *slog.Logger,
	enricher service.Enricher,
	publisher pubsubadapter.EventDispatcher,
) *MessageHandler {
	return &MessageHandler{
		hub:       hub,
		logger:    logger,
		enricher:  enricher,
		publisher: publisher,
	}
}

// RegisterHandlers configures AMQP subscriptions for the service node.
// internal/adapter/amqp/handler.go

func (h *MessageHandler) RegisterHandlers(
	router *message.Router,
	subProvider *pubsubadapter.SubscriberProvider,
) error {
	nodeID := watermill.NewShortUUID()
	queueName := fmt.Sprintf("%s.%s", MessageQueueV1, nodeID)

	// [INFRASTRUCTURE] Define poison queue for failed messages
	poisonTopic := fmt.Sprintf("%s.poison", queueName)

	sub, err := subProvider.Build(queueName, WebitelExchange, MessageTopicV1)
	if err != nil {
		return fmt.Errorf("failed to build subscriber: %w", err)
	}

	// [DLX_SETUP] Access the raw message.Publisher via .Publisher() method
	// This fixes the "does not implement message.Publisher (missing method Close)" error
	poisonMiddleware, err := middleware.PoisonQueue(h.publisher.Publisher(), poisonTopic)
	if err != nil {
		return fmt.Errorf("failed to setup poison middleware: %w", err)
	}

	// [ROUTING] Fix AddNoPublisherHandler signature: (handlerName, topic, subscriber, handlerFunc)
	handler := router.AddConsumerHandler(
		queueName+"_executor",
		MessageTopicV1, // This was missing
		sub,
		bind(h.hub, h.OnMessageCreatedV1),
	)

	// [RESILIENCE_CHAIN] Apply middlewares in correct order
	handler.AddMiddleware(
		poisonMiddleware, // Catch errors first
		middleware.NewThrottle(100, time.Second).Middleware,
		middleware.Timeout(time.Second*30),
	)

	return nil
}
