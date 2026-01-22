// internal/adapter/pubsub/event_dispatcher.go

package pubsub

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/webitel/im-delivery-service/internal/domain/event"
)

// EventDispatcher defines the high-level contract for outgoing events.
// This allows the handler to stay agnostic of the transport implementation.
type EventDispatcher interface {
	Publish(ctx context.Context, ev event.Eventer) error
	Publisher() message.Publisher
}

// eventDispatcher is the concrete implementation (private).
type eventDispatcher struct {
	publisher message.Publisher
}

// NewEventDispatcher returns the interface instead of the pointer to the struct.
func NewEventDispatcher(pub message.Publisher) EventDispatcher {
	return &eventDispatcher{
		publisher: pub,
	}
}

func (d *eventDispatcher) Publish(ctx context.Context, ev event.Eventer) error {
	if ev == nil {
		return fmt.Errorf("event dispatcher: cannot publish nil event")
	}

	payload, err := json.Marshal(ev)
	if err != nil {
		return fmt.Errorf("event dispatcher: marshal failure: %w", err)
	}

	msg := message.NewMessage(watermill.NewUUID(), payload)
	msg.SetContext(ctx)

	fmt.Printf("Publishing event to topic: %s\n", ev.GetRoutingKey())
	if err := d.publisher.Publish(ev.GetRoutingKey(), msg); err != nil {
		return fmt.Errorf("event dispatcher: failed to publish to topic %s: %w", ev.GetRoutingKey(), err)
	}

	return nil
}

func (d *eventDispatcher) Publisher() message.Publisher {
	return d.publisher
}
