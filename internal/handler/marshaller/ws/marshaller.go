package wsmarshaller

import (
	"encoding/json"

	"github.com/google/uuid"

	"github.com/webitel/im-delivery-service/internal/domain/event"
	"github.com/webitel/im-delivery-service/internal/domain/model"
	"github.com/webitel/im-delivery-service/internal/handler/marshaller"
)

// [INTERFACE_GUARD] Ensure Marshaller implements the global interface.
var _ marshaller.EventMarshaller = (*Marshaller)(nil)

type Marshaller struct{}

func New() *Marshaller { return &Marshaller{} }

// Marshal transforms a domain event into JSON bytes, utilizing cache if available.
func (m *Marshaller) Marshal(ev event.Eventer, viewer uuid.UUID) (any, error) {
	// A reaction carries a per-viewer reacted_by_me flag, so its bytes cannot be
	// shared across recipients — bypass the cross-stream cache for it.
	_, perViewer := ev.GetPayload().(*model.MessageReaction)

	// 1. [PERFORMANCE] Check cache
	if !perViewer {
		if cached := ev.GetCached(); cached != nil {
			if data, ok := cached.([]byte); ok {
				return data, nil
			}
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

	case *model.MessageEdited:
		res.Payload[EventMessageEdited.String()] = mapMessageEdited(p)

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
	case *model.InteractiveCallback:
		res.Payload[EventInteractiveCallback.String()] = p

	case *model.MessageStatusUpdate:
		res.Payload[EventMessageStatus.String()] = p

	case *model.MessageDeleted:
		res.Payload[EventMessageDeleted.String()] = mapMessageDeleted(p)

	case *model.MessageReaction:
		res.Payload[EventMessageReaction.String()] = mapReaction(p, viewer)

	case *model.Typing:
		res.Payload["typing_event"] = mapTyping(p)

	default:
		// [FALLBACK] Use the event kind name for unknown types.
		res.Payload[ev.GetKind().String()] = p
	}

	// 4. [SERIALIZATION]
	data, err := json.Marshal(res)
	if err != nil {
		return nil, err
	}

	// 5. [CACHE] Store for reuse across concurrent WebSocket streams. Skipped for
	// per-viewer payloads, whose bytes differ between recipients.
	if !perViewer {
		ev.SetCached(data)
	}

	return data, nil
}
