// internal/handler/ws/helpers.go
package ws

import (
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/webitel/im-delivery-service/internal/domain/event"
)

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
	h.send(c, event.NewSystemEvent(uid, kind, p), h.log)
}

func (h *WSHandler) terminate(c *websocket.Conn, code int, reason string) {
	_ = c.WriteControl(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(code, reason),
		time.Now().Add(time.Second))
	_ = c.Close()
}
