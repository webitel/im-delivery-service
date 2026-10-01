package httpsrv

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"

	"go.uber.org/fx"

	"github.com/webitel/im-delivery-service/config"
)

var Module = fx.Module("http-server",
	fx.Provide(http.NewServeMux), // [ROUTER] Provides central *http.ServeMux
	fx.Provide(New),              // [LISTENER] Binds the HTTP/WS port
	fx.Invoke(func(*Server) {}),  // [LIFECYCLE] Starts the listener
)

type Server struct {
	*http.Server

	listener net.Listener
}

func New(lc fx.Lifecycle, mux *http.ServeMux, log *slog.Logger, cfg *config.Config) (*Server, error) {
	srv := &http.Server{
		Addr:    cfg.Service.HTTPAddr,
		Handler: mux,
	}

	l, err := net.Listen("tcp", srv.Addr)
	if err != nil {
		return nil, fmt.Errorf("listen http %s: %w", srv.Addr, err)
	}

	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			log.Info("HTTP_SERVER_STARTED", slog.String("addr", srv.Addr))
			// [IO] Run in background to not block app startup
			go func() {
				if err := srv.Serve(l); err != nil && err != http.ErrServerClosed {
					log.Error("HTTP_SERVER_CRASHED", slog.Any("err", err))
				}
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Info("HTTP_SERVER_STOPPING")

			return srv.Shutdown(ctx)
		},
	})

	return &Server{Server: srv, listener: l}, nil
}

func (s *Server) Listener() net.Listener {
	return s.listener
}
