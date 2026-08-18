package marshaller

import (
	"github.com/google/uuid"

	"github.com/webitel/im-delivery-service/internal/domain/event"
)

// EventMarshaller is a generic contract for all transport protocols. viewer is
// the contact id of the recipient this event is being serialized for; it lets
// per-viewer fields (e.g. a reaction's reacted_by_me) be derived at marshal
// time. Pass uuid.Nil when the recipient is irrelevant (system events).
type EventMarshaller interface {
	// Marshal single event (for WS, gRPC)
	Marshal(ev event.Eventer, viewer uuid.UUID) (any, error)
}

// BatchMarshaller is for protocols that support multiple events at once (like LP).
type BatchMarshaller interface {
	EventMarshaller
	MarshalBatch(events []event.Eventer, viewer uuid.UUID) (any, error)
}
