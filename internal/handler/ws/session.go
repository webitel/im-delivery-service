package ws

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/webitel/im-delivery-service/internal/domain/event"
	"github.com/webitel/im-delivery-service/internal/domain/model"
	"google.golang.org/grpc/metadata"
)

func (h *WSHandler) waitAuthFrame(ctx context.Context, c *websocket.Conn) {
	remote := c.RemoteAddr().String()
	_ = c.SetReadDeadline(time.Now().Add(authTimeout))

	var req struct {
		Token  string `json:"x-webitel-access"`
		Client string `json:"x-webitel-client"`
	}

	if err := c.ReadJSON(&req); err != nil {
		h.log.Warn("ws: auth frame read failed",
			slog.String("remote", remote),
			slog.Any("err", err),
		)
		h.terminate(c, websocket.ClosePolicyViolation, "401_unauthorized")
		return
	}

	// [INSPECT] Call gRPC service to verify credentials
	auth, err := h.auther.Inspect(metadata.NewIncomingContext(ctx, metadata.Pairs(
		"x-webitel-access", req.Token,
		"x-webitel-client", req.Client,
	)))
	if err != nil {
		h.log.Error("ws: delayed auth denied",
			slog.String("remote", remote),
			slog.Any("err", err),
		)
		h.terminate(c, websocket.ClosePolicyViolation, "401_invalid_token")
		return
	}

	h.log.Info("ws: delayed auth success",
		slog.String("user_id", auth.ContactID),
		slog.String("remote", remote),
	)

	_ = c.SetReadDeadline(time.Time{})
	h.initSession(c, auth)
}

func (h *WSHandler) initSession(c *websocket.Conn, auth *model.AuthContact) {
	uid, _ := uuid.Parse(auth.ContactID)
	remote := c.RemoteAddr().String()

	// Create context for the whole session lifecycle
	sessionCtx, cancel := context.WithCancel(context.Background())

	// [ATTACH] Register session in manager
	sub, err := h.sessionManager.Attach(sessionCtx, uid, auth.Devices[0].ID)
	if err != nil {
		h.log.Error("ws: session attach failed",
			slog.String("user_id", uid.String()),
			slog.Any("err", err),
		)
		h.terminate(c, websocket.CloseInternalServerErr, "session_attach_failed")
		cancel()
		return
	}

	cid := sub.GetID()
	log := h.log.With(
		slog.String("uid", uid.String()),
		slog.String("cid", cid.String()),
		slog.String("remote", remote),
	)

	defer func() {
		cancel()
		h.sessionManager.Detach(context.Background(), uid, cid)
		_ = h.presenceManager.Offline(context.Background(), uid, cid)
		_ = c.Close()
		log.Info("ws: session terminated")
	}()

	log.Info("ws: session established and attached")

	h.sendSystem(c, uid, event.Connected, &model.ConnectedPayload{
		Ok:           true,
		ConnectionID: cid.String(),
	})

	go h.readPump(c, uid, cid)
	h.writePump(sessionCtx, c, sub, log)
}
