package grpc

import (
	"log/slog"

	"github.com/google/uuid"
	impb "github.com/webitel/im-delivery-service/gen/go/api/v1"
	"github.com/webitel/im-delivery-service/internal/domain/model"
	"github.com/webitel/im-delivery-service/internal/handler/marshaller"
	"github.com/webitel/im-delivery-service/internal/service"
)

// Interface guard
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

// Stream implements [deliveryv1.DeliveryServer].
// Stream implements [deliveryv1.DeliveryServer] and manages the real-time event lifecycle.
func (d *DeliveryService) Stream(req *impb.StreamRequest, stream impb.Delivery_StreamServer) error {
	// EXTRACT USER ID (Hardcoded for debugging, should come from auth metadata)
	userID, err := uuid.Parse("019bb6d7-8bb8-7a5c-b163-8cf8d362a474")
	if err != nil {
		return err
	}

	// SUBSCRIBE VIA SERVICE
	// Creates a unique connector for this stream and registers it in the Hub.
	conn, err := d.deliverer.Subscribe(stream.Context(), userID)
	if err != nil {
		d.logger.Error("failed to subscribe", "user_id", userID, "error", err)
		return err
	}

	// [GRACEFUL_CLEANUP] Ensure the session is removed from Hub when the stream ends.
	defer d.deliverer.Unsubscribe(userID, conn.GetID())

	d.logger.Info("stream opened", "user_id", userID, "conn_id", conn.GetID())

	if err := stream.Send(marshaller.MarshallDeliveryEvent(model.NewConnectedEvent(userID, conn.GetID().String(), "1.0.0"))); err != nil {
		d.logger.Warn("failed to send welcome event", "user_id", userID, "error", err)
		return err
	}

	// MAIN DELIVERY LOOP
	for {
		select {
		// Terminate if the client disconnects or the server context is cancelled.
		case <-stream.Context().Done():
			d.logger.Info("stream closed by client", "user_id", userID)
			return nil

		// Receive events routed to this specific user from the Hub.
		case ev, ok := <-conn.Recv():
			if !ok {
				return nil // Connector channel closed by the service
			}

			// TRANSMIT OVER HTTP/2
			if err := stream.Send(marshaller.MarshallDeliveryEvent(ev)); err != nil {
				d.logger.Warn("failed to transmit event", "user_id", userID, "error", err)
				return err
			}
		}
	}
}
