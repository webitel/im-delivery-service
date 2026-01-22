package pubsub

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/webitel/im-delivery-service/internal/domain/event"
)

type EventDispatcher interface {
	Publish(ctx context.Context, ev event.Eventer) error
	Publisher() message.Publisher
}

type eventDispatcher struct {
	publisher message.Publisher
}

func NewEventDispatcher(pub message.Publisher) EventDispatcher {
	return &eventDispatcher{publisher: pub}
}

func (d *eventDispatcher) Publish(ctx context.Context, ev event.Eventer) error {
	if ev == nil {
		return nil
	}

	// [CONTRACT] Check if the event is meant for external AMQP delivery
	exportable, ok := ev.(event.Exportable)
	if !ok {
		return nil
	}

	payload, err := json.Marshal(ev)
	if err != nil {
		return fmt.Errorf("dispatcher: marshal error: %w", err)
	}

	// [ENVELOPE] Create a clean message without Watermill infrastructure noise
	msg := message.NewMessage(watermill.NewUUID(), payload)
	msg.SetContext(ctx)

	// [ROUTING] The first argument is the Routing Key.
	// In your Factory, GenerateRoutingKey: func(s string) string { return s }
	// so the routing key will be exactly what 'exportable.GetRoutingKey()' returns.
	if err := d.publisher.Publish(exportable.GetRoutingKey(), msg); err != nil {
		return fmt.Errorf("dispatcher: publish failed: %w", err)
	}

	return nil
}

func (d *eventDispatcher) Publisher() message.Publisher { return d.publisher }
