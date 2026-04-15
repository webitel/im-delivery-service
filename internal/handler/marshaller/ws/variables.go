package wsmarshaller

import "github.com/webitel/im-delivery-service/internal/domain/model"

// WSVariables represents the cleaned variables payload for WebSocket clients.
// We removed DomainID, ThreadID, and ContactID as they were redundant/empty.
type WSVariables struct {
	Variables model.VariablesMap `json:"variables,omitempty"`
	Action    string             `json:"action"`
}

// mapVariables converts domain model.VariablesPayload to a minimal WSVariables DTO.
func mapVariables(p *model.VariablesPayload) *WSVariables {
	if p == nil {
		return nil
	}

	// Ensure we have a valid action string from constants
	action := p.Action
	if action == "" {
		action = model.VariableActionSet
	}

	return &WSVariables{
		Variables: p.Variables,
		Action:    action,
	}
}
