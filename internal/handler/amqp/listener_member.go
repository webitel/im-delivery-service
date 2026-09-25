package amqp

import (
	"context"

	"github.com/google/uuid"

	"github.com/webitel/im-delivery-service/internal/domain/event"
	"github.com/webitel/im-delivery-service/internal/handler/amqp/payload"
)

// OnMemberAddedV1 handles member_added system events.
func (h *MessageHandler) OnMemberAddedV1(ctx context.Context, raw *payload.MemberEventV1) ([]event.Eventer, error) {
	return h.handleMemberEvent(raw, event.MemberAdded)
}

// OnMemberLeftV1 handles member_left system events.
func (h *MessageHandler) OnMemberLeftV1(ctx context.Context, raw *payload.MemberEventV1) ([]event.Eventer, error) {
	return h.handleMemberEvent(raw, event.MemberLeft)
}

func (h *MessageHandler) handleMemberEvent(raw *payload.MemberEventV1, kind event.EventKind) ([]event.Eventer, error) {
	m := raw.ToDomain()

	if m.ContactID == uuid.Nil || m.ThreadID == uuid.Nil {
		return nil, nil
	}

	// Every member sees the join/leave live, not only the subject.
	targets := h.computeLocalTargets(m.ContactID, memberEventRecipients(m.ContactID, raw.Participants))
	if len(targets) == 0 {
		return nil, nil
	}

	//nolint:exhaustive // Callers only route MemberAdded/MemberLeft.
	switch kind {
	case event.MemberAdded:
		m.Action = "joined"
	case event.MemberLeft:
		m.Action = "left"
	}

	events := make([]event.Eventer, 0, len(targets))

	for _, target := range targets {
		copied := *m
		events = append(events, event.NewMemberEvent(&copied, target, kind))
	}

	return events, nil
}

func memberEventRecipients(subject uuid.UUID, participants []string) []uuid.UUID {
	seen := map[uuid.UUID]struct{}{subject: {}}
	res := []uuid.UUID{subject}

	for _, raw := range participants {
		id, err := uuid.Parse(raw)
		if err != nil || id == uuid.Nil {
			continue
		}

		if _, ok := seen[id]; !ok {
			seen[id] = struct{}{}
			res = append(res, id)
		}
	}

	return res
}
