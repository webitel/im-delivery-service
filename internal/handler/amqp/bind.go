package amqp

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime/debug"
	"strings"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/google/uuid"
	"github.com/webitel/im-delivery-service/internal/domain/event"
)

// DomainHandler defines the functional signature for business logic.
type DomainHandler[T any] func(ctx context.Context, userID uuid.UUID, payload *T) (event.Eventer, error)

// [INFRASTRUCTURE_BRIDGE]
// Bind connects Watermill to Domain logic, handling Panic Recovery, Locality, and Fan-out.
func Bind[T any](h *MessageHandler, fn DomainHandler[T]) message.NoPublishHandlerFunc {
	return func(msg *message.Message) error {
		// [PANIC_RECOVERY]
		// Safely handle runtime panics to keep the consumer alive.
		defer func() {
			if r := recover(); r != nil {
				h.logger.Error("PANIC_RECOVERED",
					"err", r,
					"stack", string(debug.Stack()),
					"msg_id", msg.UUID)
			}
		}()

		// [IDENTIFICATION]
		// Extract recipient UUID from metadata for routing decisions.
		userID, ok := resolveUserID(msg)
		if !ok {
			h.logger.Warn("ROUTING_FAILED: recipient_missing", "msg_id", msg.UUID)
			return nil // ACK: Invalid routing is a terminal state.
		}

		// [LOCALITY_FILTER]
		// Distributed scaling: process only if the target user is connected to THIS node.
		if !h.hub.IsConnected(userID) {
			return nil // ACK: Handled by another instance.
		}

		// [DECODING]
		payload := new(T)
		if err := json.Unmarshal(msg.Payload, payload); err != nil {
			h.logger.Error("DECODE_FAILED", "err", err, "msg_id", msg.UUID)
			return nil // ACK: Poison Pill protection.
		}

		// [EXECUTION]
		// Domain logic execution with enriched context (TraceID).
		ev, err := fn(msg.Context(), userID, payload)
		if err != nil {
			return err // NACK: Business failure triggers Retry policy.
		}

		if ev == nil {
			return nil
		}

		// [FAN_OUT_DISPATCH]
		// 1. Local delivery (WebSockets/gRPC).
		h.hub.Broadcast(ev)

		// 2. Global delivery (RabbitMQ) for multi-node synchronization.
		if _, ok := ev.(event.Exportable); ok {
			if err := h.dispatcher.Publish(msg.Context(), ev); err != nil {
				return fmt.Errorf("GLOBAL_DISPATCH_FAILED: %w", err)
			}
		}

		return nil
	}
}

func resolveUserID(msg *message.Message) (uuid.UUID, bool) {
	rk := msg.Metadata.Get("x-routing-key")
	if rk == "" {
		rk = msg.Metadata.Get("routing_key")
	}

	for part := range strings.SplitSeq(rk, ".") {
		if uid, err := uuid.Parse(part); err == nil {
			return uid, true
		}
	}
	return uuid.Nil, false
}
