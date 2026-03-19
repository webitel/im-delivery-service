// internal/store/interface.go
package store

import (
	"context"
	"time"

	"github.com/webitel/im-delivery-service/internal/domain/event"
)

// [DELIVERY_SCHEDULER] Unified interface for task & event management.
type DeliveryScheduler interface {
	// Schedule persists the event and its delay.
	Schedule(ctx context.Context, ev event.Eventer, delay time.Duration) error
	// PullReady returns full event objects and cleans up storage atomically.
	PullReady(ctx context.Context) ([]event.Eventer, error)
}
