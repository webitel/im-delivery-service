// [internal/domain/model/read.go]

package model

import "github.com/google/uuid"

// [READ_PAYLOAD] Slim payload for notification dismissal.
type MessageReadPayload struct {
	MessageID uuid.UUID `json:"message_id"`
}
