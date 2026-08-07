package lpmarshaller

import (
	"encoding/json"

	"github.com/webitel/im-delivery-service/internal/domain/event"
	"github.com/webitel/im-delivery-service/internal/domain/model"
	"github.com/webitel/im-delivery-service/internal/handler/marshaller"
)

// [INTERFACE_GUARD]
var _ marshaller.BatchMarshaller = (*Marshaller)(nil)

type LPEvent struct {
	Type    string `json:"type"`
	ID      string `json:"id"`
	Payload any    `json:"payload"`
}

type Response struct {
	Events []LPEvent `json:"events"`
}

type Marshaller struct{}

func New() *Marshaller { return &Marshaller{} }

// Marshal returns a single LPEvent.
// Note: We return the structure itself here to make MarshalBatch more efficient.
func (m *Marshaller) Marshal(ev event.Eventer) (any, error) {
	// [PERFORMANCE] Check for cached LPEvent structure
	if cached := ev.GetCached(); cached != nil {
		if lp, ok := cached.(LPEvent); ok {
			return lp, nil
		}
	}

	lp := LPEvent{
		ID:      ev.GetID(),
		Payload: ev.GetPayload(),
	}

	// [STRATEGY] Map domain types to LP type strings
	switch ev.GetPayload().(type) {
	case *model.Message:
		lp.Type = "message_created"
	case *model.MessageDeleted:
		lp.Type = "message_deleted"
	case *model.MessageReaction:
		lp.Type = "message_reaction"
	case *model.ConnectedPayload:
		lp.Type = "system_connected"
	case *model.DisconnectedPayload:
		lp.Type = "system_disconnected"
	case *model.Typing:
		lp.Type = "typing_event"
	default:
		lp.Type = ev.GetKind().String()
	}

	// [CACHE] Store the structure for reuse in batches
	ev.SetCached(lp)

	return lp, nil
}

// MarshalBatch handles the collection of events and returns final JSON bytes.
func (m *Marshaller) MarshalBatch(events []event.Eventer) (any, error) {
	res := Response{
		Events: make([]LPEvent, 0, len(events)),
	}

	for _, ev := range events {
		val, err := m.Marshal(ev)
		if err != nil {
			continue
		}

		// [TYPE_ASSERTION] Ensure we got the structure
		if lp, ok := val.(LPEvent); ok {
			res.Events = append(res.Events, lp)
		}
	}

	// [SERIALIZATION] Final batch conversion to JSON
	return json.Marshal(res)
}
