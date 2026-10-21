package payload

import (
	"github.com/webitel/im-delivery-service/internal/domain/model"
	"github.com/webitel/im-delivery-service/internal/domain/util"
)

// ThreadTagsChangedV1 is the im_thread.<thread_id>.thread.tags.v1 event published by
// im-thread-service when a contact's own tags on a thread change.
type ThreadTagsChangedV1 struct {
	ThreadID      string   `json:"thread_id"`
	DomainID      int32    `json:"domain_id"`
	ContactID     string   `json:"contact_id"`
	Tags          []string `json:"tags"`
	UpdatesCursor string   `json:"updates_cursor"`
}

func (t *ThreadTagsChangedV1) ToDomain() *model.ThreadTagsChanged {
	tags := t.Tags
	if tags == nil {
		tags = make([]string, 0)
	}

	return &model.ThreadTagsChanged{
		ThreadID:      util.SafeParseUUID(t.ThreadID),
		Tags:          tags,
		UpdatesCursor: t.UpdatesCursor,
	}
}
