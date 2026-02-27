package payload

import (
	"encoding/json"
	"strings"

	"github.com/webitel/im-delivery-service/internal/domain/model"
	"github.com/webitel/im-delivery-service/internal/domain/util"
)

// Kind represents the classification of the communication thread.
type Kind int32

const (
	Direct Kind = iota + 1
	Group
	Channel
)

// UnmarshalJSON handles conversion from string labels (e.g., "ThreadDirect")
// to internal integer-based enum values.
func (k *Kind) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	// Match the incoming string to the correct enum constant.
	switch {
	case strings.Contains(s, "Direct"):
		*k = Direct
	case strings.Contains(s, "Group"):
		*k = Group
	case strings.Contains(s, "Channel"):
		*k = Channel
	default:
		*k = 0 // Unknown
	}
	return nil
}

// String returns the normalized lowercase representation of the thread kind.
func (k Kind) String() string {
	switch k {
	case Direct:
		return "direct"
	case Group:
		return "group"
	case Channel:
		return "channel"
	default:
		return "unknown"
	}
}

// Recipient defines the primary participant receiving the thread notification.
type Recipient struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ThreadCreatedV1 represents the AMQP message payload for a new thread.
type ThreadCreatedV1 struct {
	ThreadID  string    `json:"id"`
	DomainID  int32     `json:"domain_id"`
	CreatedAt string    `json:"created_at"`
	Subject   string    `json:"subject"`
	Recipient Recipient `json:"recipient"`
	Kind      Kind      `json:"kind"`
	Members   []string  `json:"members"`
}

// ToDomain maps the DTO payload to the internal business domain model.
func (t *ThreadCreatedV1) ToDomain() *model.Thread {
	return &model.Thread{
		ID:        util.SafeParseUUID(t.ThreadID),
		DomainID:  t.DomainID,
		CreatedAt: util.SafeParseRFC3339(t.CreatedAt),
		Subject:   t.Subject,
		Recipient: model.Recipient{
			ID:   util.SafeParseUUID(t.Recipient.ID),
			Name: t.Recipient.Name,
		},
		Type: t.Kind.String(),
	}
}
