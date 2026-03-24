package ws

import (
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/webitel/im-delivery-service/internal/domain/event"
)

// [TERMINATE] Sends a disconnect event, then a close frame, and finally tears down TCP.
func (h *WSHandler) terminate(c *websocket.Conn, code int, reason string) {
	// [1] Safety check for RFC reason length (max 123 bytes)
	shortReason := reason
	if len(shortReason) > 123 {
		shortReason = shortReason[:120] + "..."
	}

	// [2] Send a high-level JSON event so the client application can parse the reason easily.
	// We use a nil UUID if we're not sure about the user's identity yet.
	h.sendSystem(c, uuid.Nil, event.Disconnected, map[string]string{
		"reason": reason,
		"code":   string(rune(code)),
	})

	// [3] Give the JSON message a moment to leave the userspace buffer.
	time.Sleep(50 * time.Millisecond)

	// [4] Send the native WebSocket Close Frame.
	_ = c.SetWriteDeadline(time.Now().Add(time.Second * 1))
	msg := websocket.FormatCloseMessage(code, shortReason)
	_ = c.WriteMessage(websocket.CloseMessage, msg)

	// [5] Force-expire read deadline to kill any active readPump goroutines.
	_ = c.SetReadDeadline(time.Now().Add(time.Millisecond * 10))

	// [6] Final physical closure after a short linger.
	time.AfterFunc(150*time.Millisecond, func() {
		_ = c.Close()
		h.log.Debug("ws: connection physically closed", slog.String("reason", reason))
	})
}

// [SEND] Marshalling and writing a text message to the socket.
func (h *WSHandler) send(c *websocket.Conn, ev event.Eventer, log *slog.Logger) {
	raw, err := h.marshaller.Marshal(ev)
	if err != nil {
		log.Error("ws: marshal failed", slog.Any("err", err))
		return
	}

	if data, ok := raw.([]byte); ok {
		_ = c.SetWriteDeadline(time.Now().Add(writeWait))
		if err := c.WriteMessage(websocket.TextMessage, data); err != nil {
			log.Warn("ws: failed to send message", slog.Any("err", err))
		}
	}
}

// [SEND_SYSTEM] Helper to wrap payloads into system events.
func (h *WSHandler) sendSystem(c *websocket.Conn, uid uuid.UUID, kind event.EventKind, p any) {
	h.send(c, event.NewSystemEvent(uid, kind, p), h.log)
}
