package ws

import (
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/webitel/im-delivery-service/internal/domain/event"
)

// [TERMINATE] Gracefully terminates the connection and sends a formatted disconnected_event.
func (h *WSHandler) terminate(c *websocket.Conn, wsCode int, reason string) {
	// [1] Define the payload structure to match the user's requirement.
	// We extract HTTP status and string representation from the reason.
	type disconnectedPayload struct {
		Reason string `json:"reason"`
		Code   int    `json:"code"`
		Status string `json:"status"`
	}

	payload := disconnectedPayload{
		Reason: strings.ToLower(reason),
		Code:   500,
		Status: "INTERNAL_SERVER_ERROR",
	}

	// [2] Map common failure scenarios to 401 Unauthorized.
	if strings.Contains(strings.ToLower(reason), "401") || wsCode == websocket.ClosePolicyViolation {
		payload.Code = 401
		payload.Status = "UNAUTHORIZED"
	} else if strings.Contains(strings.ToLower(reason), "403") {
		payload.Code = 403
		payload.Status = "FORBIDDEN"
	}

	// [3] Wrap into the specific 'disconnected_event' key.
	h.sendSystem(c, uuid.Nil, event.Disconnected, map[string]any{
		"disconnected_event": payload,
	})

	// [4] Linger to allow the JSON buffer to be flushed to the client.
	time.Sleep(60 * time.Millisecond)

	// [5] Send the RFC-compliant WebSocket Close Frame.
	shortReason := reason
	if len(shortReason) > 123 {
		shortReason = shortReason[:120] + "..."
	}

	_ = c.SetWriteDeadline(time.Now().Add(time.Second * 1))
	msg := websocket.FormatCloseMessage(wsCode, shortReason)
	_ = c.WriteMessage(websocket.CloseMessage, msg)

	// [6] Set immediate deadline to break any active read loops.
	_ = c.SetReadDeadline(time.Now().Add(time.Millisecond * 10))

	// [7] Final physical connection closure.
	time.AfterFunc(150*time.Millisecond, func() {
		_ = c.Close()
		h.log.Debug("ws: connection terminated",
			slog.Int("http_code", payload.Code),
			slog.String("status", payload.Status),
		)
	})
}

// [SEND] Marshals and writes a message to the websocket connection.
func (h *WSHandler) send(c *websocket.Conn, ev event.Eventer, log *slog.Logger) {
	raw, err := h.marshaller.Marshal(ev)
	if err != nil {
		log.Error("ws: marshal failed", slog.Any("err", err))
		return
	}

	if data, ok := raw.([]byte); ok {
		_ = c.SetWriteDeadline(time.Now().Add(writeWait))
		if err := c.WriteMessage(websocket.TextMessage, data); err != nil {
			log.Warn("ws: socket write failed", slog.Any("err", err))
		}
	}
}

// [SEND_SYSTEM] Helper to wrap any payload into a System Event structure.
func (h *WSHandler) sendSystem(c *websocket.Conn, uid uuid.UUID, kind event.EventKind, p any) {
	h.send(c, event.NewSystemEvent(uid, kind, p), h.log)
}
