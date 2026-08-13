package amqp

import (
	"context"

	"github.com/google/uuid"

	"github.com/webitel/im-delivery-service/internal/domain/event"
	"github.com/webitel/im-delivery-service/internal/domain/model"
	"github.com/webitel/im-delivery-service/internal/domain/util"
	"github.com/webitel/im-delivery-service/internal/handler/amqp/payload"
)

// OnTypingV1 fans out an ephemeral "…is typing" indicator to the online
// sessions of the thread's recipients. It is real-time only: never persisted,
// never pushed (CanPush=false), never replayed on reconnect. The live-preview
// draft is attached only to recipients present in the allow-list.
func (h *MessageHandler) OnTypingV1(_ context.Context, raw *payload.TypingV1) ([]event.Eventer, error) {
	senderID, _ := uuid.Parse(raw.MemberID)

	recipients := parseTypingRecipients(raw.Recipients(), senderID)
	if len(recipients) == 0 {
		return nil, nil
	}

	// Only recipients connected to this node (the leader handles all).
	targets := h.computeLocalTargets(senderID, recipients)
	if len(targets) == 0 {
		return nil, nil
	}

	previewAllowed := make(map[uuid.UUID]struct{}, len(raw.PreviewVisibleTo))

	for _, id := range raw.PreviewVisibleTo {
		if uid, err := uuid.Parse(id); err == nil {
			previewAllowed[uid] = struct{}{}
		}
	}

	occurredAt := util.SafeParseRFC3339(raw.OccurredAt)

	events := make([]event.Eventer, 0, len(targets))

	for _, targetID := range targets {
		if targetID == senderID {
			continue // no self-echo for typing
		}

		// Populate member identity from the payload.
		var typingMember *model.TypingMember
		if raw.Member != nil {
			typingMember = &model.TypingMember{
				ID:     raw.MemberID,
				Name:   raw.Member.Name,
				Issuer: raw.Member.Issuer,
			}
		}

		t := &model.Typing{
			ThreadID:  raw.ThreadID,
			MemberID:  raw.MemberID,
			TimeoutMs: raw.TimeoutMS,
			Member:    typingMember,
		}

		// Privacy gate: attach the draft only for authorized recipients.
		if _, ok := previewAllowed[targetID]; ok {
			t.PreviewText = raw.PreviewText
		}

		events = append(events, event.NewTypingEvent(t, targetID, 0, occurredAt))
	}

	return events, nil
}

// parseTypingRecipients parses recipient contact ids, dropping the sender,
// duplicates and invalid ids.
func parseTypingRecipients(ids []string, sender uuid.UUID) []uuid.UUID {
	seen := make(map[uuid.UUID]struct{}, len(ids))
	out := make([]uuid.UUID, 0, len(ids))

	for _, s := range ids {
		id, err := uuid.Parse(s)
		if err != nil || id == uuid.Nil || id == sender {
			continue
		}

		if _, dup := seen[id]; dup {
			continue
		}

		seen[id] = struct{}{}
		out = append(out, id)
	}

	return out
}
