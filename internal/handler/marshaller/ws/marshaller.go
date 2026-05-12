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
	// 1. [PERFORMANCE] Check cache
	if cached := ev.GetCached(); cached != nil {
		if data, ok := cached.([]byte); ok {
			return data, nil
		}
	}

	// 2. [ENVELOPE] Initialize with prioritized delivery metadata.
	res := &ServerEvent{
		ID:        ev.GetID(),
		CreatedAt: ev.GetOccurredAt(),
		Priority:  PriorityHigh,
		Payload:   make(map[string]any),
	}

	// 3. [STRATEGY] Map domain payload to transport schema.
	switch p := ev.GetPayload().(type) {
	case *model.Message:
		res.Payload[EventMessage.String()] = mapMessage(p)

	case *model.ConnectedPayload:
		res.Payload[EventConnected.String()] = p

	case *model.DisconnectedPayload:
		res.Payload[EventDisconnected.String()] = p

	case *model.Thread:
		res.Payload[EventThreadCreated.String()] = mapThread(p)

	case *model.VariablesPayload:
		if ev.GetKind() == event.VariableFlush {
			res.Payload["variable_flush_event"] = mapVariables(p)
		} else {
			res.Payload["variable_set_event"] = mapVariables(p)
		}

	case *model.MemberEvent:
		if ev.GetKind() == event.MemberLeft {
			res.Payload[EventMemberLeft.String()] = p
		} else {
			res.Payload[EventMemberAdded.String()] = p
		}

	default:
		// [FALLBACK] Use the event kind name for unknown types.
		res.Payload[ev.GetKind().String()] = p
	}

	// 4. [SERIALIZATION]
	data, err := json.Marshal(res)
	if err != nil {
		return nil, err
	}

	// 5. [CACHE] Store for reuse across concurrent WebSocket streams.
	ev.SetCached(data)

	return data, nil
}
