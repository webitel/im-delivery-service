package ws

import (
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"github.com/webitel/im-delivery-service/internal/domain/event"
	"github.com/webitel/im-delivery-service/internal/domain/model"
)

// [TERMINATE] Gracefully terminates the connection using model.DisconnectedPayload.
func (h *WSHandler) terminate(c *websocket.Conn, wsCode int, reason string) {
	// [1] Initialize the payload using the domain model.
	payload := &model.DisconnectedPayload{
		Reason: strings.ToLower(reason),
		Code:   500,
		Status: "INTERNAL_SERVER_ERROR",
	}

	// [2] Map failure scenarios to specific codes and statuses.
	lowerReason := strings.ToLower(reason)
	if strings.Contains(lowerReason, "401") || wsCode == websocket.ClosePolicyViolation {
		payload.Code = 401
		payload.Status = model.UNAUTHORIZED
	} else if strings.Contains(lowerReason, "403") {
		payload.Code = 403
		payload.Status = "FORBIDDEN"
	} else if strings.Contains(lowerReason, "timeout") {
		payload.Code = 408
		payload.Status = model.StatusTimeout
	}

	// [3] Send the system event directly, similar to how Connected is sent.
	// NewSystemEvent will wrap this into "Disconnected": { ... }
	h.sendSystem(c, uuid.Nil, event.DisconnectedEvent, payload)

	// [4] Linger to ensure the JSON event is flushed to the client.
	time.Sleep(100 * time.Millisecond)

	// [5] Send the RFC-compliant WebSocket Close Frame.
	shortReason := reason
	if len(shortReason) > 123 {
		shortReason = shortReason[:120] + "..."
	}

	_ = c.SetWriteDeadline(time.Now().Add(time.Second * 1))
	msg := websocket.FormatCloseMessage(wsCode, shortReason)
	_ = c.WriteMessage(websocket.CloseMessage, msg)

	// [6] Set read deadline to force-close any active readPump.
	_ = c.SetReadDeadline(time.Now().Add(time.Millisecond * 10))

	// [7] Physical connection closure.
	time.AfterFunc(150*time.Millisecond, func() {
		_ = c.Close()

		h.log.Debug("ws: connection terminated",
			slog.Int("http_code", payload.Code),
			slog.String("status", payload.Status),
		)
	})
}

// [SEND] Standard marshaling and transmission logic.
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

// [SEND_SYSTEM] Factory-to-socket bridge for system events.
func (h *WSHandler) sendSystem(c *websocket.Conn, uid uuid.UUID, kind event.EventKind, p any) {
	h.send(c, event.NewSystemEvent(uid, kind, p), h.log)
}
