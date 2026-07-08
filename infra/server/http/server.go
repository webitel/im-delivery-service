package httpsrv

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.uber.org/fx"

	"github.com/webitel/webitel-go-kit/pkg/depenlog"
	"github.com/webitel/webitel-go-kit/pkg/logger"
	"github.com/webitel/webitel-go-kit/pkg/semconv"

	"go.uber.org/fx"

	"github.com/webitel/im-delivery-service/config"
)

var Module = fx.Module("http-server",
	fx.Provide(http.NewServeMux), // [ROUTER] Provides central *http.ServeMux
	fx.Invoke(Start),             // [LIFECYCLE] Starts the listener
)

func Start(lc fx.Lifecycle, mux *http.ServeMux, log *slog.Logger, kit logger.Logger, cfg *config.Config) {
	instrumented := otelhttp.NewHandler(depenlog.Middleware(kit)(mux), "http")
	root := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isWebSocketUpgrade(r) {
			mux.ServeHTTP(w, r)
			return
		}
		instrumented.ServeHTTP(w, r)
	})

	srv := &http.Server{
		Addr:     cfg.Service.HTTPAddr,
		Handler:  root,
		ErrorLog: depenlog.ErrorLog(kit),
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Info("HTTP_SERVER_STARTED", slog.String("addr", srv.Addr))
			// [IO] Run in background to not block app startup
			go func() {
				if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					log.Error("HTTP_SERVER_CRASHED", slog.Any(semconv.ErrorKey, err))
				}
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Info("HTTP_SERVER_STOPPING")

			return srv.Shutdown(ctx)
		},
	})
}

func isWebSocketUpgrade(r *http.Request) bool {
	return strings.EqualFold(r.Header.Get("Upgrade"), "websocket")
}
