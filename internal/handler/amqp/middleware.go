package amqp

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/google/uuid"
	"github.com/webitel/im-delivery-service/internal/domain/registry"
)

// bind wraps the domain handler to provide routing key extraction and node-local filtering.
func bind[T any](hub registry.Hubber, fn func(context.Context, uuid.UUID, *T) error) message.NoPublishHandlerFunc {
	return func(msg *message.Message) error {
		// 1. Try to extract the routing key to identify the recipient
		rk := msg.Metadata.Get("x-routing-key")
		if rk == "" {
			rk = msg.Metadata.Get("routing_key")
		}

		// If no routing key is present, we cannot route the message
		if rk == "" {
			return nil // Ack: ignore messages without routing info
		}

		// Expecting routing key format: im_message.{user_id}.message.created.v1
		parts := strings.Split(rk, ".")
		if len(parts) < 2 {
			return nil // Ack: invalid topic structure
		}

		userID, err := uuid.Parse(parts[1])
		if err != nil {
			return nil // Ack: malformed UUID
		}

		// 2. [LOCALITY_CHECK] Only process the message if the user is connected to THIS node.
		// Since we use fan-out, every node gets the message, but only the owner node handles it.
		if !hub.IsConnected(userID) {
			return nil // Ack: user not on this node
		}

		// 3. Decode the message payload
		payload := new(T)
		if err := json.Unmarshal(msg.Payload, payload); err != nil {
			// Nack/Retry could be used here, but for now we just log and Ack to avoid poison pills
			return nil
		}

		// 4. Pass the data to the service handler
		if err := fn(msg.Context(), userID, payload); err != nil {
			// If service logic fails, return the error to trigger Watermill's retry/nack mechanism
			return err
		}

		return nil // Success: Automatic Ack
	}
}
