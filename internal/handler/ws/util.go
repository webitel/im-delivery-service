package ws

import (
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/webitel/im-delivery-service/internal/domain/event"
)

// [TERMINATE] Gracefully terminates the connection with a custom event.
func (h *WSHandler) terminate(c *websocket.Conn, wsCode int, reason string) {
	// [1] Extract HTTP-like status code from the reason string (e.g., "401_UNAUTHORIZED" -> "401")
	// If no HTTP code is found, we fallback to a string representation of the WS code.
	httpCode := "500"
	if strings.Contains(reason, "401") {
		httpCode = "401"
	} else if strings.Contains(reason, "403") {
		httpCode = "403"
	} else if wsCode == websocket.ClosePolicyViolation {
		httpCode = "401"
	}

	// [2] Send Disconnected system event with the mapped HTTP code.
	h.sendSystem(c, uuid.Nil, event.Disconnected, map[string]string{
		"reason": reason,
		"code":   httpCode,
	})

	// [3] Linger to allow the JSON buffer to be transmitted.
	time.Sleep(60 * time.Millisecond)

	// [4] Prepare and send the RFC-compliant WebSocket Close Frame.
	shortReason := reason
	if len(shortReason) > 123 {
		shortReason = shortReason[:120] + "..."
	}

	_ = c.SetWriteDeadline(time.Now().Add(time.Second * 1))
	msg := websocket.FormatCloseMessage(wsCode, shortReason)
	_ = c.WriteMessage(websocket.CloseMessage, msg)

	// [5] Set immediate deadline to interrupt any active read pumps.
	_ = c.SetReadDeadline(time.Now().Add(time.Millisecond * 10))

	// [6] Final physical connection tear down.
	time.AfterFunc(150*time.Millisecond, func() {
		_ = c.Close()
		h.log.Debug("ws: connection terminated",
			slog.String("http_code", httpCode),
			slog.Int("ws_code", wsCode),
		)
	})
}

// [SEND] Marshals and writes a message to the websocket.
func (h *WSHandler) send(c *websocket.Conn, ev event.Eventer, log *slog.Logger) {
	raw, err := h.marshaller.Marshal(ev)
	if err != nil {
		log.Error("ws: marshal failed", slog.Any("err", err))
		return
	}

	if data, ok := raw.([]byte); ok {
		_ = c.SetWriteDeadline(time.Now().Add(writeWait))
		if err := c.WriteMessage(websocket.TextMessage, data); err != nil {
			log.Warn("ws: write failed", slog.Any("err", err))
		}
	}
}

// [SEND_SYSTEM] Helper to wrap payloads into system events.
func (h *WSHandler) sendSystem(c *websocket.Conn, uid uuid.UUID, kind event.EventKind, p any) {
	h.send(c, event.NewSystemEvent(uid, kind, p), h.log)
}
