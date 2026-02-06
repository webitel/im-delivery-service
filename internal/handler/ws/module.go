package ws

import (
	"log/slog"
	"net/http"

	wsmarshaller "github.com/webitel/im-delivery-service/internal/handler/marshaller/ws"
	"go.uber.org/fx"
)

var Module = fx.Module("delivery-ws",
	fx.Provide(
		wsmarshaller.New,
		// [REGISTRATION]
		fx.Annotate(
			NewWSHandler,
			fx.As(new(http.Handler)),
			fx.ResultTags(`name:"ws_handler"`),
		),
	),

	// [DECORATION]
	fx.Decorate(
		fx.Annotate(
			func(handler http.Handler, log *slog.Logger) http.Handler {
				return ContextBridge(handler, log)
			},
			fx.ParamTags(`name:"ws_handler"`),
			fx.ResultTags(`name:"ws_handler"`),
		),
	),

	// [EXECUTION]
	fx.Invoke(
		fx.Annotate(
			RegisterRoutes,
			fx.ParamTags("", `name:"ws_handler"`),
		),
	),
)

func RegisterRoutes(mux *http.ServeMux, handler http.Handler) {
	// [BINDING]
	mux.Handle("/ws", handler)
}
