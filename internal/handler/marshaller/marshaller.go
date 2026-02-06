package marshaller

import "github.com/webitel/im-delivery-service/internal/domain/event"

// EventMarshaller is a generic contract for all transport protocols.
type EventMarshaller interface {
	// Marshal single event (for WS, gRPC)
	Marshal(ev event.Eventer) (any, error)
}

// BatchMarshaller is for protocols that support multiple events at once (like LP).
type BatchMarshaller interface {
	EventMarshaller
	MarshalBatch(events []event.Eventer) (any, error)
}
