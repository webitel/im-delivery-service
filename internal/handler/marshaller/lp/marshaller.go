package lpmarshaller

import (
	"encoding/json"

	"github.com/webitel/im-delivery-service/internal/domain/model"
)

// LPEvent represents a single event structured for long-polling consumers.
type LPEvent struct {
	Type    string `json:"type"`
	ID      string `json:"id"`
	Payload any    `json:"payload"`
}

// Response defines the top-level JSON array to support event batching.
type Response struct {
	Events []LPEvent `json:"events"`
}

// MarshallEvents converts a slice of domain events into a single JSON batch.
func MarshallEvents(events []model.Eventer) ([]byte, error) {
	res := Response{
		Events: make([]LPEvent, 0, len(events)),
	}

	for _, ev := range events {
		lpEv := LPEvent{
			ID:      ev.GetID(),
			Payload: ev.GetPayload(),
		}

		// Map domain payload types to string identifiers for the frontend.
		switch ev.GetPayload().(type) {
		case *model.Message:
			lpEv.Type = "message_created"
		case *model.ConnectedPayload:
			lpEv.Type = "system_connected"
		default:
			lpEv.Type = "unknown"
		}
		res.Events = append(res.Events, lpEv)
	}

	return json.Marshal(res)
}
