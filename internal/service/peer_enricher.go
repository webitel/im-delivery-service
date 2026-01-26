package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	lru "github.com/hashicorp/golang-lru/v2"
	contactv1 "github.com/webitel/im-delivery-service/gen/go/contact/v1"
	imcontact "github.com/webitel/im-delivery-service/infra/client/im-contact"
	"github.com/webitel/im-delivery-service/internal/domain/model"
	"golang.org/x/sync/errgroup"
)

// Enricher defines the high-level contract for participant data augmentation.
type Enricher interface {
	// ResolvePeers performs concurrent enrichment for multiple participants.
	ResolvePeers(ctx context.Context, from, to model.Peer, domainID int32) (model.Peer, model.Peer, error)
	// ResolvePeer handles the logic for a single participant based on their type.
	ResolvePeer(ctx context.Context, peer model.Peer, domainID int32) (model.Peer, error)
}

type PeerEnricher struct {
	contacts *imcontact.Client
	cache    *lru.Cache[string, model.Peer]
}

// NewPeerEnricherService provides a thread-safe service with an internal LRU cache.
func NewPeerEnricherService(contacts *imcontact.Client) *PeerEnricher {
	// [MEMORY_MANAGEMENT] Pre-allocated LRU cache to minimize GC pressure and store "hot" identities.
	cache, _ := lru.New[string, model.Peer](10000)

	return &PeerEnricher{
		contacts: contacts,
		cache:    cache,
	}
}

// ResolvePeers executes parallel enrichment flows for 'from' and 'to' peers.
// [CONCURRENCY_OPTIMIZATION] Uses errgroup to ensure both lookups complete or fail together.
func (e *PeerEnricher) ResolvePeers(ctx context.Context, from, to model.Peer, domainID int32) (model.Peer, model.Peer, error) {
	g, gCtx := errgroup.WithContext(ctx)

	// Clone peers to avoid side effects during concurrent execution
	resFrom := from
	resTo := to

	g.Go(func() error {
		var err error
		resFrom, err = e.ResolvePeer(gCtx, from, domainID)
		return err
	})

	g.Go(func() error {
		var err error
		resTo, err = e.ResolvePeer(gCtx, to, domainID)
		return err
	})

	if err := g.Wait(); err != nil {
		return from, to, fmt.Errorf("parallel enrichment failed: %w", err)
	}

	return resFrom, resTo, nil
}

// ResolvePeer orchestrates the cache-aside strategy and polymorphic dispatching.
func (e *PeerEnricher) ResolvePeer(ctx context.Context, peer model.Peer, domainID int32) (model.Peer, error) {
	// [IDENTITY_GUARD] Ensure we have a valid ID before proceeding
	if peer.ID == uuid.Nil {
		return peer, nil
	}

	// [HOT_PATH] Check LRU cache first to avoid unnecessary network/logic overhead
	cacheKey := peer.ID.String()
	if cached, ok := e.cache.Get(cacheKey); ok {
		return cached, nil
	}

	var enriched model.Peer
	var err error

	// [POLYMORPHIC_DISPATCH] Route enrichment logic based on PeerType
	switch peer.Type {
	case model.PeerUser:
		// [EXTERNAL_GRPC_CALL] Fetch data from Contact Service
		enriched, err = e.enrichFromContacts(ctx, peer, domainID)

	case model.PeerGroup:
		// [STUB] Future logic for Chat Groups/Rooms metadata
		enriched = e.mockEnrich(peer, "Peer Group")

	case model.PeerChannel:
		// [STUB] Future logic for Broadcast Channels
		enriched = e.mockEnrich(peer, "Peer Channel")

	default:
		// [FALLBACK] Return original peer if type is unknown or doesn't require enrichment
		enriched = peer
	}

	// [CACHE_POPULATION] Save successful result (even if it's a fallback)
	if err == nil {
		e.cache.Add(cacheKey, enriched)
	}

	return enriched, err
}

// enrichFromContacts communicates with the gRPC Contact service.
func (e *PeerEnricher) enrichFromContacts(ctx context.Context, peer model.Peer, domainID int32) (model.Peer, error) {
	res, err := e.contacts.SearchContact(ctx, &contactv1.SearchContactRequest{
		Ids:      []string{peer.ID.String()},
		DomainId: domainID,
		Size:     1,
	})
	if err != nil {
		// [RESILIENCE] Graceful fallback: return original peer to keep the message moving
		return peer, nil
	}

	contacts := res.GetContacts()
	if len(contacts) == 0 {
		return peer, nil
	}

	contact := contacts[0]
	name := contact.GetName()
	if name == "" {
		name = contact.GetUsername()
	}

	// [SUCCESS] Populate peer with identity data
	peer.Name = name
	peer.Sub = contact.GetSubject()
	peer.Issuer = contact.GetIssId()

	return peer, nil
}

// mockEnrich is a helper for types not yet fully implemented.
func (e *PeerEnricher) mockEnrich(peer model.Peer, placeholder string) model.Peer {
	if peer.Name == "" {
		peer.Name = fmt.Sprintf("%s (%s)", placeholder, peer.ID.String()[:8])
	}
	return peer
}
