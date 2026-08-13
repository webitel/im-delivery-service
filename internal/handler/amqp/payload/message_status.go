package payload

import (
	"github.com/google/uuid"

	"github.com/webitel/im-delivery-service/internal/domain/model"
	"github.com/webitel/im-delivery-service/internal/domain/util"
)

// MessageStatusV1 is the im_message.<thread_id>.message.status.v1 event
// published by im-thread-service when per-recipient delivery statuses
// actually change (delivered/read/failed).
type MessageStatusV1 struct {
	ThreadID   string         `json:"thread_id"`
	DomainID   int32          `json:"domain_id"`
	MemberID   string         `json:"member_id"`
	MessageIDs []string       `json:"message_ids"`
	Status     string         `json:"status"`
	Via        string         `json:"via,omitempty"`
	Error      map[string]any `json:"error,omitempty"`
	OccurredAt string         `json:"occurred_at"`
	// Participants are contact ids of all current thread members, stamped by
	// im-thread-service so the event can be fanned out without a thread lookup.
	Participants []string `json:"participants,omitempty"`
	// UpToMessageID is the highest message id covered by this status change
	// (delivered/read watermark). When present, clients can advance their
	// local peer-horizon in O(1) instead of iterating message_ids.
	UpToMessageID string `json:"up_to_message_id,omitempty"`
	// UpToSeq is the per-thread sequence number of the delivered/read-up-to boundary
	// (preferred watermark; supercedes UpToMessageID).
	UpToSeq int64 `json:"up_to_seq,omitempty"`
}

// ToDomain converts the AMQP payload into the internal domain model.
func (d *MessageStatusV1) ToDomain() *model.MessageStatusUpdate {
	messageIDs := make([]uuid.UUID, 0, len(d.MessageIDs))
	for _, raw := range d.MessageIDs {
		if id := util.SafeParseUUID(raw); id != uuid.Nil {
			messageIDs = append(messageIDs, id)
		}
	}

	return &model.MessageStatusUpdate{
		ThreadID:      util.SafeParseUUID(d.ThreadID),
		MemberID:      util.SafeParseUUID(d.MemberID),
		MessageIDs:    messageIDs,
		Status:        d.Status,
		Via:           d.Via,
		Error:         d.Error,
		OccurredAt:    util.SafeParseRFC3339(d.OccurredAt),
		UpToMessageID: util.SafeParseUUID(d.UpToMessageID),
		UpToSeq:       d.UpToSeq,
	}
}

// ParticipantIDs parses the participant contact ids, falling back to the
// affected member when the event carries no participant list.
func (d *MessageStatusV1) ParticipantIDs() []uuid.UUID {
	res := make([]uuid.UUID, 0, len(d.Participants))
	seen := make(map[uuid.UUID]struct{}, len(d.Participants))

	for _, raw := range d.Participants {
		id := util.SafeParseUUID(raw)
		if id == uuid.Nil {
			continue
		}

		if _, ok := seen[id]; ok {
			continue
		}

		seen[id] = struct{}{}
		res = append(res, id)
	}

	if len(res) == 0 {
		if id := util.SafeParseUUID(d.MemberID); id != uuid.Nil {
			res = append(res, id)
		}
	}

	return res
}
