package lp

import (
	"github.com/go-chi/chi/v5"
	lpmarshaller "github.com/webitel/im-delivery-service/internal/handler/marshaller/lp"
	"go.uber.org/fx"
)

// Module defines the lifecycle and dependencies for the Long Polling transport layer.
var Module = fx.Module("lp-module",
	fx.Provide(
		// Provide the specific LP implementation of the marshaller.
		lpmarshaller.New,
		// Provide the LP handler which depends on the lpmarshaller.
		NewLPHandler,
	),

	// Automatically register routes in the HTTP router during the application bootstrap.
	fx.Invoke(RegisterLPRoutes),
)

// RegisterLPRoutes binds the Poll method to the corresponding HTTP path.
// It uses chi.Router as the primary multiplexer.
func RegisterLPRoutes(r chi.Router, h *LPHandler) {
	// [ROUTE] The {userID} parameter is extracted inside h.Poll using chi.URLParam.
	r.Get("/poll/{userID}", h.Poll)
}
