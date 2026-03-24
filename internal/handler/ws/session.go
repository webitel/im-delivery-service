package ws

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/webitel/im-delivery-service/internal/domain/event"
	"github.com/webitel/im-delivery-service/internal/domain/model"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// [WAIT_AUTH_FRAME] Handles late-binding authentication.
func (h *WSHandler) waitAuthFrame(ctx context.Context, c *websocket.Conn) {
	// [1] Skip if middleware already did the job
	if auth, ok := ctx.Value(authInfoKey).(*model.AuthContact); ok {
		h.initSession(c, auth)
		return
	}

	remote := c.RemoteAddr().String()
	_ = c.SetReadDeadline(time.Now().Add(authTimeout))

	var req struct {
		Token  string `json:"x-webitel-access"`
		Client string `json:"x-webitel-client"`
	}

	if err := c.ReadJSON(&req); err != nil {
		h.log.Warn("ws: auth frame malformed", slog.String("remote", remote), slog.Any("err", err))
		h.terminate(c, websocket.ClosePolicyViolation, "INVALID_AUTH_PAYLOAD")
		return
	}

	authCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	auth, err := h.auther.Inspect(metadata.NewIncomingContext(authCtx, metadata.Pairs(
		"x-webitel-access", req.Token,
		"x-webitel-client", req.Client,
	)))
	if err != nil {
		st, _ := status.FromError(err)

		// [LOG_CRITICAL] Detailed debug info about why identity failed
		h.log.Error("ws: late-auth inspection failed",
			slog.String("remote", remote),
			slog.Int("grpc_code", int(st.Code())),
			slog.String("grpc_status", st.Code().String()),
			slog.String("grpc_msg", st.Message()), // Complete reason (e.g. "expired")
		)

		switch st.Code() {
		case codes.Unauthenticated, codes.InvalidArgument, codes.Unknown:
			h.terminate(c, websocket.ClosePolicyViolation, "401_UNAUTHORIZED")
		case codes.PermissionDenied:
			h.terminate(c, websocket.ClosePolicyViolation, "403_FORBIDDEN")
		default:
			h.terminate(c, 1011, "500_INTERNAL_AUTH_ERROR")
		}
		return
	}

	_ = c.SetReadDeadline(time.Time{})
	h.initSession(c, auth)
}

// [INIT_SESSION] Bootstraps the message delivery pipeline.
func (h *WSHandler) initSession(c *websocket.Conn, auth *model.AuthContact) {
	uid, _ := uuid.Parse(auth.ContactID)
	sessionCtx, cancel := context.WithCancel(context.Background())

	sub, err := h.sessionManager.Attach(sessionCtx, uid, auth.Devices[0].ID)
	if err != nil {
		h.log.Error("ws: session attach failed", slog.String("uid", uid.String()), slog.Any("err", err))
		h.terminate(c, websocket.CloseInternalServerErr, "SESSION_ATTACH_ERROR")
		cancel()
		return
	}

	cid := sub.GetID()
	log := h.log.With(slog.String("uid", uid.String()), slog.String("cid", cid.String()))

	// [DEFER_CLEANUP] Managed session teardown.
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
