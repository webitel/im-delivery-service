package event

type EventPriority int32

const (
	PriorityLow    EventPriority = 10
	PriorityNormal EventPriority = 20
	PriorityHigh   EventPriority = 30
)
