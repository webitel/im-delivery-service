package imadmin

import (
	"context"
	"fmt"
	"log/slog"

	adminv1 "github.com/webitel/im-delivery-service/gen/go/admin/v1"
	webitel "github.com/webitel/im-delivery-service/infra/client"
	infratls "github.com/webitel/im-delivery-service/infra/tls"
	"github.com/webitel/webitel-go-kit/infra/discovery"
	rpc "github.com/webitel/webitel-go-kit/infra/transport/gRPC"
	"google.golang.org/grpc"
)

const ServiceName string = "im-account-service"

// [INTERFACE_GUARD] Now correctly matches the CLIENT interface.
var _ adminv1.ApplicationsClient = (*Client)(nil)

type Client struct {
	logger *slog.Logger
	rpc    *rpc.Client[adminv1.ApplicationsClient]
	tls    *infratls.Config
}

// CreateApp implements [admin.ApplicationsClient].
func (c *Client) CreateApp(ctx context.Context, in *adminv1.CreateAppRequest, opts ...grpc.CallOption) (*adminv1.Application, error) {
	panic("unimplemented")
}

// DeleteApps implements [admin.ApplicationsClient].
func (c *Client) DeleteApps(ctx context.Context, in *adminv1.DeleteAppRequest, opts ...grpc.CallOption) (*adminv1.ApplicationList, error) {
	panic("unimplemented")
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

// UpdateApp implements [admin.ApplicationsClient].
func (c *Client) UpdateApp(ctx context.Context, in *adminv1.UpdateAppRequest, opts ...grpc.CallOption) (*adminv1.Application, error) {
	panic("unimplemented")
}

// New initializes a resilient gRPC client for the Auth service.
func New(logger *slog.Logger, discovery discovery.DiscoveryProvider, tls *infratls.Config) (*Client, error) {
	factory := func(conn *grpc.ClientConn) adminv1.ApplicationsClient {
		return adminv1.NewApplicationsClient(conn)
	}

	c, err := webitel.New(logger, discovery, ServiceName, tls, factory)
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
