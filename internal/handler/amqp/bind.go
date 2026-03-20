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
		if ev == nil {
			continue
		}

		// [DELIVERY_ORCHESTRATION]
		// This single call now handles:
		// 1. Socket broadcast to local cells.
		// 2. Redis tracking (if EnablePush is true).
		// 3. Background wait for ACK.
		// 4. Push fallback.
		h.orchestrator.Notify(ctx, ev)

		// [GLOBAL_REPLICATION]
		// Dispatcher decides internally whether the event is routable.
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
