package amqp

import (
	"context"
	"encoding/json"
	"runtime/debug"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/webitel/im-delivery-service/internal/domain/event"
)

// [DOMAIN_HANDLER] Defines a generic function that processes a payload into multiple events.
type DomainHandler[T any] func(ctx context.Context, payload *T) ([]event.Eventer, error)

func (h *MessageHandler) Dispatch(ctx context.Context, events []event.Eventer) {
	for _, ev := range events {
		// [GUARD] Skip empty events to prevent downstream nil pointer dereferences.
		if ev == nil {
			continue
		}

		// [LOCAL_DELIVERY] Always notify local sessions via the sharded Hub.
		// This is non-blocking and handles direct WebSocket communication.
		h.hub.Broadcast(ev)

		// [GLOBAL_DELIVERY] Check if the event satisfies the Exportable interface.
		// ---------------------------------------------------------------------------------
		// [LOGIC]
		// If 'ok' is true, the event is an ExportableEnvelope (e.g., MessageCreated).
		// If 'ok' is false, it's a base Envelope (e.g., ThreadCreated, System signals),
		// which are strictly node-local and should NOT be published to the global bus.
		// ---------------------------------------------------------------------------------
		if _, ok := ev.(event.Exportable); ok {
			// Publish to RabbitMQ via the atomic proxy (checked for leader/active node inside).
			if err := h.dispatcher.Publish(ctx, ev); err != nil {
				h.logger.Error("GLOBAL_DISPATCH_FAILED",
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
		p := json.Unmarshal(msg.Payload, payload)
		println(p)
		if err := json.Unmarshal(msg.Payload, payload); err != nil {
			return nil // ACK invalid payloads
		}

		events, err := fn(msg.Context(), payload)
		if err != nil {
			return err // NACK for retries
		}

		h.Dispatch(msg.Context(), events)
		return nil
	}
}
