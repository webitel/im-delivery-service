package payload

import (
	"encoding/json"
	"strings"

	"github.com/webitel/im-delivery-service/internal/domain/model"
	"github.com/webitel/im-delivery-service/internal/domain/util"
)

type Kind int32

const (
	Direct Kind = iota + 1
	Group
	Channel
)

func (k *Kind) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	switch {
	case strings.Contains(s, "Direct"):
		*k = Direct
	case strings.Contains(s, "Group"):
		*k = Group
	case strings.Contains(s, "Channel"):
		*k = Channel
	default:
		*k = 0
	}

	return nil
}

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

type ThreadRecipient struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// 👇 members нової структури
type ThreadMember struct {
	MemberID  string `json:"member_id"`
	ContactID string `json:"contact_id"`
	Role      int    `json:"role"`
}

type ThreadCreatedV1 struct {
	ThreadID  string          `json:"id"`
	DomainID  int32           `json:"domain_id"`
	CreatedAt string          `json:"created_at"`
	Subject   string          `json:"subject"`
	Recipient ThreadRecipient `json:"recipient"`
	Kind      Kind            `json:"kind"`
	Members   []ThreadMember  `json:"members"`
}

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
