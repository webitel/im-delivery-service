package lp

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/webitel/im-delivery-service/internal/domain/model"
	lpmarshaller "github.com/webitel/im-delivery-service/internal/handler/marshaller/lp"
	"github.com/webitel/im-delivery-service/internal/service"
)

type LPHandler struct {
	deliverer service.Deliverer
}

func NewLPHandler(deliverer service.Deliverer) *LPHandler {
	return &LPHandler{
		deliverer: deliverer,
	}
}

// Poll handles the long-polling request.
// It holds the connection until an event arrives or timeout occurs.
func (h *LPHandler) Poll(w http.ResponseWriter, r *http.Request) {
	// 1. Extract Identity (UserID should be validated via middleware in production).
	userIDStr := chi.URLParam(r, "userID")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	// 2. Temporary Subscription.
	// We create a connector that will live only for the duration of this HTTP request.
	conn, err := h.deliverer.Subscribe(r.Context(), userID)
	if err != nil {
		http.Error(w, "failed to subscribe", http.StatusInternalServerError)
		return
	}

	// Ensure cleanup: remove from registry and return to pool when request finishes.
	defer h.deliverer.Unsubscribe(userID, conn.GetID())
	defer conn.Close()

	var events []model.Eventer

	// 3. Wait for data or timeout.
	select {
	case <-r.Context().Done():
		// Client disconnected.
		return

	case <-time.After(30 * time.Second):
		// Standard Long-Polling timeout to prevent hanging connections.
		w.WriteHeader(http.StatusNoContent)
		return

	case ev, ok := <-conn.Recv():
		if !ok {
			return
		}
		events = append(events, ev)

		// [OPTIONAL] Drain remaining events from buffer to provide batching.
		// This minimizes the number of subsequent HTTP requests.
	drainLoop:
		for range 15 {
			select {
			case nextEv := <-conn.Recv():
				events = append(events, nextEv)
			default:
				break drainLoop
			}
		}
	}

	// 4. Final transmission.
	data, err := lpmarshaller.MarshallEvents(events)
	if err != nil {
		http.Error(w, "marshal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}
