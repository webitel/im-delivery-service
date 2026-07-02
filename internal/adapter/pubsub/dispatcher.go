package pubsub

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

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
	logger    *slog.Logger
}

func NewEventDispatcher(pub message.Publisher, logger *slog.Logger) EventDispatcher {
	return &eventDispatcher{
		publisher: pub,
		logger:    logger,
	}
}

func (d *eventDispatcher) IsLeader() bool               { return true }
func (d *eventDispatcher) Publisher() message.Publisher { return d.publisher }

func (d *eventDispatcher) Publish(ctx context.Context, ev event.Eventer) error {
	if ev == nil {
		return nil
	}

	// [ROUTABLE_CHECK] Only payloads that define routing are exported.
	routable, ok := ev.GetPayload().(event.Routable)
	if !ok {
		return nil
	}

	// [ROUTING_KEY] Resolve routing from payload.
	routingKey := routable.RoutingKey()
	if routingKey == "" {
		return nil
	}

	payload, err := json.Marshal(ev)
	if err != nil {
		return fmt.Errorf("dispatcher: marshal error: %w", err)
	}

	// [DEBUG_PRETTY_PRINT] Colorized output for outgoing RabbitMQ messages.
	if d.logger != nil && d.logger.Enabled(ctx, slog.LevelDebug) {
		var pretty bytes.Buffer
		if err := json.Indent(&pretty, payload, "", "  "); err == nil {
			// ANSI Escape Codes for professional terminal styling
			const (
				colorPurple = "\033[1;35m"
				colorYellow = "\033[1;33m"
				colorCyan   = "\033[0;36m"
				colorReset  = "\033[0m"
				colorGray   = "\033[0;90m"
			)

			fmt.Printf("\n%s>>> [OUTGOING] PUBLISHING TO RABBITMQ <<<%s\n", colorPurple, colorReset)
			fmt.Printf("%sTarget Exchange/Key:%s %s%s%s\n", colorCyan, colorReset, colorYellow, routingKey, colorReset)
			fmt.Printf("%sPayload:%s\n%s%s%s\n", colorCyan, colorReset, colorGray, pretty.String(), colorReset)
			fmt.Printf("%s%s%s\n\n", colorPurple, strings.Repeat("-", 40), colorReset)
		}
	}

	// [MESSAGE_ENVELOPE] Create AMQP message.
	msg := message.NewMessage(watermill.NewUUID(), payload)
	msg.SetContext(ctx)

	if metadata, ok := event.TryGetMetadataFromContext(ctx); ok {
		for k, v := range metadata { // TODO: add headers validating to skip unnecessary
			msg.Metadata.Set(k, v)
		}
	}

	if err := d.publisher.Publish(routingKey, msg); err != nil {
		return fmt.Errorf("dispatcher: publish failed: %w", err)
	}

	return nil
}
