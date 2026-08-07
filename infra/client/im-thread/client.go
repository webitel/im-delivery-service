// Package imthread wraps the im-thread-service MessageStatus gRPC API used
// to report per-recipient delivery confirmations (stream ACKs, pushes).
package imthread

import (
	"context"
	"fmt"
	"log/slog"

	"google.golang.org/grpc"

	"github.com/webitel/webitel-go-kit/infra/discovery"
	rpc "github.com/webitel/webitel-go-kit/infra/transport/gRPC"

	threadv1 "github.com/webitel/im-delivery-service/gen/go/thread/v1"
	webitel "github.com/webitel/im-delivery-service/infra/client"
	infratls "github.com/webitel/im-delivery-service/infra/tls"
)

const ServiceName string = "im-thread-service"

type Client struct {
	logger *slog.Logger
	// [GENERIC_RPC] Holds the go-kit RPC client for the thread status service
	rpc *rpc.Client[threadv1.MessageStatusClient]
}

func New(logger *slog.Logger, discovery discovery.DiscoveryProvider, tls *infratls.Config) (*Client, error) {
	// [FACTORY] Required by go-kit to instantiate the gRPC stub
	factory := func(conn *grpc.ClientConn) threadv1.MessageStatusClient {
		return threadv1.NewMessageStatusClient(conn)
	}

	// [INIT] Initialize the shared RPC client wrapper
	c, err := webitel.New(logger, discovery, ServiceName, tls, factory, true)
	if err != nil {
		return nil, fmt.Errorf("[im-thread-client] initialization failed: %w", err)
	}

	return &Client{
		logger: logger,
		rpc:    c,
	}, nil
}

// MarkDelivered reports batched per-recipient delivery confirmations.
func (c *Client) MarkDelivered(ctx context.Context, req *threadv1.MarkDeliveredRequest) (*threadv1.MarkStatusResponse, error) {
	var resp *threadv1.MarkStatusResponse

	err := c.rpc.Execute(ctx, func(api threadv1.MessageStatusClient) error {
		c.logger.Debug("THREAD.MARK_DELIVERED", slog.Int("receipts", len(req.GetReceipts())))

		var err error
		resp, err = api.MarkDelivered(ctx, req)

		return err
	})

	return resp, err
}

// MarkRead reports batched per-recipient read confirmations (read-up-to).
func (c *Client) MarkRead(ctx context.Context, req *threadv1.MarkReadRequest) (*threadv1.MarkStatusResponse, error) {
	var resp *threadv1.MarkStatusResponse

	err := c.rpc.Execute(ctx, func(api threadv1.MessageStatusClient) error {
		c.logger.Debug("THREAD.MARK_READ", slog.Int("receipts", len(req.GetReceipts())))

		var err error
		resp, err = api.MarkRead(ctx, req)

		return err
	})

	return resp, err
}

// Close gracefully shuts down the underlying gRPC connection pool.
func (c *Client) Close() error {
	if c.rpc != nil {
		return c.rpc.Close()
	}

	return nil
}
