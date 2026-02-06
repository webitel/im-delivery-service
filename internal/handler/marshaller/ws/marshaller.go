package wsmarshaller

import (
	"encoding/json"

	"github.com/webitel/im-delivery-service/internal/domain/event"
	"github.com/webitel/im-delivery-service/internal/domain/model"
	"github.com/webitel/im-delivery-service/internal/handler/marshaller"
)

// [INTERFACE_GUARD] Ensure Marshaller implements the global interface.
var _ marshaller.EventMarshaller = (*Marshaller)(nil)

type Marshaller struct{}

func New() *Marshaller { return &Marshaller{} }

// Marshal transforms a domain event into JSON bytes, utilizing cache if available.
func (m *Marshaller) Marshal(ev event.Eventer) (any, error) {
	// 1. [PERFORMANCE] Check if this event was already marshaled for WS.
	// Since different protocols (WS vs gRPC) have different binary outputs,
	// we should ideally use a protocol-specific cache key or check the type.
	if cached := ev.GetCached(); cached != nil {
		if data, ok := cached.([]byte); ok {
			return data, nil
		}
	}

	// 2. [ENVELOPE] Initialize the transport-specific structure.
	res := &ServerEvent{
		ID:        ev.GetID(),
		CreatedAt: ev.GetOccurredAt(),
		Priority:  PriorityHigh, // Or map from ev.GetPriority()
		Payload:   make(map[string]any),
	}

	// 3. [STRATEGY] Map domain payload to transport schema.
	switch p := ev.GetPayload().(type) {
	case *model.Message:
		res.Payload[EventMessage] = mapMessage(p)
	case *model.ConnectedPayload:
		res.Payload[EventConnected] = p
	case *model.DisconnectedPayload:
		res.Payload[EventDisconnected] = p
	default:
		// Fallback for system or unknown events.
		res.Payload[ev.GetKind().String()] = p
	}

	// 4. [SERIALIZATION] Convert to JSON bytes.
	data, err := json.Marshal(res)
	if err != nil {
		return nil, err
	}

	// 5. [CACHE] Store the serialized JSON for reuse by other concurrent WS streams.
	ev.SetCached(data)

	return data, nil
}
