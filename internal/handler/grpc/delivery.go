package grpc

import (
	"log/slog"
	"time"

	"github.com/google/uuid"
	impb "github.com/webitel/im-delivery-service/gen/go/delivery/v1"
	"github.com/webitel/im-delivery-service/internal/domain/event"
	"github.com/webitel/im-delivery-service/internal/domain/model"
	grpcmarshaller "github.com/webitel/im-delivery-service/internal/handler/marshaller/gprc"
	"github.com/webitel/im-delivery-service/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	serverVersion = "v1"
)

type DeliveryService struct {
	logger    *slog.Logger
	deliverer service.Deliverer
	impb.UnimplementedDeliveryServer
}

func NewDeliveryService(logger *slog.Logger, deliverer service.Deliverer) *DeliveryService {
	return &DeliveryService{
		logger:    logger,
		deliverer: deliverer,
	}
}

// Stream manages the lifecycle of a long-lived HTTP/2 bidirectional/server-streaming session.
func (d *DeliveryService) Stream(req *impb.StreamRequest, stream impb.Delivery_StreamServer) error {
	// [HTTP2_LIFECYCLE]
	// The context is tied to the underlying HTTP/2 stream. If the TCP connection drops
	// or the client sends a RST_STREAM frame, ctx.Done() will trigger.
	ctx := stream.Context()
	startTime := time.Now()

	// [IDENTITY_RESOLVER]
	// In production, this should be extracted from context metadata (JWT/OAuth2).
	userID := uuid.MustParse("019bb6d7-8bb8-7a5c-b163-8cf8d362a474")

	l := d.logger.With(
		slog.String("user_id", userID.String()),
		slog.String("version", serverVersion),
	)

	// [ACTOR_ATTACHMENT]
	// Subscribe links this specific gRPC stream to the User's Virtual Cell (Actor).
	// This ensures all events routed to the Hub for this UserID will reach this stream.
	conn, err := d.deliverer.Subscribe(ctx, userID)
	if err != nil {
		l.Error("SUBSCRIPTION_REJECTED", slog.Any("err", err))
		return status.Error(codes.Internal, "failed to establish connection session")
	}

	connID := conn.GetID()
	l = l.With(slog.String("conn_id", connID.String()))

	// [RESOURCE_RECLAMATION]
	// Ensure the connector is detached from the Hub when the function returns.
	// This prevents memory leaks and ensures the Hub doesn't try to send to a dead stream.
	defer func() {
		d.deliverer.Unsubscribe(userID, connID)
		l.Info("STREAM_TERMINATED", slog.Duration("duration", time.Since(startTime)))
	}()

	l.Info("STREAM_ESTABLISHED")

	// [HANDSHAKE_LOGIC]
	// Create the payload from model package.
	welcomeEv := event.NewSystemEvent(userID, event.Connected, event.PriorityNormal, &model.ConnectedPayload{
		Ok:            true,
		ConnectionID:  connID.String(),
		ServerVersion: serverVersion,
	})
	if err := stream.Send(grpcmarshaller.MarshallDeliveryEvent(welcomeEv)); err != nil {
		return err
	}

	// [EVENT_LOOP]
	// Main delivery loop that bridges the internal Actor mailbox with the gRPC stream.
	for {
		select {
		case <-ctx.Done():
			// [GHOST_CLEANUP]
			// Triggers on client disconnect, timeout, or KeepAlive failure.
			l.Debug("CLIENT_DISCONNECTED_SIGTERM")
			return nil

		case ev, ok := <-conn.Recv():
			if !ok {
				// [BACKPRESSURE_SIGNAL]
				// Internal channel closed by the Hub (e.g., during forceful eviction).
				l.Warn("HUB_FORCED_DISCONNECT")
				return status.Error(codes.Unavailable, "session_terminated_by_server")
			}

			// [TRANSMIT_OVER_HTTP2]
			// Serialize and push the event into the gRPC transmit buffer.
			// gRPC handles internal flow control and HTTP/2 framing.
			pbEv := grpcmarshaller.MarshallDeliveryEvent(ev)
			if err := stream.Send(pbEv); err != nil {
				l.Error("TRANSMISSION_ERROR", slog.Any("err", err))
				// Returning error here triggers a gRPC status code (DataLoss) to the client.
				return status.Error(codes.DataLoss, "stream_transmission_failed")
			}
		}
	}
}
