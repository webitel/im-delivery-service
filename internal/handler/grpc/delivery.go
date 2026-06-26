package grpc

import (
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	impb "github.com/webitel/im-delivery-service/gen/go/delivery/v1"
	grpcinterceptors "github.com/webitel/im-delivery-service/infra/server/grpc/interceptors"
	"github.com/webitel/im-delivery-service/internal/domain/event"
	"github.com/webitel/im-delivery-service/internal/domain/model"
	"github.com/webitel/im-delivery-service/internal/handler/marshaller"
	grpcmarshaller "github.com/webitel/im-delivery-service/internal/handler/marshaller/gprc"
	"github.com/webitel/im-delivery-service/internal/service"
)

var _ impb.DeliveryServer = (*DeliveryHandler)(nil)

type DeliveryHandler struct {
	logger         *slog.Logger
	sessionManager service.SessionManager
	marshaller     marshaller.EventMarshaller
	impb.UnimplementedDeliveryServer
}

func NewDeliveryHandler(
	logger *slog.Logger,
	sessionManager service.SessionManager,
	marshaller *grpcmarshaller.Marshaller,
) *DeliveryHandler {
	return &DeliveryHandler{
		logger:         logger,
		sessionManager: sessionManager,
		marshaller:     marshaller,
	}
}

// Stream manages the lifecycle of a long-lived HTTP/2 bidirectional/server-streaming session.
func (d *DeliveryHandler) Stream(req *impb.StreamRequest, stream impb.Delivery_StreamServer) error {
	// [IDENTITY_EXTRACTION] Retrieve pre-validated contact from interceptor context
	auth, ok := grpcinterceptors.GetAuthContact(stream.Context())
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

	l.Info("[STREAM] incoming connection requолest", slog.String("version", model.ServerVersion))

	// [ACTOR_ATTACHMENT]
	// Subscribe links this specific gRPC stream to the User's Virtual Cell (Actor).
	// This ensures all events routed to the Hub for this UserID will reach this stream.
	conn, err := d.sessionManager.Attach(stream.Context(), userID, auth.Devices[0].ID)
	if err != nil {
		l.Error("[STREAM] failed to attach session",
			slog.Any("err", err),
			slog.String("device_id", auth.Devices[0].ID),
		)

		return status.Error(codes.Internal, "failed to establish session presence")
	}

	// [RESOURCE_RECLAMATION]
	// Ensure the connector is detached from the Hub when the function returns.
	// This prevents memory leaks and ensures the Hub doesn't try to send to a dead stream.
	defer func() {
		d.sessionManager.Detach(stream.Context(), userID, conn.GetID())
		l.Info("[STREAM] connection closed and resources reclaimed",
			slog.String("conn_id", conn.GetID().String()),
		)
	}()

	l.Info("[STREAM] session established", slog.String("conn_id", conn.GetID().String()))

	// [HANDSHAKE] Using the new Generic System Event with Functional Options
	welcomeEv := event.NewSystemEvent(
		userID,
		event.Connected,
		&model.ConnectedPayload{
			Ok:            true,
			ConnectionID:  conn.GetID().String(),
			ServerVersion: model.ServerVersion,
		},
		event.WithPriority[*model.ConnectedPayload](event.PriorityNormal),
	)

	// [MARSHALING]
	val, err := d.marshaller.Marshal(welcomeEv)
	if err != nil {
		l.Error("[STREAM] handshake mapping failed", slog.Any("err", err))

		return status.Error(codes.Internal, "mapping error")
	}

	// [ASSERTION] Ensure it's a Proto message
	if pb, ok := val.(*impb.ServerEvent); ok {
		if err := stream.Send(pb); err != nil {
			l.Error("[STREAM] handshake delivery failed", slog.Any("err", err))

			return err
		}
	} else {
		l.Error("[STREAM] marshaller returned invalid type", slog.String("got", fmt.Sprintf("%T", val)))

		return status.Error(codes.Internal, "unexpected data type")
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

				// [TERMINATION] Mailbox closed, send final signal
				terminationEv := event.NewSystemEvent(
					userID,
					event.DisconnectedEvent,
					&model.DisconnectedPayload{Reason: "session_closed_by_server"},
					event.WithPriority[*model.DisconnectedPayload](event.PriorityHigh),
				)

				// [MARSHALL & ASSERT]
				if val, err := d.marshaller.Marshal(terminationEv); err == nil {
					if pb, ok := val.(*impb.ServerEvent); ok {
						_ = stream.Send(pb)
					}
				}

				return status.Error(codes.Unavailable, "session_terminated_by_server")
			}

			// [TRANSMIT_OVER_HTTP2]
			// Serialize and push the event into the gRPC transmit buffer.
			// gRPC handles internal flow control and HTTP/2 framing.
			val, err := d.marshaller.Marshal(ev)
			if err != nil {
				l.Error("[STREAM] marshaling error", slog.Any("err", err))

				continue
			}

			// [ASSERT & SEND]
			if pb, ok := val.(*impb.ServerEvent); ok {
				if err := stream.Send(pb); err != nil {
					l.Error("[STREAM] transmission error", slog.Any("err", err))

					return status.Error(codes.DataLoss, "stream_transmission_failed")
				}
			} else {
				l.Error("[STREAM] type mismatch", slog.String("got", fmt.Sprintf("%T", val)))

				continue
			}

			l.Debug("[STREAM] event pushed to wire", slog.String("event_type", ev.GetKind().String()))
		}
	}
}
