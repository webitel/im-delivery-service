package cmd

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/redis/go-redis/v9"

	"github.com/webitel/webitel-go-kit/infra/health"
	healthhttp "github.com/webitel/webitel-go-kit/infra/health/http"

	grpcsrv "github.com/webitel/im-delivery-service/infra/server/grpc"
	httpsrv "github.com/webitel/im-delivery-service/infra/server/http"
)

func registerHealth(
	log *slog.Logger,
	h *health.Registry,
	grpcServer *grpcsrv.Server,
	httpServer *httpsrv.Server,
	mux *http.ServeMux,
	rdb *redis.Client,
) {
	mux.Handle("/livez", healthhttp.LivenessHandler(h, healthhttp.WithLogger(log)))
	mux.Handle("/readyz", healthhttp.ReadinessHandler(h, healthhttp.WithLogger(log)))
	mux.Handle("/healthz", healthhttp.HealthHandler(h, healthhttp.WithLogger(log)))

	h.Critical("grpc", health.ListenerCheck(grpcServer.Listener()))
	h.Critical("http", health.ListenerCheck(httpServer.Listener()))
	h.Informational("redis", func(ctx context.Context) error { return rdb.Ping(ctx).Err() })
}
