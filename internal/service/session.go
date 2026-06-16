package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/webitel/im-delivery-service/internal/domain/registry"
	"github.com/webitel/im-delivery-service/internal/store"
	"github.com/webitel/webitel-go-kit/pkg/semconv"
)

// [SESSION_MANAGER] Interface for handling physical connection lifecycle.
type SessionManager interface {
	// Attach now requires deviceID to synchronize presence state.
	Attach(ctx context.Context, uid uuid.UUID, deviceID string) (registry.Connector, error)
	Detach(ctx context.Context, uid, cid uuid.UUID)
	Close()
}

type SessionService struct {
	hub      registry.Hubber
	presence store.PresenceStore
	log      *slog.Logger
}

var _ SessionManager = (*SessionService)(nil)

func NewSessionService(hub registry.Hubber, presence store.PresenceStore, log *slog.Logger) *SessionService {
	return &SessionService{
		hub:      hub,
		presence: presence,
		log:      log.With("service", "session_service"),
	}
}

// [ATTACH] Registers a new active session into the Hub and marks user as Online.
func (s *SessionService) Attach(ctx context.Context, uid uuid.UUID, deviceID string) (registry.Connector, error) {
	conn := registry.NewConnector(ctx, uid, 1024)

	// [PRESENCE] Mark session as online before registering in the hub.
	// We use the provided context to respect the connection handshake timeout.
	if err := s.presence.Online(ctx, uid, conn.GetID(), deviceID); err != nil {
		s.log.Error("PRESENCE_ONLINE_FAILED",
			slog.Any(semconv.ErrorKey, err),
			slog.String("uid", uid.String()),
		)
		return nil, err
	}

	s.hub.Register(conn)
	s.log.Debug("SESSION_ATTACHED",
		slog.String("uid", uid.String()),
		slog.String("cid", conn.GetID().String()),
		slog.String("did", deviceID),
	)

	return conn, nil
}

// [DETACH] Removes the session from Hub and updates presence state to offline.
func (s *SessionService) Detach(ctx context.Context, uid, cid uuid.UUID) {
	s.hub.Unregister(uid, cid)

	// [ASYNC] Decouple presence update to avoid blocking the transport layer.
	go func() {
		detachCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.presence.Offline(detachCtx, uid, cid); err != nil {
			s.log.Error("PRESENCE_OFFLINE_SYNC_FAILED",
				slog.Any(semconv.ErrorKey, err),
				slog.String("cid", cid.String()),
			)
		}
	}()
}

func (s *SessionService) Close() {
	if s.hub != nil {
		s.hub.Shutdown()
	}
}
