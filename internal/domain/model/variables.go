package model

import (
	"fmt"
)

const (
	VariableActionSet   = "set"
	VariableActionFlush = "flush"
)

type VariableEntry struct {
	Value any    `json:"value"`
	SetBy string `json:"set_by"`
	SetAt string `json:"set_at"`
}

type VariablesMap map[string]VariableEntry

type VariablesPayload struct {
	// Fields for internal routing (Routable interface)
	// Added omitempty or removed from JSON to keep socket payload clean
	DomainID  int64  `json:"-"`
	ContactID string `json:"-"`

	// Data for the client
	Variables VariablesMap `json:"variables"`
	Action    string       `json:"action"`
}

// RoutingKey satisfies the Routable interface.
func (v *VariablesPayload) RoutingKey() string {
	if v.Action == "" {
		v.Action = VariableActionSet
	}
	// Even if not sent to socket, we can use them for AMQP routing internally
	return fmt.Sprintf("im_delivery.v1.%d.contact.%s.variables.%s",
		v.DomainID,
		v.ContactID,
		v.Action,
	)
}
