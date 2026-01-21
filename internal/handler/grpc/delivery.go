package grpc

import (
	"log/slog"
	"time"

	"github.com/google/uuid"
	impb "github.com/webitel/im-delivery-service/gen/go/delivery/v1"
	"github.com/webitel/im-delivery-service/internal/domain/model"
	grpcmarshaller "github.com/webitel/im-delivery-service/internal/handler/marshaller/gprc"
	"github.com/webitel/im-delivery-service/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	serverVersion = "v1"
	maxBatchSize  = 20
	batchTimeout  = 30 * time.Millisecond
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

func (d *DeliveryService) Stream(req *impb.StreamRequest, stream impb.Delivery_StreamServer) error {
	ctx := stream.Context()
	startTime := time.Now()

	// [SECURITY] In production, extract userID from metadata/JWT
	// For now, keeping your hardcoded ID for consistency
	userID := uuid.MustParse("019bb6d7-8bb8-7a5c-b163-8cf8d362a474")

	// Create a scoped logger for this specific session
	l := d.logger.With(
		slog.String("user_id", userID.String()),
		slog.String("version", serverVersion),
	)

	// [SUBSCRIPTION]
	conn, err := d.deliverer.Subscribe(ctx, userID)
	if err != nil {
		l.Error("SUBSCRIPTION_REJECTED", slog.Any("err", err))
		return status.Error(codes.Internal, "failed to establish connection session")
	}

	connID := conn.GetID()
	l = l.With(slog.String("conn_id", connID.String()))

	// [CLEANUP] Ensure resources are released on disconnect
	defer func() {
		d.deliverer.Unsubscribe(userID, connID)
		l.Info("STREAM_TERMINATED", slog.Duration("session_duration", time.Since(startTime)))
	}()

	l.Info("STREAM_ESTABLISHED")

	// [INITIAL_HANDSHAKE] Send Welcome Event immediately
	welcome := model.NewConnectedEvent(userID, connID.String(), serverVersion)
	if err := stream.Send(grpcmarshaller.MarshallDeliveryEvent(welcome)); err != nil {
		l.Warn("HANDSHAKE_FAILED", slog.Any("err", err))
		return err
	}

	// Internal state for batching
	for {
		select {
		case <-ctx.Done():
			l.Debug("CLIENT_DISCONNECTED_CONTEXT_DONE")
			return nil

		case ev, ok := <-conn.Recv():
			if !ok {
				l.Warn("HUB_FORCED_DISCONNECT", slog.String("reason", "inbound_channel_closed"))
				return status.Error(codes.Unavailable, "session_terminated_by_server")
			}

			// [COALESCING_STRATEGY] Start batching upon arrival of the first message
			batch := make([]*impb.ServerEvent, 0, maxBatchSize)
			batch = append(batch, grpcmarshaller.MarshallDeliveryEvent(ev))

			// Use a reusable timer logic
			timer := time.NewTimer(batchTimeout)
			reason := "capacity_reached"

		collect:
			for len(batch) < maxBatchSize {
				select {
				case ev, ok := <-conn.Recv():
					if !ok {
						reason = "upstream_closed"
						break collect
					}
					batch = append(batch, grpcmarshaller.MarshallDeliveryEvent(ev))

				case <-timer.C:
					reason = "timeout_reached"
					break collect

				case <-ctx.Done():
					if !timer.Stop() {
						select {
						case <-timer.C:
						default:
						}
					}
					return nil
				}
			}

			// Securely stop timer to prevent memory leaks
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}

			// [METRICS_LOGGING] Trace batch efficiency
			if len(batch) > 1 {
				l.Debug("BATCH_COALESCED",
					slog.Int("count", len(batch)),
					slog.String("trigger", reason),
				)
			}

			// [TRANSMISSION] Flushing the batch to gRPC stream
			for i, ev := range batch {
				if err := stream.Send(ev); err != nil {
					l.Error("TRANSMISSION_ERROR",
						slog.Any("err", err),
						slog.Int("failed_at_index", i),
						slog.Int("total_batch", len(batch)),
					)
					return status.Error(codes.DataLoss, "stream_transmission_failed")
				}
			}
		}
	}
}
