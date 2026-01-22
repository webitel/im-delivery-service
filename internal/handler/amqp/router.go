package amqp

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/google/uuid"
	"github.com/webitel/im-delivery-service/internal/adapter/pubsub"
	"github.com/webitel/im-delivery-service/internal/domain/registry"
	"github.com/webitel/im-delivery-service/internal/service"
)

const (
	// ------------------- EXCHANGES (SOURCES) -------------------
	MessageEventsExchange = "im_message.events"
	SystemEventsExchange  = "im_system.events"

	// ------------------- TOPICS (ROUTING KEYS) -----------------
	TopicMessageCreated = "im_message.#.message.created.v1"
	TopicMessageDeleted = "im_message.#.message.deleted.v1"
	TopicUserStatus     = "im_system.#.user.status.v1"

	// ------------------- QUEUES (CONSUMERS) --------------------
	DeliveryProcessorQueue = "im-delivery.incoming-processor.v1"
	DeliveryPoisonTopic    = "im-delivery.incoming-processor.v1.poison"
)

type MessageHandler struct {
	hub        registry.Hubber
	logger     *slog.Logger
	enricher   service.Enricher
	dispatcher pubsub.EventDispatcher
}

func NewMessageHandler(hub registry.Hubber, logger *slog.Logger, enricher service.Enricher, dispatcher pubsub.EventDispatcher) *MessageHandler {
	return &MessageHandler{hub, logger, enricher, dispatcher}
}

// [REGISTRATION_PIPELINE]
func (h *MessageHandler) RegisterHandlers(router *message.Router, subProvider *pubsub.SubscriberProvider) error {
	poison, err := middleware.PoisonQueue(h.dispatcher.Publisher(), DeliveryPoisonTopic)
	if err != nil {
		return fmt.Errorf("POISON_SETUP_FAILED: %w", err)
	}

	configs := []struct {
		name     string
		exchange string
		topic    string
		handler  message.NoPublishHandlerFunc
	}{
		{"ON_MSG_CREATED", MessageEventsExchange, TopicMessageCreated, Bind(h, h.OnMessageCreatedV1)},

		// [ARCHITECTURAL_PLACEHOLDERS]
		// The following handlers serve as blueprints for scaling the system.
		// Add new domain listeners here by following this table-driven pattern.
		{"ON_MSG_DELETED", MessageEventsExchange, TopicMessageDeleted, Bind(h, h.OnMessageDeletedV1)},
		{"ON_USR_STATUS", SystemEventsExchange, TopicUserStatus, Bind(h, h.OnStatusChangedV1)},
	}

	for _, c := range configs {
		instanceID := uuid.NewString()[:8]
		// [UNIQUE_HANDLER_QUEUE]
		// We create a unique queue for EACH handler on THIS node.
		// Format: im-delivery.node.b23a8f12.ON_MSG_CREATED
		handlerQueue := fmt.Sprintf("%s.%s.%s", DeliveryProcessorQueue, instanceID, c.name)

		sub, err := subProvider.Build(handlerQueue, c.exchange, c.topic)
		if err != nil {
			return err
		}

		router.AddConsumerHandler(c.name, c.topic, sub, c.handler).AddMiddleware(
			TraceIDMiddleware,
			LoggingMiddleware(h.logger),
			NewRetryMiddleware().Middleware,
			poison,
			middleware.NewThrottle(100, time.Second).Middleware,
			middleware.Timeout(time.Second*30),
		)
	}

	h.logger.Info("AMQP_PIPELINE_READY", "queue", DeliveryProcessorQueue)
	return nil
}
