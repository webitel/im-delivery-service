package wsmarshaller

import (
	"encoding/json"

	"github.com/webitel/im-delivery-service/internal/domain/model"
)

// WSEvent is a generic wrapper for WebSocket messages to provide consistent structure
type WSEvent struct {
	Event   string `json:"event"` // e.g., "message_created", "connected"
	ID      string `json:"id"`    // message or event ID
	SentAt  int64  `json:"sent_at"`
	Payload any    `json:"payload"`
}

// MarshallDeliveryEvent prepares data for WebSocket transmission.
func MarshallDeliveryEvent(ev model.InboundEventer) ([]byte, error) {
	// We don't use gRPC cache here because WS uses JSON.
	// Instead, we map domain model to a friendly JSON structure.

	res := &WSEvent{
		ID:     ev.GetID(),
		SentAt: ev.GetOccurredAt(),
	}

	switch p := ev.GetPayload().(type) {
	case *model.Message:
		res.Event = "message_created"
		res.Payload = mapMessage(p)
	case *model.ConnectedPayload:
		res.Event = "connected"
		res.Payload = p
	}

	return json.Marshal(res)
}
