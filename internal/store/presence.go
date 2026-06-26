package store

import (
	"context"

	"github.com/google/uuid"

	"github.com/webitel/im-delivery-service/internal/domain/model"
)

// PresenceStore defines the data layer for tracking user online status
// and managing push notification device metadata across the cluster.
type PresenceStore interface {
	// --- Session Management ---

	// Online registers an active WebSocket connection.
	// The deviceID parameter allows mapping a specific socket session (CID)
	// to a physical hardware device, preventing redundant push notifications
	// to the device that is already actively communicating via WebSockets.
	Online(ctx context.Context, uid, cid uuid.UUID, deviceID string) error

	// Offline removes a specific connection record and its associated
	// device mapping when the WebSocket is closed or lost.
	Offline(ctx context.Context, uid, cid uuid.UUID) error

	// Heartbeat extends the time-to-live (TTL) of the user's presence
	// in the cache, confirming the connection is still healthy.
	Heartbeat(ctx context.Context, uid, cid uuid.UUID) error

	// ActiveSessions retrieves all connection IDs (CIDs) currently
	// associated with a user across all distributed service instances.
	ActiveSessions(ctx context.Context, uid uuid.UUID) ([]uuid.UUID, error)

	// GetSessionDevice returns the unique hardware identifier linked
	// to a specific connection, used for smart push filtering.
	GetSessionDevice(ctx context.Context, uid, cid uuid.UUID) (string, error)

	// --- Device & Push Metadata ---

	// UserDevices retrieves cached push notification tokens and
	// platform-specific metadata for a given user.
	UserDevices(ctx context.Context, uid uuid.UUID) (*[]model.Device, error)

	// SyncDevices performs an atomic update of the user's device list,
	// ensuring the push notification cache remains consistent with the auth service.
	SyncDevices(ctx context.Context, uid uuid.UUID, devices []model.Device) error
}
