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

// [READ_PUMP] Continuous read loop.
func (h *WSHandler) readPump(conn *websocket.Conn, uid, cid uuid.UUID) {
	defer func() { _ = conn.Close() }()

	conn.SetReadLimit(maxMessageSize)
	_ = conn.SetReadDeadline(time.Now().Add(pongWait))

	conn.SetPongHandler(func(string) error {
		_ = conn.SetReadDeadline(time.Now().Add(pongWait))
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			_ = h.presenceManager.Heartbeat(ctx, uid, cid)
		}()
		return nil
	})

	for {
		var req struct {
			Type string    `json:"type"`
			EID  uuid.UUID `json:"event_id"`
		}
		if err := conn.ReadJSON(&req); err != nil {
			break // Loop breaks immediately if terminate() is called.
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

// [WRITE_PUMP] Outbound event stream.
func (h *WSHandler) writePump(ctx context.Context, c *websocket.Conn, sub registry.Connector, log *slog.Logger) {
	t := time.NewTicker(pingPeriod)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case ev, ok := <-sub.Recv():
			if !ok { return }
			h.send(c, ev, log)
		case <-t.C:
			_ = c.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}