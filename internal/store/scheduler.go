// internal/store/interface.go
package store

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// DeliveryScheduler manages persistent, time-delayed tasks.
// It ensures that even if a node fails, the scheduled delivery check will
// be picked up by another instance once the timeout expires.
type DeliveryScheduler interface {
	// Schedule adds an event for a delayed delivery check.
	Schedule(ctx context.Context, eid, uid uuid.UUID, delay time.Duration) error

	// PullReady retrieves and atomically removes tasks that reached their execution time.
	PullReady(ctx context.Context) ([]ScheduledTask, error)
}

// ScheduledTask carries the minimum identity required to process a push check.
type ScheduledTask struct {
	EventID uuid.UUID
	UserID  uuid.UUID
}
