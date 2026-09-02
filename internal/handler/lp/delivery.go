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
	logger         *slog.Logger
	sessionManager service.SessionManager
	marshaller     marshaller.BatchMarshaller
}

func NewLPHandler(
	logger *slog.Logger,
	sessionManager service.SessionManager,
	marshaller *lpmarshaller.Marshaller,
) *LPHandler {
	return &LPHandler{
		logger:         logger,
		sessionManager: sessionManager,
		marshaller:     marshaller,
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
	// TODO pass real deviceID and handle error
	conn, _ := h.sessionManager.Attach(r.Context(), userID, uuid.New().String(), "")

	// [LIFECYCLE] Cleanup connector resources on request completion.
	// Defer order matters: defers run LIFO, so we defer Close first (runs last at unwind)
	// and Detach second (runs first at unwind). This ensures the connector is removed
	// from the Cell's sessions map before it's closed and recycled to the pool, preventing
	// a race where concurrent delivery on another goroutine observes a torn/reused object.
	defer conn.Close()
	defer h.sessionManager.Detach(r.Context(), userID, conn.GetID())

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
	val, err := h.marshaller.MarshalBatch(events, userID)
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
