package amqp

import (
	"github.com/google/uuid"
)

// [FILTER] Decides which IDs should be processed by this node based on Hub connections.
func (h *MessageHandler) filter(senderIDStr string, all []uuid.UUID) (uuid.UUID, []uuid.UUID) {
	sID, _ := uuid.Parse(senderIDStr)
	if h.leader.IsLeader() {
		return sID, all
	}

	res := make([]uuid.UUID, 0, len(all))
	for _, id := range all {
		if id == sID || h.hub.Connected(id) {
			res = append(res, id)
		}
	}

	return sID, res
}

// [TO_UUIDS] Helper for converting a slice of strings to UUIDs.
func toUUIDs(src []string) []uuid.UUID {
	res := make([]uuid.UUID, 0, len(src))
	for _, s := range src {
		if id, err := uuid.Parse(s); err == nil {
			res = append(res, id)
		}
	}

	return res
}
