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

// waitAuthFrame handles authentication for connections that missed header-based auth.
func (h *WSHandler) waitAuthFrame(ctx context.Context, c *websocket.Conn) {
	remote := c.RemoteAddr().String()

	// [1] Set a strict read deadline for the initial auth frame (5s).
	// This prevents the connection from hanging if the client doesn't send anything.
	_ = c.SetReadDeadline(time.Now().Add(authTimeout))

	var req struct {
		Token  string `json:"x-webitel-access"`
		Client string `json:"x-webitel-client"`
	}

	// [2] Read the JSON auth frame.
	if err := c.ReadJSON(&req); err != nil {
		h.log.Warn("ws: auth frame read failed or timed out",
			slog.String("remote", remote),
			slog.Any("err", err),
		)
		h.terminate(c, websocket.ClosePolicyViolation, "401_unauthorized")
		return
	}

	// [3] Create a dedicated context with a short timeout for the gRPC call.
	authCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	// [4] Verify credentials via gRPC account service using the timed-out context.
	auth, err := h.auther.Inspect(metadata.NewIncomingContext(authCtx, metadata.Pairs(
		"x-webitel-access", req.Token,
		"x-webitel-client", req.Client,
	)))
	if err != nil {
		st, ok := status.FromError(err)

		// [5] Handle infrastructure/timeout errors (500 range).
		if ok && (st.Code() == codes.DeadlineExceeded || st.Code() == codes.Unavailable || st.Code() == codes.Internal) {
			h.log.Error("ws: auth service slow or unavailable",
				slog.String("remote", remote),
				slog.String("grpc_code", st.Code().String()),
				slog.Any("err", err),
			)
			h.terminate(c, 500, "500_internal_error")
			return
		}

		// [6] Handle invalid credentials (401 range).
		h.log.Warn("ws: auth denied",
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

	// [7] CRITICAL: Reset the read deadline for normal message exchange.
	// Otherwise, the connection will close after the initial 5s window.
	_ = c.SetReadDeadline(time.Time{})

	h.initSession(c, auth)
}

// [INIT_SESSION] Bootstraps the message delivery pipeline.
func (h *WSHandler) initSession(c *websocket.Conn, auth *model.AuthContact) {
	uid, _ := uuid.Parse(auth.ContactID)
	remote := c.RemoteAddr().String()

	// [LIFECYCLE] Root context for this specific connection.
	sessionCtx, cancel := context.WithCancel(context.Background())

	// [ATTACH] Register connection to the delivery orchestrator.
	// Uses the first detected device ID from auth info.
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

	// [DEFER_CLEANUP] Ensure all resources are released on socket closure.
	defer func() {
		cancel()
		h.sessionManager.Detach(context.Background(), uid, cid)
		_ = h.presenceManager.Offline(context.Background(), uid, cid)
		_ = c.Close()
		log.Info("ws: session terminated")
	}()

	log.Info("ws: session established and attached")

	// [WELCOME] Notify client about successful setup.
	h.sendSystem(c, uid, event.Connected, &model.ConnectedPayload{
		Ok:           true,
		ConnectionID: cid.String(),
	})

	// [PUMPS] Start full-duplex transmission.
	go h.readPump(c, uid, cid)
	h.writePump(sessionCtx, c, sub, log)
}
