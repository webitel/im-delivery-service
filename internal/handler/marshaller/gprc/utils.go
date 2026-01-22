package grpcmarshaller

import (
	impb "github.com/webitel/im-delivery-service/gen/go/delivery/v1"
	"github.com/webitel/im-delivery-service/internal/domain/event"
)

func mapPriority(p event.EventPriority) impb.EventPriority {
	switch p {
	case event.PriorityLow:
		return impb.EventPriority_LOW
	case event.PriorityNormal:
		return impb.EventPriority_NORMAL
	case event.PriorityHigh:
		return impb.EventPriority_HIGH
	default:
		return impb.EventPriority_PRIORITY_UNSPECIFIED
	}
}
