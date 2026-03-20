// internal/handler/ws/handler.go
package ws

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/webitel/im-delivery-service/internal/domain/model"
	"github.com/webitel/im-delivery-service/internal/handler/marshaller"
	"github.com/webitel/im-delivery-service/internal/service"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
	authTimeout    = 5 * time.Second
)

type WSHandler struct {
	log             *slog.Logger
	sessionManager  service.SessionManager
	orchestrator    service.Orchestrator
	presenceManager service.PresenceManager
	auther          service.Auther
	marshaller      marshaller.EventMarshaller
	upgrader        websocket.Upgrader
}

func NewWSHandler(
	log *slog.Logger,
	sessionManager service.SessionManager,
	orchestrator service.Orchestrator,
	presence service.PresenceManager,
	auther service.Auther,
	marshaller marshaller.EventMarshaller,
) *WSHandler {
	return &WSHandler{
		log:             log.With("component", "ws_handler"),
		sessionManager:  sessionManager,
		orchestrator:    orchestrator,
		presenceManager: presence,
		auther:          auther,
		marshaller:      marshaller,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin:     func(r *http.Request) bool { return true },
		},
	}
}

// ServeHTTP implements the clean entry point for WebSocket connections.
func (h *WSHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.log.Error("ws: upgrade failed", slog.Any("err", err))
		return
	}

	// [1] Check if middleware already resolved the user.
	if auth, ok := r.Context().Value(authInfoKey).(*model.AuthContact); ok {
		h.initSession(conn, auth)
		return
	}

	// [2] Late-binding auth for clients without headers.
	h.waitAuthFrame(r.Context(), conn)
}
