package grpc

import (
	"log/slog"

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

var _ impb.DeliveryServer = (*DeliveryService)(nil)

type DeliveryService struct {
	logger    *slog.Logger
	deliverer service.Deliverer
	auther    service.Auther
	impb.UnimplementedDeliveryServer
}

func NewDeliveryService(logger *slog.Logger, deliverer service.Deliverer, auther service.Auther) *DeliveryService {
	return &DeliveryService{
		logger:    logger,
		deliverer: deliverer,
		auther:    auther,
	}
}

// Stream manages the lifecycle of a long-lived HTTP/2 bidirectional/server-streaming session.
func (d *DeliveryService) Stream(req *impb.StreamRequest, stream impb.Delivery_StreamServer) error {
	ctx := stream.Context()

	// [AUTH_CHECK]
	auth, err := d.auther.Inspect(ctx)
	if err != nil {
		d.logger.Warn("[AUTH] access denied", slog.Any("err", err))
		return err
	}

	userID, err := uuid.Parse(auth.ContactID)
	if err != nil {
		d.logger.Error("[AUTH] failed to parse contact identity",
			slog.String("contact_id", auth.ContactID),
			slog.Any("err", err),
		)
		return status.Error(codes.InvalidArgument, "invalid user id format")
	}

	// Create a stream-scoped logger to track this specific connection
	l := d.logger.With(
		slog.String("user_id", userID.String()),
		slog.String("session_id", uuid.NewString()),
	)

	l.Info("[STREAM] incoming connection request", slog.String("version", serverVersion))

	// [ACTOR_ATTACHMENT]
	// Subscribe links this specific gRPC stream to the User's Virtual Cell (Actor).
	// This ensures all events routed to the Hub for this UserID will reach this stream.
	conn, err := d.deliverer.Subscribe(ctx, userID)
	if err != nil {
		l.Error("[HUB] subscription rejected", slog.Any("err", err))
		return status.Error(codes.Internal, "failed to establish connection session")
	}

	// [RESOURCE_RECLAMATION]
	// Ensure the connector is detached from the Hub when the function returns.
	// This prevents memory leaks and ensures the Hub doesn't try to send to a dead stream.
	defer func() {
		d.deliverer.Unsubscribe(userID, conn.GetID())
		l.Info("[STREAM] connection closed and resources reclaimed",
			slog.String("conn_id", conn.GetID().String()),
		)
	}()

	l.Info("[STREAM] session established", slog.String("conn_id", conn.GetID().String()))

	// [HANDSHAKE_LOGIC]
	// Create the payload from model package.
	welcomeEv := event.NewSystemEvent(userID, event.Connected, event.PriorityNormal, &model.ConnectedPayload{
		Ok:            true,
		ConnectionID:  conn.GetID().String(),
		ServerVersion: serverVersion,
	})

	if err := stream.Send(grpcmarshaller.MarshallDeliveryEvent(welcomeEv)); err != nil {
		l.Error("[STREAM] handshake delivery failed", slog.Any("err", err))
		return err
	}

	// [EVENT_LOOP]
	// Main delivery loop that bridges the internal Actor mailbox with the gRPC stream.
	for {
		select {
		case <-ctx.Done():
			// [GHOST_CLEANUP]
			// Triggers on client disconnect, timeout, or KeepAlive failure.
			l.Info("[STREAM] client terminated connection", slog.Any("reason", ctx.Err()))
			return nil

		case ev, ok := <-conn.Recv():
			if !ok {
				// [BACKPRESSURE_SIGNAL]
				// Internal channel closed by the Hub (e.g., during forceful eviction).
				l.Warn("[HUB] mailbox closed, forcing disconnect")
				return status.Error(codes.Unavailable, "session_terminated_by_server")
			}

			// [TRANSMIT_OVER_HTTP2]
			// Serialize and push the event into the gRPC transmit buffer.
			// gRPC handles internal flow control and HTTP/2 framing.
			if err := stream.Send(grpcmarshaller.MarshallDeliveryEvent(ev)); err != nil {
				l.Error("[STREAM] transmission error",
					slog.Any("err", err),
					slog.String("event_id", ev.GetID()),
				)
				// Returning error here triggers a gRPC status code (DataLoss) to the client.
				return status.Error(codes.DataLoss, "stream_transmission_failed")
			}

			l.Debug("[STREAM] event pushed to wire", slog.String("event_type", ev.GetKind().String()))
		}
	}
}
