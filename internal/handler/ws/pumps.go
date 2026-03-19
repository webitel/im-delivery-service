// internal/handler/ws/pumps.go
package ws

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/webitel/im-delivery-service/internal/domain/event"
	"github.com/webitel/im-delivery-service/internal/domain/registry"
)

func (h *WSHandler) readPump(conn *websocket.Conn, uid, cid uuid.UUID) {
	conn.SetReadLimit(maxMessageSize)
	_ = conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		_ = conn.SetReadDeadline(time.Now().Add(pongWait))
		_ = h.presenceManager.Heartbeat(context.Background(), uid, cid)
		return nil
	})

	for {
		var req struct {
			Type string    `json:"type"`
			EID  uuid.UUID `json:"event_id"`
		}
		if err := conn.ReadJSON(&req); err != nil {
			break
		}

		ctx := context.Background()
		switch req.Type {
		case "ack":
			_ = h.orchestrator.Ack(ctx, req.EID, cid)
		case "read":
			h.orchestrator.Dismiss(ctx, event.NewReadEvent(req.EID, uid))
		}
	}
}

func (h *WSHandler) writePump(ctx context.Context, c *websocket.Conn, sub registry.Connector, log *slog.Logger) {
	t := time.NewTicker(pingPeriod)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case ev, ok := <-sub.Recv():
			if !ok {
				return
			}
			h.send(c, ev, log)
		case <-t.C:
			_ = c.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
