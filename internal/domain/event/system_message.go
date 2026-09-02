package event

import "github.com/webitel/im-delivery-service/internal/domain/model"

// SystemMessageType returns the model.System.Type carried by a MessageCreated
// chat "system" message event (e.g. "user_joined", "user_left"), and ok=true
// only when ev is such an event. Any other event kind, or a MessageCreated
// event whose payload has no System block, or a System block with an empty Type,
// returns ok=false so callers treat it as "not subject to system-message filtering"
// and deliver it unconditionally. An empty Type can never match any app's allow-list,
// so filtering would silently drop the message for restricted apps -- instead, we
// treat it as an unfiltered message to avoid data loss.
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
