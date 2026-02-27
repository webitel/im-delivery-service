package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	lru "github.com/hashicorp/golang-lru/v2"
	contactv1 "github.com/webitel/im-delivery-service/gen/go/contact/v1"
	imcontact "github.com/webitel/im-delivery-service/infra/client/im-contact"
	"github.com/webitel/im-delivery-service/internal/domain/model"
)

// Contacter provides identity discovery and data enrichment for messaging peers.
type Contacter interface {
	Resolve(ctx context.Context, peer model.Peer, domainID int32) (model.Peer, error)
	ResolvePair(ctx context.Context, first, second model.Peer, domainID int32) (model.Peer, model.Peer, error)
	ResolveMany(ctx context.Context, ids []uuid.UUID, domainID int32) ([]model.Peer, error)
}

// ContactEnricher implements Contacter with local LRU caching and gRPC fetching.
type ContactEnricher struct {
	client *imcontact.Client
	cache  *lru.Cache[string, model.Peer]
	logger *slog.Logger
}

// NewContactEnricher creates a service instance with a 10k entries cache.
func NewContactEnricher(client *imcontact.Client, logger *slog.Logger) *ContactEnricher {
	cache, _ := lru.New[string, model.Peer](10000)
	return &ContactEnricher{
		client: client,
		cache:  cache,
		logger: logger,
	}
}

// Resolve pulls a single peer through the batch pipeline.
func (e *ContactEnricher) Resolve(ctx context.Context, peer model.Peer, domainID int32) (model.Peer, error) {
	res, err := e.ResolveMany(ctx, []uuid.UUID{peer.ID}, domainID)
	if err != nil || len(res) == 0 {
		return peer, err
	}
	return res[0], nil
}

// ResolvePair optimizes dual-peer lookups by grouping them into one request.
func (e *ContactEnricher) ResolvePair(ctx context.Context, first, second model.Peer, domainID int32) (model.Peer, model.Peer, error) {
	res, err := e.ResolveMany(ctx, []uuid.UUID{first.ID, second.ID}, domainID)
	if err != nil {
		return first, second, err
	}
	return res[0], res[1], nil
}

// ResolveMany is the core engine for identity enrichment with cache-aside pattern.
func (e *ContactEnricher) ResolveMany(ctx context.Context, ids []uuid.UUID, domainID int32) ([]model.Peer, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	start := time.Now()
	res := make([]model.Peer, len(ids))
	pending := make(map[string]int, len(ids))
	var lookupIDs []string

	// 1. [CACHE_LOOKUP]
	for i, id := range ids {
		if id == uuid.Nil {
			continue
		}
		key := id.String()
		if cached, ok := e.cache.Get(key); ok {
			res[i] = cached
			continue
		}
		lookupIDs = append(lookupIDs, key)
		pending[key] = i
	}

	// [HOT_PATH] Return immediately if everything is cached
	if len(lookupIDs) == 0 {
		return res, nil
	}

	// 2. [REMOTE_FETCH]
	e.fetch(ctx, lookupIDs, domainID, res, pending)

	// 3. [FALLBACK] Fill gaps for missing contacts
	for key, idx := range pending {
		res[idx] = model.Peer{
			ID:   uuid.MustParse(key),
			Type: model.PeerUser,
			Name: fmt.Sprintf("Unknown (%s)", key[:8]),
		}
	}

	// [LOGGING] Record batch execution summary
	e.logger.Debug("PEER_ENRICHMENT_COMPLETED",
		"total", len(ids),
		"fetched", len(lookupIDs),
		"domain_id", domainID,
		"duration_ms", time.Since(start).Milliseconds(),
	)

	return res, nil
}

func (e *ContactEnricher) fetch(ctx context.Context, ids []string, domainID int32, res []model.Peer, pending map[string]int) {
	resp, err := e.client.SearchContact(ctx, &contactv1.SearchContactRequest{
		Ids:      ids,
		DomainId: domainID,
		Size:     int32(len(ids)),
	})
	if err != nil {
		e.logger.Error("CONTACT_GRPC_FETCH_FAILED", "err", err, "ids", ids)
		return
	}

	for _, contact := range resp.GetContacts() {
		id := contact.GetId()
		idx, ok := pending[id]
		if !ok {
			continue
		}

		peer := e.mapToPeer(contact)
		res[idx] = peer

		e.cache.Add(id, peer)
		delete(pending, id)
	}
}

func (e *ContactEnricher) mapToPeer(c *contactv1.Contact) model.Peer {
	name := c.GetName()
	if name == "" {
		name = c.GetUsername()
	}
	return model.Peer{
		ID:     uuid.MustParse(c.GetId()),
		Type:   e.mapType(c.GetType()),
		Name:   name,
		Sub:    c.GetSubject(),
		Issuer: c.GetIssId(),
	}
}

func (e *ContactEnricher) mapType(t string) model.PeerType {
	switch t {
	case "user":
		return model.PeerUser
	case "group":
		return model.PeerGroup
	case "channel":
		return model.PeerChannel
	default:
		return model.PeerUser
	}
}
