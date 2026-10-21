package amqp

import (
	"context"

	"github.com/google/uuid"

	"github.com/webitel/im-delivery-service/internal/domain/event"
	"github.com/webitel/im-delivery-service/internal/domain/util"
	"github.com/webitel/im-delivery-service/internal/handler/amqp/payload"
)

// OnThreadTagsChangedV1 sends a tag change to its owner's sessions only: tags are private.
func (h *MessageHandler) OnThreadTagsChangedV1(_ context.Context, raw *payload.ThreadTagsChangedV1) ([]event.Eventer, error) {
	owner := util.SafeParseUUID(raw.ContactID)
	update := raw.ToDomain()

	if owner == uuid.Nil || update.ThreadID == uuid.Nil {
		return nil, nil
	}

	targets := h.computeLocalTargets(uuid.Nil, []uuid.UUID{owner})
	if len(targets) == 0 {
		return nil, nil
	}

	return []event.Eventer{event.NewThreadTagsChangedEvent(update, owner, int64(raw.DomainID))}, nil
}
