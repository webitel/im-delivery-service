package amqp

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/google/uuid"
	leader "github.com/webitel/im-delivery-service/infra/discovery/consul"
	"github.com/webitel/im-delivery-service/internal/adapter/pubsub"
	"github.com/webitel/im-delivery-service/internal/domain/registry"
	"github.com/webitel/im-delivery-service/internal/service"
)

const (
	// Exchanges
	MessageEventsExchange = "im_message.events"
	SystemEventsExchange  = "im_system.events"

	// Routing Keys
	TopicMessageCreated   = "im_message.#.message.created.v1"
	TopicThreadCreated    = "im_thread.#.thread.created.v1"
	TopicDeviceRegister   = "updates.device.register.#"
	TopicDeviceUnregister = "updates.device.unregister.#"
	TopicDeviceLogout     = "updates.device.logout.#"
	TopicVariablesSet     = "im_thread.#.variables.set.#"
	TopicVariablesFlush   = "im_thread.#.variables.flush.#"

	// Queues
	DeliveryProcessorQueue = "im-delivery.incoming-processor.v1"
	DeliveryPoisonTopic    = "im-delivery.incoming-processor.v1.poison"
)

// Pipeline defines a specific AMQP message flow configuration.
type Pipeline struct {
	Name     string
	Exchange string
	Topic    string
	Handler  message.NoPublishHandlerFunc
}

type MessageHandler struct {
	orchestrator   service.Orchestrator
	hub            registry.Hubber
	logger         *slog.Logger
	enricher       service.Contacter
	leader         leader.LeaderAwarer
	dist           pubsub.EventDispatcher
	deviceProvider service.DeviceProvider
}

func NewMessageHandler(
	orchestrator service.Orchestrator,
	hub registry.Hubber,
	logger *slog.Logger,
	enricher service.Contacter,
	leader leader.LeaderAwarer,
	dist pubsub.EventDispatcher,
	deviceProvider service.DeviceProvider,
) *MessageHandler {
	return &MessageHandler{
		orchestrator:   orchestrator,
		hub:            hub,
		logger:         logger,
		enricher:       enricher,
		leader:         leader,
		dist:           dist,
		deviceProvider: deviceProvider,
	}
}

// RegisterHandlers sets up all AMQP consumers with unique transient queues.
func (h *MessageHandler) RegisterHandlers(router *message.Router, subProvider *pubsub.SubscriberProvider) error {
	poison, err := middleware.PoisonQueue(h.dist.Publisher(), DeliveryPoisonTopic)
	if err != nil {
		return fmt.Errorf("poison_setup_failed: %w", err)
	}

	// Single Node ID for all handlers in this session
	nodeID := uuid.NewString()[:8]

	// 1. Declarative pipeline configuration
	pipelines := []Pipeline{
		{
			Name:     "ON_MESSAGE_CREATED",
			Exchange: MessageEventsExchange,
			Topic:    TopicMessageCreated,
			Handler:  Bind(h, h.OnMessageCreatedV1),
		},
		{
			Name:     "ON_THREAD_CREATED",
			Exchange: MessageEventsExchange,
			Topic:    TopicThreadCreated,
			Handler:  Bind(h, h.OnThreadCreatedV1),
		},
		{
			Name:     "ON_DEVICE_REGISTERED",
			Exchange: SystemEventsExchange,
			Topic:    TopicDeviceRegister,
			Handler:  Bind(h, h.OnDeviceRegisteredV1),
		},
		{
			Name:     "ON_DEVICE_UNREGISTERED",
			Exchange: SystemEventsExchange,
			Topic:    TopicDeviceUnregister,
			Handler:  Bind(h, h.OnDeviceUnregisteredV1),
		},
		{
			Name:     "ON_DEVICE_LOGOUT",
			Exchange: SystemEventsExchange,
			Topic:    TopicDeviceLogout,
			Handler:  Bind(h, h.OnDeviceLogoutV1),
		},
		{
			Name:     "ON_VARIABLES_SET",
			Exchange: MessageEventsExchange,
			Topic:    TopicVariablesSet,
			Handler:  Bind(h, h.OnVariablesSetV1),
		},
		{
			Name:     "ON_VARIABLES_FLUSH",
			Exchange: MessageEventsExchange,
			Topic:    TopicVariablesFlush,
			Handler:  Bind(h, h.OnVariablesFlushV1),
		},
	}

	// 2. Iterate and register
	for _, p := range pipelines {
		queueName := h.fmtQueueName(nodeID, p.Name)

		subscriber, err := subProvider.Build(queueName, p.Exchange, p.Topic)
		if err != nil {
			return fmt.Errorf("subscriber_build_failed [%s]: %w", p.Name, err)
		}

		router.AddConsumerHandler(
			p.Name,
			p.Topic,
			subscriber,
			p.Handler,
		).AddMiddleware(
			poison,
			middleware.NewThrottle(100, time.Second).Middleware,
			middleware.Timeout(30*time.Second),
		)
	}

	h.logger.Info("amqp_pipelines_ready", "node_id", nodeID, "count", len(pipelines))
	return nil
}

// fmtQueueName generates a unique queue name for the current instance.
func (h *MessageHandler) fmtQueueName(nodeID, name string) string {
	return fmt.Sprintf("%s.%s.%s", DeliveryProcessorQueue, nodeID, name)
}
