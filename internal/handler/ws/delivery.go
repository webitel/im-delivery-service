package ws

import (
	"context"
	"fmt"
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
)

const (
	// [CONFIG] Time allowed to write a message to the peer.
	writeWait = 10 * time.Second
	// [CONFIG] Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second
	// [CONFIG] Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
	// [CONFIG] Maximum message size allowed from peer.
	maxMessageSize = 512
)

// [INTERFACE_GUARD] Ensure WSHandler implements http.Handler interface
var _ http.Handler = (*WSHandler)(nil)

type WSHandler struct {
	logger     *slog.Logger
	deliverer  service.Deliverer
	auther     service.Auther
	marshaller marshaller.EventMarshaller
	upgrader   websocket.Upgrader
}

func NewWSHandler(
	logger *slog.Logger,
	deliverer service.Deliverer,
	auther service.Auther,
	marshaller *wsmarshaller.Marshaller,
) *WSHandler {
	return &WSHandler{
		logger:     logger,
		deliverer:  deliverer,
		auther:     auther,
		marshaller: marshaller,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
	}
}

// ServeHTTP manages the WebSocket handshake, authentication, and session orchestration.
func (h *WSHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// [IDENTITY_EXTRACTION] Retrieve and validate identity from metadata-enriched context
	auth, err := h.auther.Inspect(r.Context())
	if err != nil {
		h.logger.Warn("WS_AUTH_DENIED", slog.Any("err", err))
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userID, err := uuid.Parse(auth.ContactID)
	if err != nil {
		h.logger.Error("WS_INVALID_USER_ID", slog.String("contact_id", auth.ContactID))
		http.Error(w, "Invalid Identity", http.StatusBadRequest)
		return
	}

	// [UPGRADE] Protocol switch from HTTP/1.1 to WebSocket
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error("WS_UPGRADE_FAILED", slog.Any("err", err))
		return
	}

	// [ACTOR_ATTACHMENT] Link this socket to the User's delivery hub
	sub := h.deliverer.Subscribe(r.Context(), userID)

	log := h.logger.With(
		slog.String("user_id", userID.String()),
		slog.String("conn_id", sub.GetID().String()),
	)

	// [LIFECYCLE_MANAGEMENT] Ensure resources are reclaimed on connection loss
	ctx, cancel := context.WithCancel(r.Context())
	defer func() {
		// [TERMINATION_SENTINEL] Attempt to push a final Disconnected event
		terminationEv := event.NewSystemEvent(
			userID,
			event.Disconnected,
			event.PriorityHigh,
			&model.DisconnectedPayload{
				Reason: "session_terminated",
			})

		// [TERMINATION_SENTINEL]
		if val, err := h.marshaller.Marshal(terminationEv); err == nil {
			// Type assertion from any to []byte
			if data, ok := val.([]byte); ok {
				_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
				_ = conn.WriteMessage(websocket.TextMessage, data)
			}
		}

		cancel()
		h.deliverer.Unsubscribe(userID, sub.GetID())
		conn.Close()
		log.Info("WS_SESSION_TERMINATED")
	}()

	log.Info("WS_SESSION_ESTABLISHED")

	// [WELCOME_HANDSHAKE] Synchronously send the connection metadata to the client
	welcomeEv := event.NewSystemEvent(
		userID,
		event.Connected,
		event.PriorityNormal,
		&model.ConnectedPayload{
			Ok:            true,
			ConnectionID:  sub.GetID().String(),
			ServerVersion: model.ServerVersion,
		})

	// [WELCOME_HANDSHAKE]
	if val, err := h.marshaller.Marshal(welcomeEv); err == nil {
		// [ASSERTION] Cast any to []byte for WebSocket transmission
		if data, ok := val.([]byte); ok {
			_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
				log.Error("WS_WELCOME_SEND_FAILED", slog.Any("err", err))
				return
			}
		} else {
			log.Error("WS_MARSHALL_TYPE_MISMATCH", slog.String("expected", "[]byte"))
		}
	}

	// [CONCURRENCY] Spin up the I/O pumps
	go h.readPump(ctx, conn, log)
	h.writePump(ctx, conn, sub, log)
}

// readPump maintains connection health by consuming control frames and heartbeats.
func (h *WSHandler) readPump(ctx context.Context, conn *websocket.Conn, log *slog.Logger) {
	conn.SetReadLimit(maxMessageSize)
	_ = conn.SetReadDeadline(time.Now().Add(pongWait))

	// [HEARTBEAT] Reset read deadline upon receiving a Pong from the client
	conn.SetPongHandler(func(string) error {
		_ = conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		select {
		case <-ctx.Done():
			return
		default:
			// [DRAIN] Continuously read messages to process control frames (Ping/Pong/Close)
			if _, _, err := conn.NextReader(); err != nil {
				return
			}
		}
	}
}

// // writePump handles event dispatching from the internal hub to the WebSocket peer.
func (h *WSHandler) writePump(ctx context.Context, conn *websocket.Conn, sub registry.Connector, log *slog.Logger) {
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			// [SHUTDOWN] Graceful WebSocket closure sequence
			msg := websocket.FormatCloseMessage(websocket.CloseNormalClosure, "server_shutdown")
			_ = conn.WriteControl(websocket.CloseMessage, msg, time.Now().Add(writeWait))
			return

		case ev, ok := <-sub.Recv():
			if !ok {
				return
			}

			// [SERIALIZATION] Convert domain event to wire-ready JSON
			val, err := h.marshaller.Marshal(ev)
			if err != nil {
				log.Error("WS_MARSHALL_FAILED", slog.Any("err", err))
				continue
			}

			// [ASSERTION] Check if the marshaler returned []byte
			data, ok := val.([]byte)
			if !ok {
				log.Error("WS_TYPE_ASSERTION_FAILED",
					slog.String("expected", "[]byte"),
					slog.Any("received", fmt.Sprintf("%T", val)),
				)
				continue
			}

			// [TRANSMISSION] Push the data message to the peer
			_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
				log.Error("WS_WRITE_FAILED", slog.Any("err", err))
				return
			}

		case <-ticker.C:
			// [KEEP_ALIVE] Proactively send Pings to keep the connection alive through proxies
			_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
