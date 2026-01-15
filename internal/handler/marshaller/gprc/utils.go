package grpcmarshaller

import (
	impb "github.com/webitel/im-delivery-service/gen/go/api/v1"
	"github.com/webitel/im-delivery-service/internal/domain/model"
)

func mapPriority(p model.InboundEventPriority) impb.EventPriority {
	switch p {
	case model.PriorityLow:
		return impb.EventPriority_LOW
	case model.PriorityNormal:
		return impb.EventPriority_NORMAL
	case model.PriorityHigh:
		return impb.EventPriority_HIGH
	default:
		return impb.EventPriority_PRIORITY_UNSPECIFIED
	}
}
