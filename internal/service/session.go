package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/webitel/im-delivery-service/internal/domain/registry"
	"github.com/webitel/im-delivery-service/internal/store"
)

// appConfigResolveTimeout bounds the admin lookup in Attach so a slow
// admin-service can't block session attachment; the filter fails open on timeout.
const appConfigResolveTimeout = 3 * time.Second

// [SESSION_MANAGER] Interface for handling physical connection lifecycle.
type SessionManager interface {
	// Attach now requires deviceID to synchronize presence state.
	Attach(ctx context.Context, uid uuid.UUID, deviceID, appID string) (registry.Connector, error)
	Detach(ctx context.Context, uid, cid uuid.UUID)
	Close()
}

type SessionService struct {
	hub       registry.Hubber
	presence  store.PresenceStore
	appConfig AppConfigProvider
	log       *slog.Logger
}

var _ SessionManager = (*SessionService)(nil)

func NewSessionService(hub registry.Hubber, presence store.PresenceStore, appConfig AppConfigProvider, log *slog.Logger) *SessionService {
	return &SessionService{
		hub:       hub,
		presence:  presence,
		appConfig: appConfig,
		log:       log.With("service", "session_service"),
	}
}

// [ATTACH] Registers a new active session into the Hub and marks user as Online.
// The appID parameter is the application's client_id (Authorization.AppId, see
// auth.go's AuthService.Inspect), the same identifier and provenance
// device_configuration.go uses for push-config AppID lookups -- so the live-session
// and push paths resolve system-message policy against the same admin-service key.
func (s *SessionService) Attach(ctx context.Context, uid uuid.UUID, deviceID, appID string) (registry.Connector, error) {
	// Resolve the policy once at Attach time with a bounded, independent context.
	// We deliberately use context.Background() + WithTimeout instead of the session's ctx
	// because the filter closure is stored on a pooled connector and invoked later from Cell.deliver,
	// possibly after the session context is canceled. We want the resolution to complete here
	// and never re-invoke the RPC, leaving only the immutable policy's .Allows method (zero I/O).
	var filter registry.SystemMessageFilter

	if appID != "" && s.appConfig != nil {
		resolveCtx, cancel := context.WithTimeout(context.Background(), appConfigResolveTimeout)
		policy := s.appConfig.ResolvePolicy(resolveCtx, appID)

		cancel()

		filter = policy.Allows
	}

	// Per-session send buffer. The write pump drains this to the socket; 128
	// slots smooth out network jitter while cutting ~14 KB of idle buffer per
	// session versus the previous 1024. Overflow is handled by Send's
	// priority-aware backpressure, not by hoarding memory.
	conn := registry.NewConnector(ctx, uid, 128, filter)

	// [PRESENCE] Mark session as online before registering in the hub.
	// We use the provided context to respect the connection handshake timeout.
	if err := s.presence.Online(ctx, uid, conn.GetID(), deviceID); err != nil {
		s.log.Error("PRESENCE_ONLINE_FAILED",
			slog.Any("err", err),
			slog.String("uid", uid.String()),
		)

		conn.Close() // release the pooled connector -- Attach is failing, nothing else will close it

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
				slog.Any("err", err),
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
