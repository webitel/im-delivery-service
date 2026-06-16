package amqp

import (
	"context"

	"github.com/google/uuid"
	"github.com/webitel/im-delivery-service/internal/domain/event"
	"github.com/webitel/im-delivery-service/internal/domain/model"
	"github.com/webitel/im-delivery-service/internal/handler/amqp/payload"
	"github.com/webitel/webitel-go-kit/pkg/semconv"
)

// [ON_VARIABLES_SET] Entry point for variable updates.
func (h *MessageHandler) OnVariablesSetV1(ctx context.Context, raw *payload.VariablesV1) ([]event.Eventer, error) {
	return h.handleVariables(raw, model.VariableActionSet, event.VariableSet)
}

// [ON_VARIABLES_FLUSH] Entry point for variable cleanup.
func (h *MessageHandler) OnVariablesFlushV1(ctx context.Context, raw *payload.VariablesV1) ([]event.Eventer, error) {
	return h.handleVariables(raw, model.VariableActionFlush, event.VariableFlush)
}

// handleVariables maps the event to multiple members if they are connected.
func (h *MessageHandler) handleVariables(raw *payload.VariablesV1, action string, kind event.EventKind) ([]event.Eventer, error) {
	if len(raw.Members) == 0 {
		return nil, nil
	}

	domainModel := raw.ToDomain(action)

	events := make([]event.Eventer, 0, len(raw.Members))
	for _, m := range raw.Members {
		memberID, err := uuid.Parse(m)
		if err != nil {
			h.logger.Warn("failed_to_parse_member_id", "id", m, semconv.ErrorKey, err)
			continue
		}

		// Optimization: Check if target is connected locally or we are the leader node.
		if h.leader.IsLeader() || h.hub.Connected(memberID) {
			events = append(events, event.NewVariableEvent(
				domainModel,
				memberID,
				domainModel.DomainID,
				kind,
			))
		}
	}

	return events, nil
}
