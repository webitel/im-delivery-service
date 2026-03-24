package ws

import (
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/webitel/im-delivery-service/internal/domain/event"
)

// [TERMINATE] RFC-compliant shutdown.
func (h *WSHandler) terminate(c *websocket.Conn, code int, reason string) {
	if len(reason) > 123 {
		reason = reason[:120] + "..."
	}

	_ = c.SetWriteDeadline(time.Now().Add(time.Second * 1))

	// Send native Close Frame
	msg := websocket.FormatCloseMessage(code, reason)
	_ = c.WriteMessage(websocket.CloseMessage, msg)

	// Expire read deadline to kill active pumps
	_ = c.SetReadDeadline(time.Now().Add(time.Millisecond * 50))

	time.AfterFunc(250*time.Millisecond, func() {
		_ = c.Close()
	})
}

// [SEND] Internal helper for sending raw events.
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

// [SEND_SYSTEM] Re-added: helper for connected/disconnected system events.
func (h *WSHandler) sendSystem(c *websocket.Conn, uid uuid.UUID, kind event.EventKind, p any) {
	h.send(c, event.NewSystemEvent(uid, kind, p), h.log)
}
