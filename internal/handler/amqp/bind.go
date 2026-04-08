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

// Dispatch orchestrates the delivery of generated events.
// It handles local delivery (WebSockets/Cells) for every target participant,
// while ensuring global replication (RabbitMQ) happens only once per message set
// to prevent redundant event storms across the cluster.
func (h *MessageHandler) Dispatch(ctx context.Context, events []event.Eventer) {
	for i, ev := range events {
		if ev == nil {
			continue
		}

		// [LOCAL_DELIVERY]
		// Notify the local orchestrator for every event in the slice.
		// This ensures that if multiple participants (e.g., sender and receiver)
		// are connected to THIS specific service instance, they all receive
		// their respective socket updates and push notifications.
		h.orchestrator.Notify(ctx, ev)

		// [GLOBAL_REPLICATION]
		// To keep the message bus clean and efficient, we only publish to
		// the distributed broker (RabbitMQ) ONCE per batch.
		// Other nodes in the cluster will receive this single event and
		// trigger their own local Dispatch logic for their connected clients.
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

		payload := new(T)
		if err := json.Unmarshal(msg.Payload, payload); err != nil {
			h.logger.Error("payload_unmarshal_failed", "err", err, "raw", string(msg.Payload))
			return nil // ACK invalid payloads
		}

		if h.logger.Enabled(msg.Context(), slog.LevelDebug) {
			var pretty bytes.Buffer
			payloadType := fmt.Sprintf("%T", *payload)

			// ANSI Escape Codes for colors
			const (
				colorBlue  = "\033[1;34m"
				colorCyan  = "\033[0;36m"
				colorReset = "\033[0m"
				colorGray  = "\033[0;90m"
			)

			if err := json.Indent(&pretty, msg.Payload, "", "  "); err == nil {
				// We use colors to distinguish the debug block from regular logs
				fmt.Printf("\n%s--- EVENT RECEIVED ---%s\n", colorBlue, colorReset)
				fmt.Printf("%sType:%s    %s\n", colorCyan, colorReset, payloadType)
				fmt.Printf("%sPayload:%s\n%s%s%s\n", colorCyan, colorReset, colorGray, pretty.String(), colorReset)
				fmt.Printf("%s------------------------------%s\n\n", colorBlue, colorReset)
			} else {
				h.logger.Debug("EVENT_RECEIVED", "type", payloadType, "raw", string(msg.Payload))
			}
		}

		events, err := fn(msg.Context(), payload)
		if err != nil {
			return err // NACK for retries
		}

		h.Dispatch(msg.Context(), events)
		return nil
	}
}
