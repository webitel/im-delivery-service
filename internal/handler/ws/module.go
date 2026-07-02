// internal/handler/ws/module.go
package ws

import (
	"net/http"

	"go.uber.org/fx"

	"github.com/webitel/im-delivery-service/internal/handler/marshaller" // Import the interface package
	wsmarshaller "github.com/webitel/im-delivery-service/internal/handler/marshaller/ws"
)

var Module = fx.Module("delivery-ws",
	fx.Provide(
		// Use fx.Annotate to cast the concrete implementation to the interface
		fx.Annotate(
			wsmarshaller.New,
			fx.As(new(marshaller.EventMarshaller)),
		),
		NewWSHandler,
	),
	fx.Invoke(
		func(mux *http.ServeMux, h *WSHandler) {
			mux.Handle("/im/ws", h.AuthenticationMiddleware(h))
		},
	),
)
