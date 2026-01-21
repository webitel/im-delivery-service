package service

import (
	"context"

	"github.com/google/uuid"
	lru "github.com/hashicorp/golang-lru/v2"
	contactv1 "github.com/webitel/im-delivery-service/gen/go/contact/v1"
	imcontact "github.com/webitel/im-delivery-service/infra/webitel/im-contact"
	"github.com/webitel/im-delivery-service/internal/domain/model"
	"golang.org/x/sync/errgroup"
)

// Enricher defines the contract for peer data augmentation
type Enricher interface {
	ResolvePeers(ctx context.Context, fromID string, toID uuid.UUID, domainID int32) (from, to model.Peer, err error)
	ResolvePeer(ctx context.Context, id string, domainID int32) (model.Peer, error)
}

type PeerEnricher struct {
	contacts *imcontact.Client
	cache    *lru.Cache[string, model.Peer]
}

// NewPeerEnricherService initializes a thread-safe enricher with LRU cache.
// Note: Circuit Breaker and Retries are now handled by the gRPC client interceptors.
func NewPeerEnricherService(contacts *imcontact.Client) *PeerEnricher {
	// [MEMORY_MANAGEMENT] Bounded LRU cache to prevent OOM while maintaining hot data
	cache, _ := lru.New[string, model.Peer](10000)

	return &PeerEnricher{
		contacts: contacts,
		cache:    cache,
	}
}

// ResolvePeers orchestrates concurrent enrichment for message participants
// [CONCURRENCY_CONTROL] Executes parallel lookups via errgroup with context propagation
func (e *PeerEnricher) ResolvePeers(ctx context.Context, fromID string, toID uuid.UUID, domainID int32) (from, to model.Peer, err error) {
	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		var errFrom error
		from, errFrom = e.ResolvePeer(gCtx, fromID, domainID)
		return errFrom
	})

	g.Go(func() error {
		var errTo error
		to, errTo = e.ResolvePeer(gCtx, toID.String(), domainID)
		return errTo
	})

	err = g.Wait()
	return
}

// ResolvePeer executes a cache-aside enrichment flow.
func (e *PeerEnricher) ResolvePeer(ctx context.Context, id string, domainID int32) (model.Peer, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		// [DATA_INTEGRITY] Fallback for invalid UUIDs
		return model.NewPeer(uuid.Nil, model.PeerUser), nil
	}

	// [HOT_PATH] LRU Cache lookup (O(1))
	if peer, ok := e.cache.Get(id); ok {
		return peer, nil
	}

	// [NETWORK_CALL] Execution via gRPC client with built-in CB and Retry policy
	res, err := e.contacts.SearchContact(ctx, &contactv1.SearchContactRequest{
		Ids:      []string{id},
		DomainId: domainID,
		Size:     1,
		Page:     1,
	})
	if err != nil {
		// [GRACEFUL_FALLBACK] If the gRPC call fails (or CB is OPEN),
		// return a basic peer to prevent blocking the delivery flow.
		return model.NewPeer(uid, model.PeerUser), nil
	}

	// [NOT_FOUND_CHECK] Handle empty results as a successful but non-enriched peer
	contacts := res.GetContacts()
	if len(contacts) == 0 {
		peer := model.NewPeer(uid, model.PeerUser)
		e.cache.Add(id, peer)
		return peer, nil
	}

	contact := contacts[0]
	var name string
	if contact.GetName() != "" {
		name = contact.GetName()
	} else {
		name = contact.GetUsername()
	}

	// [SUCCESS] Build enriched domain entity and populate cache
	peer := model.NewPeer(uid, model.PeerUser,
		model.WithIdentity(
			contact.GetSubject(),
			contact.GetIssId(),
			name,
		),
	)

	e.cache.Add(id, peer)
	return peer, nil
}
