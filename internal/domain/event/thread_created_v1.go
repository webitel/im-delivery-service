package event

import (
	"github.com/google/uuid"
	"github.com/webitel/im-delivery-service/internal/domain/model"
)

var _ Eventer = (*ThreadCreatedV1Event)(nil)

type ThreadCreatedV1Event struct {
	ID     uuid.UUID
	Thread *model.Thread `json:"thread"`
	Cached any           `json:"-"`
}

func NewThreadCreatedV1Event(thread *model.Thread) *ThreadCreatedV1Event {
	return &ThreadCreatedV1Event{
		ID:     uuid.New(),
		Thread: thread,
	}
}

func (e *ThreadCreatedV1Event) GetID() string              { return e.ID.String() }
func (e *ThreadCreatedV1Event) GetPayload() any            { return e.Thread }
func (e *ThreadCreatedV1Event) GetUserID() uuid.UUID       { return e.Thread.Recipient.ID }
func (e *ThreadCreatedV1Event) GetOccurredAt() int64       { return e.Thread.CreatedAt }
func (e *ThreadCreatedV1Event) GetKind() EventKind         { return ThreadCreated }
func (e *ThreadCreatedV1Event) GetPriority() EventPriority { return PriorityHigh }
func (e *ThreadCreatedV1Event) GetCached() any             { return e.Cached }
func (e *ThreadCreatedV1Event) SetCached(v any)            { e.Cached = v }
