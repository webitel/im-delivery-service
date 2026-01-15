package ws

import (
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	wsmarshaller "github.com/webitel/im-delivery-service/internal/handler/marshaller/ws"
	"github.com/webitel/im-delivery-service/internal/service"
)

type WSHandler struct {
	logger    *slog.Logger
	deliverer service.Deliverer
	upgrader  websocket.Upgrader
}

func NewWSHandler(logger *slog.Logger, deliverer service.Deliverer) *WSHandler {
	return &WSHandler{
		logger:    logger,
		deliverer: deliverer,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true }, // Security: adjust for production
		},
	}
}

func (h *WSHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 1. EXTRACT USER ID (In production: from JWT/Cookie)
	userID := uuid.MustParse("019bb6d7-8bb8-7a5c-b163-8cf8d362a474")

	// 2. UPGRADE TO WEBSOCKET
	ws, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error("ws upgrade failed", "error", err)
		return
	}
	defer ws.Close()

	// 3. SUBSCRIBE VIA THE SAME SERVICE
	conn, err := h.deliverer.Subscribe(r.Context(), userID)
	if err != nil {
		return
	}
	defer h.deliverer.Unsubscribe(userID, conn.GetID())

	h.logger.Info("ws opened", "user_id", userID, "conn_id", conn.GetID())

	// 4. MAIN WS PUMP LOOP
	for {
		select {
		case <-r.Context().Done():
			return
		case ev, ok := <-conn.Recv():
			if !ok {
				return
			}

			data, err := wsmarshaller.MarshallDeliveryEvent(ev)
			if err != nil {
				h.logger.Error("failed to marshal ws event", "error", err)
				continue
			}

			if err := ws.WriteMessage(websocket.TextMessage, data); err != nil {
				h.logger.Warn("ws send failed", "error", err)
				return
			}
		}
	}
}
