package ws

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/webitel/im-delivery-service/internal/domain/event"
	"github.com/webitel/im-delivery-service/internal/domain/model"
	"github.com/webitel/webitel-go-kit/pkg/semconv"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// [WAIT_AUTH_FRAME] Now waits 5s if only client_id was provided in headers.
func (h *WSHandler) waitAuthFrame(ctx context.Context, c *websocket.Conn) {
	// [1] If Fast-track was successful in middleware
	if auth, ok := ctx.Value(authInfoKey).(*model.AuthContact); ok {
		h.log.Info("ws: using fast-track auth from middleware",
			slog.String("uid", auth.ContactID),
			slog.String("remote", c.RemoteAddr().String()))
		h.initSession(c, auth)
		return
	}

	// [2] Late-binding: Wait for the first JSON message
	h.log.Debug("ws: waiting for auth payload", slog.String("remote", c.RemoteAddr().String()))
	_ = c.SetReadDeadline(time.Now().Add(authTimeout)) // This is your 5s timeout

	var req struct {
		Token  string `json:"x-webitel-access"`
		Client string `json:"x-webitel-client"`
	}

	if err := c.ReadJSON(&req); err != nil {
		h.log.Warn("ws: auth payload timeout or invalid", slog.Any(semconv.ErrorKey, err))
		h.terminate(c, websocket.ClosePolicyViolation, "401_unauthorized")
		return
	}

	md := metadata.Pairs("x-webitel-access", req.Token)
	if req.Client != "" {
		md.Set("x-webitel-client", req.Client)
	}

	authCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	// Inspect the token (and optional client ID) via the authentication service
	auth, err := h.auther.Inspect(metadata.NewIncomingContext(authCtx, md))
	if err != nil {
		st, _ := status.FromError(err)
		h.log.Error("ws: late-auth inspection failed", slog.String(semconv.ErrorKey, st.Message()))

		switch st.Code() {
		case codes.Unauthenticated, codes.InvalidArgument, codes.Unknown:
			h.terminate(c, websocket.ClosePolicyViolation, "401_unauthorized")
		default:
			h.terminate(c, 1011, "500_internal_server_error")
		}
		return
	}

	// [4] Success: clear deadline and start session
	_ = c.SetReadDeadline(time.Time{})
	h.initSession(c, auth)
}

func (h *WSHandler) initSession(c *websocket.Conn, auth *model.AuthContact) {
	uid, _ := uuid.Parse(auth.ContactID)
	sessionCtx, cancel := context.WithCancel(context.Background())

	sub, err := h.sessionManager.Attach(sessionCtx, uid, auth.Devices[0].ID)
	if err != nil {
		h.log.Error("ws: session attach failed", slog.Any(semconv.ErrorKey, err))
		h.terminate(c, 1011, "SESSION_ATTACH_ERROR")
		cancel()
		return
	}

	cid := sub.GetID()
	log := h.log.With(slog.String("uid", uid.String()), slog.String("cid", cid.String()))

	defer func() {
		cancel()
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cleanupCancel()
		h.sessionManager.Detach(cleanupCtx, uid, cid)
		_ = h.presenceManager.Offline(cleanupCtx, uid, cid)
		_ = c.Close()
		log.Info("ws: session terminated")
	}()

	log.Info("ws: session established")
	h.sendSystem(c, uid, event.Connected, &model.ConnectedPayload{
		Ok:           true,
		ConnectionID: cid.String(),
	})

	go h.readPump(c, uid, cid)
	h.writePump(sessionCtx, c, sub, log)
}
