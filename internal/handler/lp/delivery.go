package lp

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/webitel/im-delivery-service/internal/domain/event"
	"github.com/webitel/im-delivery-service/internal/handler/marshaller"
	lpmarshaller "github.com/webitel/im-delivery-service/internal/handler/marshaller/lp"
	"github.com/webitel/im-delivery-service/internal/service"
)

type LPHandler struct {
	logger     *slog.Logger
	deliverer  service.Deliverer
	marshaller marshaller.BatchMarshaller
}

func NewLPHandler(
	logger *slog.Logger,
	deliverer service.Deliverer,
	marshaller *lpmarshaller.Marshaller,
) *LPHandler {
	return &LPHandler{
		logger:     logger,
		deliverer:  deliverer,
		marshaller: marshaller,
	}
}

// Poll handles the long-polling request cycle.
func (h *LPHandler) Poll(w http.ResponseWriter, r *http.Request) {
	// [IDENTITY_EXTRACTION]
	userIDStr := chi.URLParam(r, "userID")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		h.logger.Warn("LP_INVALID_USER_ID", slog.String("raw_id", userIDStr))
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	// [SUBSCRIPTION] Create a transient connector for this request.
	conn, err := h.deliverer.Subscribe(r.Context(), userID)
	if err != nil {
		h.logger.Error("LP_SUBSCRIPTION_FAILED", slog.Any("err", err), slog.String("user_id", userID.String()))
		http.Error(w, "failed to subscribe", http.StatusInternalServerError)
		return
	}

	// [LIFECYCLE] Cleanup connector resources on request completion.
	defer h.deliverer.Unsubscribe(userID, conn.GetID())
	defer conn.Close()

	var events []event.Eventer

	// [WAIT_STRATEGY] Block until event arrives or timeout occurs.
	select {
	case <-r.Context().Done():
		return

	case <-time.After(30 * time.Second):
		w.WriteHeader(http.StatusNoContent)
		return

	case ev, ok := <-conn.Recv():
		if !ok {
			return
		}
		events = append(events, ev)

		// [BATCHING] Proactively drain the buffer to reduce HTTP roundtrips.
	drainLoop:
		for range 15 {
			select {
			case nextEv, nextOk := <-conn.Recv():
				if !nextOk {
					break drainLoop
				}
				events = append(events, nextEv)
			default:
				break drainLoop
			}
		}
	}

	// [SERIALIZATION] Use the batch marshaller interface.
	val, err := h.marshaller.MarshalBatch(events)
	if err != nil {
		h.logger.Error("LP_MARSHALL_FAILED", slog.Any("err", err))
		http.Error(w, "internal serialization error", http.StatusInternalServerError)
		return
	}

	// [ASSERTION] Ensure result is ready for wire transmission.
	data, ok := val.([]byte)
	if !ok {
		h.logger.Error("LP_TYPE_MISMATCH", slog.Any("received", val))
		http.Error(w, "unexpected data format", http.StatusInternalServerError)
		return
	}

	// [TRANSMISSION] Final response delivery.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}
