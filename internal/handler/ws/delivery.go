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

// [INTERFACE_GUARD] Ensure WSHandler implements http.Handler at compile time.
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
			// [CORS_POLICY] Allow all origins to prevent handshake rejection.
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

// [SERVE_HTTP] Entry point for WebSocket upgrades.
func (h *WSHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// [AUDIT] Capture connection metadata for debugging.
	remote := r.RemoteAddr
	origin := r.Header.Get("Origin")

	h.log.Debug("ws: incoming handshake",
		slog.String("remote", remote),
		slog.String("origin", origin),
		slog.String("ua", r.UserAgent()),
	)

	// [1. HEADER_AUTH] Primary attempt using HTTP context (headers/query).
	authInfo, errAuth := h.auther.Inspect(r.Context())

	// [2. UPGRADE] Switch protocol from HTTP to WS.
	c, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.log.Error("ws: upgrade failed",
			slog.String("remote", remote),
			slog.Any("err", err),
		)
		return
	}

	if errAuth == nil {
		// [FAST_TRACK] User identified during handshake.
		h.log.Info("ws: session authorized via headers",
			slog.String("remote", remote),
			slog.String("user_id", authInfo.ContactID),
		)
		h.initSession(r.Context(), c, authInfo)
		return
	}

	// [3. DELAYED_AUTH] Header auth failed, wait for JSON token via socket.
	h.log.Debug("ws: header auth failed, awaiting auth frame",
		slog.String("remote", remote),
		slog.Any("reason", errAuth),
	)
	go h.waitAuthFrame(c)
}

// [WAIT_AUTH_FRAME] Handles late-binding authentication.
func (h *WSHandler) waitAuthFrame(c *websocket.Conn) {
	remote := c.RemoteAddr().String()

	if err := c.SetReadDeadline(time.Now().Add(authTimeout)); err != nil {
		h.log.Error("ws: set read deadline failed", slog.Any("err", err))
	}

	var req struct {
		Token  string `json:"x-webitel-access"`
		Client string `json:"x-webitel-client"`
	}

	// [READ_JSON] Expecting auth credentials.
	if err := c.ReadJSON(&req); err != nil {
		h.log.Warn("ws: auth frame error",
			slog.String("remote", remote),
			slog.Any("err", err),
		)
		h.terminate(c, websocket.ClosePolicyViolation, "401_unauthorized")
		return
	}

	// [CONTEXT_PREP] Mapping frame data to gRPC metadata.
	authCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	md := metadata.Pairs("x-webitel-access", req.Token)
	if req.Client != "" {
		md.Set("x-webitel-client", req.Client)
	}
	authCtx = metadata.NewIncomingContext(authCtx, md)

	// [INSPECT] Verify credentials against auth service.
	authInfo, err := h.auther.Inspect(authCtx)
	if err != nil {
		h.log.Warn("ws: delayed auth denied",
			slog.String("remote", remote),
			slog.Any("err", err),
		)
		h.terminate(c, websocket.ClosePolicyViolation, "401_invalid_token")
		return
	}

	h.log.Info("ws: session authorized via frame",
		slog.String("remote", remote),
		slog.String("user_id", authInfo.ContactID),
	)

	// [RESET] Clear deadline for normal message flow.
	_ = c.SetReadDeadline(time.Time{})
	h.initSession(context.Background(), c, authInfo)
}

// [INIT_SESSION] Lifecycle management for an active socket.
func (h *WSHandler) initSession(ctx context.Context, c *websocket.Conn, auth *model.AuthContact) {
	uid, _ := uuid.Parse(auth.ContactID)
	sCtx, cancel := context.WithCancel(ctx)

	// [SUBSCRIBE] Register connection in the deliverer.
	sub := h.deliverer.Subscribe(sCtx, uid)
	log := h.log.With(
		slog.String("user_id", uid.String()),
		slog.String("conn_id", sub.GetID().String()),
		slog.String("remote", c.RemoteAddr().String()),
	)

	defer func() {
		// [CLEANUP] Ensure resources are released on exit.
		h.sendSystem(c, uid, event.Disconnected, &model.DisconnectedPayload{
			Reason: "terminated",
			Code:   1000,
			Status: model.StatusShutdown,
		})
		cancel()
		h.deliverer.Unsubscribe(uid, sub.GetID())
		_ = c.Close()
		log.Info("ws: session terminated")
	}()

	log.Info("ws: session established")

	// [WELCOME] Notify client about successful setup.
	h.sendSystem(c, uid, event.Connected, &model.ConnectedPayload{
		Ok:            true,
		ConnectionID:  sub.GetID().String(),
		ServerVersion: model.ServerVersion,
	})

	// [PUMPS] Start full-duplex communication.
	go h.readPump(c)
	h.writePump(sCtx, c, sub, log)
}

// [READ_PUMP] Drains input and handles control frames (Ping/Pong).
func (h *WSHandler) readPump(c *websocket.Conn) {
	c.SetReadLimit(maxMessageSize)
	_ = c.SetReadDeadline(time.Now().Add(pongWait))
	c.SetPongHandler(func(string) error {
		_ = c.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		if _, _, err := c.NextReader(); err != nil {
			// [BREAK] Exit on connection loss.
			return
		}
	}
}

// [WRITE_PUMP] Forwards messages from queue to socket.
func (h *WSHandler) writePump(ctx context.Context, c *websocket.Conn, sub registry.Connector, log *slog.Logger) {
	t := time.NewTicker(pingPeriod)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			// [SHUTDOWN] Context canceled, close socket.
			msg := websocket.FormatCloseMessage(websocket.CloseNormalClosure, "shutdown")
			_ = c.WriteControl(websocket.CloseMessage, msg, time.Now().Add(writeWait))
			return

		case ev, ok := <-sub.Recv():
			if !ok {
				return
			}
			h.send(c, ev, log)

		case <-t.C:
			// [HEARTBEAT] Keep-alive ping.
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

// [TERMINATE] Closes connection with error state.
func (h *WSHandler) terminate(c *websocket.Conn, code int, reason string) {
	h.sendSystem(c, uuid.Nil, event.Disconnected, &model.DisconnectedPayload{
		Reason: reason,
		Code:   401,
		Status: model.UNAUTHORIZED,
	})

	msg := websocket.FormatCloseMessage(code, reason)
	_ = c.SetWriteDeadline(time.Now().Add(time.Second))
	_ = c.WriteControl(websocket.CloseMessage, msg, time.Now().Add(writeWait))
	_ = c.Close()
}
