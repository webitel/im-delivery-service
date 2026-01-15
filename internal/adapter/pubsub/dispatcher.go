package pubsub

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/webitel/im-delivery-service/internal/domain/model"
)

type EventDispatcher struct {
	publisher message.Publisher
}

func NewEventDispatcher(pub message.Publisher) *EventDispatcher {
	return &EventDispatcher{publisher: pub}
}

func (d *EventDispatcher) Publish(ctx context.Context, event model.OutboundEventer) error {
	if event == nil {
		return fmt.Errorf("event dispatcher: cannot publish nil event")
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("event dispatcher: marshal failure: %w", err)
	}

	msg := message.NewMessage(watermill.NewUUID(), payload)
	msg.SetContext(ctx)

	topic := event.GetRoutingKey()

	if err := d.publisher.Publish(topic, msg); err != nil {
		return fmt.Errorf("event dispatcher: failed to publish to topic %s: %w", topic, err)
	}

	return nil
}
