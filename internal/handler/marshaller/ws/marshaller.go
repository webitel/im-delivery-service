package wsmarshaller

import (
	"encoding/json"

	"github.com/webitel/im-delivery-service/internal/domain/event"
	"github.com/webitel/im-delivery-service/internal/domain/model"
)

// WSEvent is a generic wrapper for WebSocket messages to provide consistent structure.
type WSEvent struct {
	Event   string `json:"event"`   // [TYPE] e.g., "message_created", "connected", "disconnected"
	ID      string `json:"id"`      // [IDENTITY] Unique event ID
	SentAt  int64  `json:"sent_at"` // [TIMESTAMP] Unix milliseconds
	Payload any    `json:"payload"` // [DATA] Specific event content
}

// MarshallDeliveryEvent prepares data for WebSocket transmission using JSON.
func MarshallDeliveryEvent(ev event.Eventer) ([]byte, error) {
	// [INITIALIZATION] Create the base envelope
	res := &WSEvent{
		ID:     ev.GetID(),
		SentAt: ev.GetOccurredAt(),
	}

	// [PAYLOAD_MAPPING] Route domain models to JSON-friendly structures
	switch p := ev.GetPayload().(type) {

	case *model.Message:
		res.Event = "message_created"
		res.Payload = mapMessage(p)

	case *model.ConnectedPayload:
		// [SYSTEM_EVENT] Handshake success
		res.Event = "connected"
		res.Payload = p

	case *model.DisconnectedPayload:
		// [SYSTEM_EVENT] Termination notice
		res.Event = "disconnected"
		res.Payload = p

	default:
		// [FALLBACK] Use the event's own kind string if not explicitly mapped
		res.Event = ev.GetKind().String()
		res.Payload = p
	}

	return json.Marshal(res)
}
