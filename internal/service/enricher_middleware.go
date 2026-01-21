package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/webitel/im-delivery-service/internal/domain/model"
)

type enricherMiddleware struct {
	next   Enricher
	logger *slog.Logger
}

func (m *enricherMiddleware) ResolvePeers(ctx context.Context, fromID string, toID uuid.UUID, domainID int32) (model.Peer, model.Peer, error) {
	start := time.Now()

	// Call the original implementation
	from, to, err := m.next.ResolvePeers(ctx, fromID, toID, domainID)

	// [OBSERVABILITY] Log the outcome without polluting the main service
	if err != nil {
		m.logger.Error("PEER_ENRICHMENT_FAILED", "err", err, "duration", time.Since(start))
	} else {
		m.logger.Debug("PEER_ENRICHMENT_SUCCESS", "duration", time.Since(start))
	}

	return from, to, err
}

func (m *enricherMiddleware) ResolvePeer(ctx context.Context, id string, domainID int32) (model.Peer, error) {
	return m.next.ResolvePeer(ctx, id, domainID)
}
