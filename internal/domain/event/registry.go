// internal/domain/event/registry.go
package event

import (
	"github.com/webitel/im-delivery-service/internal/domain/model"
)

// [NEW_ENVELOPE_FOR_KIND] Polymorphic factory for event restoration from persistence.
// It maps EventKind to a concrete Envelope implementation to ensure correct JSON unmarshaling.
func NewEnvelopeForKind(kind EventKind) Eventer {
	switch kind {
	case MessageCreated:
		// [CONCRETE_TYPE] Restores as Envelope[*model.Message].
		// This is critical for the Notifier interface to work in push handlers.
		return &Envelope[*model.Message]{}

	case ThreadCreated:
		return &Envelope[*model.Thread]{}

	case MessageRead:
		return &Envelope[*model.MessageReadPayload]{}

	default:
		// [FALLBACK] Use generic any-envelope for unknown or raw event kinds.
		return &Envelope[any]{}
	}
}
