package event

import "github.com/google/uuid"

// [EVENTER] Base contract for all events flowing through the hub.
type Eventer interface {
	GetID() string
	GetKind() EventKind
	GetKindName() string
	GetUserID() uuid.UUID
	GetPriority() EventPriority
	GetOccurredAt() int64
	GetPayload() any
	IsEcho() bool
	GetMetadata() map[string]string
	GetCached() any
	SetCached(any)
}

// [ROUTABLE] Enables message bus (AMQP) routing via specific keys.
type Routable interface {
	RoutingKey() string
}

// [TRACKABLE] Enables delivery guarantees and push notification logic.
type IsPushable interface {
	IsPushable() bool
}

// [PUSH_PROVIDER] Provides metadata for external push gateways.
type Notifier interface {
	NotificationTitle() string
	NotificationBody() string
}
