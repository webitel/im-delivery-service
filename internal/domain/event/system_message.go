package event

import "github.com/webitel/im-delivery-service/internal/domain/model"

// SystemMessageType returns the model.System.Type carried by a MessageCreated
// chat "system" message event, and ok=true only when ev is such an event.
// An empty Type also returns ok=false: it can never match any app's allow-list,
// so treating it as filterable would silently drop the message instead of
// delivering it.
func SystemMessageType(ev Eventer) (string, bool) {
	if ev.GetKind() != MessageCreated {
		return "", false
	}

	msg, ok := ev.GetPayload().(*model.Message)
	if !ok || msg == nil || msg.System == nil || msg.System.Type == "" {
		return "", false
	}

	return msg.System.Type, true
}
