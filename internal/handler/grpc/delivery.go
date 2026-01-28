package grpc

import (
	"log/slog"

	"github.com/google/uuid"
	impb "github.com/webitel/im-delivery-service/gen/go/delivery/v1"
	server "github.com/webitel/im-delivery-service/infra/server/grpc/interceptors"
	"github.com/webitel/im-delivery-service/internal/domain/event"
	"github.com/webitel/im-delivery-service/internal/domain/model"
	grpcmarshaller "github.com/webitel/im-delivery-service/internal/handler/marshaller/gprc"
	"github.com/webitel/im-delivery-service/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ impb.DeliveryServer = (*DeliveryService)(nil)

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
	// [IDENTITY_EXTRACTION] Retrieve pre-validated contact from interceptor context
	auth, ok := server.GetAuthContact(stream.Context())
	if !ok {
		return status.Error(codes.Unauthenticated, "authentication context missing")
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

	l.Info("[STREAM] incoming connection request", slog.String("version", model.ServerVersion))

	// [ACTOR_ATTACHMENT]
	// Subscribe links this specific gRPC stream to the User's Virtual Cell (Actor).
	// This ensures all events routed to the Hub for this UserID will reach this stream.
	conn, err := d.deliverer.Subscribe(stream.Context(), userID)
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
		ServerVersion: model.ServerVersion,
	})

	if err := stream.Send(grpcmarshaller.MarshallDeliveryEvent(welcomeEv)); err != nil {
		l.Error("[STREAM] handshake delivery failed", slog.Any("err", err))
		return err
	}

	// [EVENT_LOOP]
	// Main delivery loop that bridges the internal Actor mailbox with the gRPC stream.
	for {
		select {
		case <-stream.Context().Done():
			// [GHOST_CLEANUP]
			// Triggers on client disconnect, timeout, or KeepAlive failure.
			l.Info("[STREAM] client terminated connection", slog.Any("reason", stream.Context().Err()))
			return nil

		case ev, ok := <-conn.Recv():
			if !ok {
				// [TERMINATION_SENTINEL]
				// Before returning the gRPC error, we push a final System Event to the wire.
				l.Warn("[HUB] mailbox closed, sending termination event")

				terminationEv := event.NewSystemEvent(userID, event.Disconnected, event.PriorityHigh, &model.DisconnectedPayload{
					Reason: "session_closed_by_server",
				})

				// Send the "goodbye" message. We ignore the error here because if the
				// transport is already failing, we just proceed to return the status.
				_ = stream.Send(grpcmarshaller.MarshallDeliveryEvent(terminationEv))

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
