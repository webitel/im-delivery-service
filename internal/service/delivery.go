package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/webitel/im-delivery-service/internal/domain/registry"
)

// [DELIVERY_SERVICE] PRIMARY INTERFACE FOR TRANSPORT HANDLERS (gRPC/Websocket)
type Deliverer interface {
	Subscribe(ctx context.Context, userID uuid.UUID) (registry.Connector, error)
	Unsubscribe(userID, connID uuid.UUID)
}

// [IMPLEMENTATION] PRIVATE TO ENFORCE INTERFACE USAGE
type DeliveryService struct {
	hub registry.Hubber
}

// NewDeliveryService returns a production-ready instance of the service.
func NewDeliveryService(hub registry.Hubber) *DeliveryService {
	return &DeliveryService{
		hub: hub,
	}
}

// [SUBSCRIBE] HANDLES CONNECTION LIFECYCLE INITIATION
func (s *DeliveryService) Subscribe(ctx context.Context, userID uuid.UUID) (registry.Connector, error) {
	// [STRATEGY] We can adjust buffer size based on Platform or User Priority from meta
	// In the future, StreamRequest settings can be passed here as well.
	const defaultBufferSize = 1024

	// 1. Create a connector (Internal logic uses sync.Pool for zero-allocation)
	conn := registry.NewConnector(ctx, userID, defaultBufferSize)

	// 2. Attach to the sharded dispatcher
	s.hub.Register(conn)

	// 3. Return the connector for the gRPC handler to start streaming
	return conn, nil
}

// [UNSUBSCRIBE] TRIGGERS CLEANUP AND OBJECT RECYCLING
func (s *DeliveryService) Unsubscribe(userID, connID uuid.UUID) {
	// Hub.Unregister will call conn.Close(), which resets the object
	// and puts it back into model.connectPool.
	s.hub.Unregister(userID, connID)
}
