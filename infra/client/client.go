package client

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/webitel/im-delivery-service/infra/client/interceptors"
	infratls "github.com/webitel/im-delivery-service/infra/tls"
	ds "github.com/webitel/webitel-go-kit/infra/discovery"
	rpc "github.com/webitel/webitel-go-kit/infra/transport/gRPC"
	"github.com/webitel/webitel-go-kit/infra/transport/gRPC/resolver/discovery"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
)

// New initializes a go-kit RPC client with Discovery and OPTIONAL Circuit Breaker.
// [CHANGE] Added 'withBreaker' boolean parameter.
func New[T any](
	log *slog.Logger,
	dp ds.DiscoveryProvider,
	target string,
	tlsCong *infratls.Config,
	factory rpc.ClientFactory[T],
	withBreaker bool,
) (*rpc.Client[T], error) {
	options := []grpc.DialOption{
		grpc.WithTransportCredentials(credentials.NewTLS(tlsCong.Client)),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
		grpc.WithResolvers(discovery.NewBuilder(dp, discovery.WithInsecure(true))),
	}

	// [STABILITY] Only wrap calls with Circuit Breaker if explicitly requested.
	if withBreaker {
		cb := interceptors.NewBreakerInterceptor()
		options = append(options, grpc.WithChainUnaryInterceptor(
			cb.UnaryClientInterceptor(),
		))
	}

	client, err := rpc.NewClient(
		context.Background(),
		factory,
		rpc.WithTarget(fmt.Sprintf("discovery:///%s", target)),
		rpc.WithDialOptions(options...),
		// [RETRY] Built-in transport-level retries (works regardless of breaker)
		rpc.WithRetry(rpc.DefaultRetryConfig()),
		rpc.WithKeepalive(
			keepalive.ClientParameters{
				Time:                10 * time.Minute,
				Timeout:             20 * time.Second,
				PermitWithoutStream: false,
			},
		),
	)
	if err != nil {
		return nil, err
	}

	return client, nil
}
