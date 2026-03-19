// internal/handler/ws/session.go
package ws

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/webitel/im-delivery-service/internal/domain/event"
	"github.com/webitel/im-delivery-service/internal/domain/model"
	"google.golang.org/grpc/metadata"
)

func (h *WSHandler) waitAuthFrame(ctx context.Context, c *websocket.Conn) {
	_ = c.SetReadDeadline(time.Now().Add(authTimeout))

	// Define structure to accept both access token and client ID
	var req struct {
		Token  string `json:"x-webitel-access"`
		Client string `json:"x-webitel-client"`
	}

	if err := c.ReadJSON(&req); err != nil {
		h.terminate(c, websocket.ClosePolicyViolation, "401_unauthorized")
		return
	}

	auth, err := h.auther.Inspect(metadata.NewIncomingContext(ctx, metadata.Pairs(
		"x-webitel-access", req.Token,
		"x-webitel-client", req.Client,
	)))
	if err != nil {
		h.terminate(c, websocket.ClosePolicyViolation, "401_invalid_token")
		return
	}

	_ = c.SetReadDeadline(time.Time{})
	h.initSession(c, auth)
}

func (h *WSHandler) initSession(c *websocket.Conn, auth *model.AuthContact) {
	uid, _ := uuid.Parse(auth.ContactID)
	sessionCtx, cancel := context.WithCancel(context.Background())

	sub, err := h.sessionManager.Attach(sessionCtx, uid, auth.Devices[0].ID)
	if err != nil {
		h.terminate(c, websocket.CloseInternalServerErr, "session_attach_failed")
		cancel()
		return
	}

	cid := sub.GetID()
	log := h.log.With("uid", uid, "cid", cid)

	defer func() {
		cancel()
		h.sessionManager.Detach(context.Background(), uid, cid)
		_ = h.presenceManager.Offline(context.Background(), uid, cid)
		_ = c.Close()
		log.Info("session_terminated")
	}()

	log.Info("session_established")

	h.sendSystem(c, uid, event.Connected, &model.ConnectedPayload{
		Ok:           true,
		ConnectionID: cid.String(),
	})

	go h.readPump(c, uid, cid)
	h.writePump(sessionCtx, c, sub, log)
}
