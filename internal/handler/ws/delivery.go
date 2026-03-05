package ws

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/webitel/im-delivery-service/internal/domain/event"
	"github.com/webitel/im-delivery-service/internal/domain/model"
	"github.com/webitel/im-delivery-service/internal/domain/registry"
	"github.com/webitel/im-delivery-service/internal/handler/marshaller"
	wsmarshaller "github.com/webitel/im-delivery-service/internal/handler/marshaller/ws"
	"github.com/webitel/im-delivery-service/internal/service"
	"google.golang.org/grpc/metadata"
)

const (
	// [CONFIG] I/O constraints and heartbeat timing
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
	authTimeout    = 5 * time.Second
)

// [INTERFACE GUARD] Ensure WSHandler implements http.Handler at compile time.JK
var _ http.Handler = (*WSHandler)(nil)

type WSHandler struct {
	log        *slog.Logger
	deliverer  service.Deliverer
	auther     service.Auther
	marshaller marshaller.EventMarshaller
	upgrader   websocket.Upgrader
}

func NewWSHandler(
	log *slog.Logger,
	deliverer service.Deliverer,
	auther service.Auther,
	marshaller *wsmarshaller.Marshaller,
) *WSHandler {
	return &WSHandler{
		log:        log,
		deliverer:  deliverer,
		auther:     auther,
		marshaller: marshaller,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
	}
}

// [SERVE_HTTP] Main entry point.
// ---------------------------------------------------------------------------------
// [LOGIC]
// 1. First, we try to authorize via headers using the provided context.
// 2. We perform the upgrade regardless of the auth result to avoid blocking the client.
// 3. If headers failed, we start a "late-binding" authentication via WS message.
// ---------------------------------------------------------------------------------
func (h *WSHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// [1. HEADER_AUTH] Use shadowed variable for the first attempt.
	authInfo, errAuth := h.auther.Inspect(r.Context())

	// [2. UPGRADE] Separate error handling to satisfy SA4006.
	c, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.log.Error("ws: upgrade failed", slog.Any("err", err))
		return
	}

	if errAuth == nil {
		// [FAST_TRACK] Identity verified from HTTP context.
		h.initSession(r.Context(), c, authInfo)
		return
	}

	// [3. DELAYED_AUTH] Wait for client-sent token.
	go func() {
		h.waitAuthFrame(c)
	}()
}

// [WAIT_AUTH_FRAME] Blocks until the first JSON frame with a token arrives.
func (h *WSHandler) waitAuthFrame(c *websocket.Conn) {
	// Limit how long we wait for the client to send the JSON auth frame
	if err := c.SetReadDeadline(time.Now().Add(authTimeout)); err != nil {
		h.log.Error("ws: failed to set read deadline", slog.Any("err", err))
	}

	var req struct {
		Token string `json:"token"`
	}

	// Decode the expected JSON: {"token": "ey..."}
	if err := c.ReadJSON(&req); err != nil {
		h.log.Warn("ws: auth frame invalid or missing", slog.Any("err", err))
		h.terminate(c, websocket.ClosePolicyViolation, "auth_required")
		return
	}

	// [FIX] Separate the Auth Context from the Session Context.
	// We use a short-lived context ONLY for the AuthService.Inspect call.
	authCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel() // This will trigger 5s after the function starts or when Inspect finishes

	md := metadata.Pairs("x-webitel-access", req.Token)
	authCtx = metadata.NewIncomingContext(authCtx, md)

	authInfo, err := h.auther.Inspect(authCtx)
	if err != nil {
		h.log.Warn("ws: delayed auth denied", slog.Any("err", err))
		h.terminate(c, websocket.ClosePolicyViolation, "invalid_token")
		return
	}

	// [FIX] Reset the deadline before starting the long-lived session.
	_ = c.SetReadDeadline(time.Time{})

	// [CRITICAL] Pass context.Background() to initSession.
	// If you pass authCtx here, the session will die in 5 seconds because of the 'defer cancel()' above.
	h.initSession(context.Background(), c, authInfo)
}

// [INIT_SESSION] Starts the full-duplex transmission pumps.
func (h *WSHandler) initSession(ctx context.Context, c *websocket.Conn, auth *model.AuthContact) {
	uid, _ := uuid.Parse(auth.ContactID)

	// Create a new branch of context specifically for this session's lifecycle.
	// It is now rooted in Background, so it won't be affected by the auth timeout.
	sCtx, cancel := context.WithCancel(ctx)

	sub := h.deliverer.Subscribe(sCtx, uid)
	log := h.log.With(
		slog.String("user_id", uid.String()),
		slog.String("conn_id", sub.GetID().String()),
	)

	defer func() {
		// Attempt to notify the client about the disconnection
		h.sendSystem(c, uid, event.Disconnected, &model.DisconnectedPayload{Reason: "terminated"})
		cancel()
		h.deliverer.Unsubscribe(uid, sub.GetID())
		_ = c.Close()
		log.Info("ws: session terminated")
	}()

	log.Info("ws: session established")

	// Send the welcome event to the client
	h.sendSystem(c, uid, event.Connected, &model.ConnectedPayload{
		Ok:            true,
		ConnectionID:  sub.GetID().String(),
		ServerVersion: model.ServerVersion,
	})

	// Start I/O pumps
	go h.readPump(c)
	h.writePump(sCtx, c, sub, log)
}

// [I/O_PIPELINES]

func (h *WSHandler) readPump(c *websocket.Conn) {
	c.SetReadLimit(maxMessageSize)
	_ = c.SetReadDeadline(time.Now().Add(pongWait))
	c.SetPongHandler(func(string) error {
		_ = c.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		// Drains input to process control frames (Ping/Close).
		if _, _, err := c.NextReader(); err != nil {
			return
		}
	}
}

func (h *WSHandler) writePump(ctx context.Context, c *websocket.Conn, sub registry.Connector, log *slog.Logger) {
	t := time.NewTicker(pingPeriod)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			msg := websocket.FormatCloseMessage(websocket.CloseNormalClosure, "shutdown")
			_ = c.WriteControl(websocket.CloseMessage, msg, time.Now().Add(writeWait))
			return

		case ev, ok := <-sub.Recv():
			if !ok {
				return
			}
			h.send(c, ev, log)

		case <-t.C:
			_ = c.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// [TRANSMISSION_HELPERS]
func (h *WSHandler) send(c *websocket.Conn, ev event.Eventer, log *slog.Logger) {
	raw, err := h.marshaller.Marshal(ev)
	if err != nil {
		log.Error("ws: marshal failed", slog.Any("err", err))
		return
	}
	if data, ok := raw.([]byte); ok {
		_ = c.SetWriteDeadline(time.Now().Add(writeWait))
		_ = c.WriteMessage(websocket.TextMessage, data)
	}
}

func (h *WSHandler) sendSystem(c *websocket.Conn, uid uuid.UUID, kind event.EventKind, p any) {
	ev := event.NewSystemEvent(uid, kind, p)
	if raw, err := h.marshaller.Marshal(ev); err == nil {
		if data, ok := raw.([]byte); ok {
			_ = c.SetWriteDeadline(time.Now().Add(writeWait))
			_ = c.WriteMessage(websocket.TextMessage, data)
		}
	}
}

func (h *WSHandler) terminate(c *websocket.Conn, code int, reason string) {
	msg := websocket.FormatCloseMessage(code, reason)
	_ = c.WriteControl(websocket.CloseMessage, msg, time.Now().Add(writeWait))
	_ = c.Close()
}
