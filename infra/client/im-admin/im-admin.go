package imadmin

import (
	"context"
	"fmt"
	"log/slog"

	"google.golang.org/grpc"

	"github.com/webitel/webitel-go-kit/infra/discovery"
	rpc "github.com/webitel/webitel-go-kit/infra/transport/gRPC"

	adminv1 "github.com/webitel/im-delivery-service/gen/go/admin/v1"
	webitel "github.com/webitel/im-delivery-service/infra/client"
	infratls "github.com/webitel/im-delivery-service/infra/tls"
)

const ServiceName string = "im-account-service"

type Client struct {
	logger *slog.Logger
	rpc    *rpc.Client[adminv1.ApplicationsClient]
	tls    *infratls.Config
}

// SearchApps implements [admin.ApplicationsClient].
func (c *Client) SearchApps(ctx context.Context, in *adminv1.SearchAppRequest, opts ...grpc.CallOption) (*adminv1.ApplicationList, error) {
	var resp *adminv1.ApplicationList

	err := c.rpc.Execute(ctx, func(api adminv1.ApplicationsClient) error {
		var err error

		resp, err = api.SearchApps(ctx, in, opts...)

		return err
	})

	return resp, err
}

// New initializes a resilient gRPC client for the Auth service.
func New(logger *slog.Logger, discovery discovery.DiscoveryProvider, tls *infratls.Config) (*Client, error) {
	factory := func(conn *grpc.ClientConn) adminv1.ApplicationsClient {
		return adminv1.NewApplicationsClient(conn)
	}

	c, err := webitel.New(logger, discovery, ServiceName, tls, factory, true)
	if err != nil {
		return nil, fmt.Errorf("[im-admin-client] initialization failed: %w", err)
	}

	return &Client{
		logger: logger,
		rpc:    c,
	}, nil
}

func (c *Client) Close() error {
	if c.rpc != nil {
		return c.rpc.Close()
	}

	return nil
}
