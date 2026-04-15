package payload

import "github.com/webitel/im-delivery-service/internal/domain/model"

type VariablesV1 struct {
	Members   []string                       `json:"members"`
	Variables map[string]model.VariableEntry `json:"variables"`
	DomainID  int32                          `json:"domain_id"`
}

func (p *VariablesV1) ToDomain(action string) *model.VariablesPayload {
	return &model.VariablesPayload{
		DomainID:  int64(p.DomainID),
		Variables: p.Variables,
		Action:    action,
	}
}
