package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/webitel/im-delivery-service/internal/domain/model"
)

// EnricherMiddleware implements [DECORATOR_PATTERN] to add observability
// to the enrichment process without touching business logic.
type EnricherMiddleware struct {
	Next   Enricher
	Logger *slog.Logger
}

// NewEnricherMiddleware creates a new logging decorator for the Enricher.
func NewEnricherMiddleware(next Enricher, logger *slog.Logger) Enricher {
	return &EnricherMiddleware{
		Next:   next,
		Logger: logger,
	}
}

// ResolvePeers wraps the concurrent enrichment with execution timing and outcome logging.
func (m *EnricherMiddleware) ResolvePeers(ctx context.Context, from, to model.Peer, domainID int32) (model.Peer, model.Peer, error) {
	start := time.Now()

	// [EXECUTION] Pass enriched objects to the core service
	f, t, err := m.Next.ResolvePeers(ctx, from, to, domainID)

	// [OBSERVABILITY] Scoped logging for performance auditing
	duration := time.Since(start)

	if err != nil {
		m.Logger.Error("PEER_ENRICHMENT_BATCH_FAILED",
			"err", err,
			"from_id", from.ID,
			"to_id", to.ID,
			"duration_ms", duration.Milliseconds(),
		)
	} else {
		m.Logger.Debug("PEER_ENRICHMENT_BATCH_COMPLETED",
			"duration_ms", duration.Milliseconds(),
			"domain_id", domainID,
		)
	}

	return f, t, err
}

// ResolvePeer wraps a single peer enrichment lookup.
func (m *EnricherMiddleware) ResolvePeer(ctx context.Context, peer model.Peer, domainID int32) (model.Peer, error) {
	start := time.Now()

	res, err := m.Next.ResolvePeer(ctx, peer, domainID)
	if err != nil {
		m.Logger.Warn("SINGLE_PEER_ENRICHMENT_FAILED",
			"peer_id", peer.ID,
			"peer_type", peer.Type,
			"err", err,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	}

	return res, err
}
