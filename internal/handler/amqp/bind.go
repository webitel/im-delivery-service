package amqp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"runtime/debug"

	"github.com/ThreeDotsLabs/watermill/message"

	"github.com/webitel/im-delivery-service/internal/domain/event"
)

// [DOMAIN_HANDLER] Defines a generic function that processes a payload into multiple events.
type DomainHandler[T any] func(ctx context.Context, payload *T) ([]event.Eventer, error)

func (h *MessageHandler) Dispatch(ctx context.Context, events []event.Eventer) {
	for i, ev := range events {
		if ev == nil {
			continue
		}

		if h.logger.Enabled(ctx, slog.LevelDebug) {
			h.logger.Debug("LOCAL_DELIVERY",
				"kind", ev.GetKindName(),
				"recipient", ev.GetUserID(),
				"ws_connected", h.hub.Connected(ev.GetUserID()),
			)
		}

		// [LOCAL_DELIVERY] WebSocket / gRPC / Long-Polling dispatch for connected clients.
		h.orchestrator.Notify(ctx, ev)

		// [GLOBAL_REPLICATION] RabbitMQ publish for cross-instance delivery and fan-out.
		if i == 0 {
			if err := h.dist.Publish(ctx, ev); err != nil {
				h.logger.Error(
					"GLOBAL_DISPATCH_FAILED",
					"err", err,
					"user_id", ev.GetUserID(),
					"kind", ev.GetKind(),
				)
			}
		}
	}
}

// Bind creates a Watermill handler with automatic decoding, recovery, and dispatching.
func Bind[T any](h *MessageHandler, fn DomainHandler[T]) message.NoPublishHandlerFunc {
	return func(msg *message.Message) error {
		defer func() {
			if r := recover(); r != nil {
				h.logger.Error("PANIC_RECOVERED", "err", r, "stack", string(debug.Stack()))
			}
		}()

		ctx := event.ContextWithMetadata(msg.Context(), msg.Metadata)

		payload := new(T)
		if err := json.Unmarshal(msg.Payload, payload); err != nil {
			h.logger.Error("payload_unmarshal_failed", "err", err, "raw", string(msg.Payload))

			return nil
		}

		if h.logger.Enabled(msg.Context(), slog.LevelDebug) {
			var pretty bytes.Buffer

			payloadType := fmt.Sprintf("%T", *payload)

			const (
				colorBlue  = "\033[1;34m"
				colorCyan  = "\033[0;36m"
				colorReset = "\033[0m"
				colorGray  = "\033[0;90m"
			)

			if err := json.Indent(&pretty, msg.Payload, "", "  "); err == nil {
				fmt.Printf("\n%s--- EVENT RECEIVED ---%s\n", colorBlue, colorReset)
				fmt.Printf("%sType:%s    %s\n", colorCyan, colorReset, payloadType)
				fmt.Printf("%sPayload:%s\n%s%s%s\n", colorCyan, colorReset, colorGray, pretty.String(), colorReset)
				fmt.Printf("%s------------------------------%s\n\n", colorBlue, colorReset)
			} else {
				h.logger.Debug("EVENT_RECEIVED", "type", payloadType, "raw", string(msg.Payload))
			}
		}

		events, err := fn(ctx, payload)
		if err != nil {
			return err
		}

		h.Dispatch(ctx, events)

		return nil
	}
}
