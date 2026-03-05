package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	lru "github.com/hashicorp/golang-lru/v2"
	contactv1 "github.com/webitel/im-delivery-service/gen/go/contact/v1"
	imcontact "github.com/webitel/im-delivery-service/infra/client/im-contact"
	"github.com/webitel/im-delivery-service/internal/domain/model"
)

type Contacter interface {
	Resolve(ctx context.Context, domainID int32, ids ...uuid.UUID) ([]model.Peer, error)
}

type ContactEnricher struct {
	client *imcontact.Client
	cache  *lru.Cache[uuid.UUID, model.Peer]
	logger *slog.Logger
}

func NewContactEnricher(client *imcontact.Client, logger *slog.Logger) *ContactEnricher {
	cache, _ := lru.New[uuid.UUID, model.Peer](10000)
	return &ContactEnricher{
		client: client,
		cache:  cache,
		logger: logger,
	}
}

func (e *ContactEnricher) Resolve(ctx context.Context, domainID int32, ids ...uuid.UUID) ([]model.Peer, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	res := make([]model.Peer, len(ids))
	missing := make(map[uuid.UUID][]int)

	// 1. [CACHE_LOOKUP] Try to fill as much as possible from local LRU
	for i, id := range ids {
		if id == uuid.Nil {
			continue
		}

		if cached, ok := e.cache.Get(id); ok {
			res[i] = cached
		} else {
			missing[id] = append(missing[id], i)
		}
	}

	// [HOT_PATH] All identities found in cache
	if len(missing) == 0 {
		return res, nil
	}

	// 2. [REMOTE_FETCH] Request missing data from the contact service
	e.fetch(ctx, domainID, missing, res)

	// 3. [FALLBACK] Ensure no empty slots remain for failed/missing contacts
	for id, indices := range missing {
		unknown := model.Peer{
			ID:   id,
			Type: model.PeerUser,
			Name: fmt.Sprintf("Unknown (%s)", id.String()[:8]),
		}
		for _, idx := range indices {
			res[idx] = unknown
		}
	}

	return res, nil
}

func (e *ContactEnricher) fetch(ctx context.Context, domainID int32, missing map[uuid.UUID][]int, res []model.Peer) {
	searchIDs := make([]string, 0, len(missing))
	for id := range missing {
		searchIDs = append(searchIDs, id.String())
	}

	resp, err := e.client.SearchContact(ctx, &contactv1.SearchContactRequest{
		Ids:      searchIDs,
		DomainId: domainID,
		Size:     int32(len(searchIDs)),
	})
	if err != nil {
		e.logger.Error("CONTACT_FETCH_FAILED", "err", err)
		return
	}

	// [MAP_RESULTS] Link back to original slice positions and update cache
	for _, c := range resp.GetContacts() {
		id, err := uuid.Parse(c.GetId())
		if err != nil {
			continue
		}

		indices, found := missing[id]
		if !found {
			continue
		}

		peer := e.toPeer(c)
		for _, idx := range indices {
			res[idx] = peer
		}

		e.cache.Add(id, peer)
		delete(missing, id)
	}
}

func (e *ContactEnricher) toPeer(c *contactv1.Contact) model.Peer {
	name := c.GetName()
	if name == "" {
		name = c.GetUsername()
	}

	return model.Peer{
		ID:     uuid.MustParse(c.GetId()),
		Type:   e.parseType(c.GetType()),
		Name:   name,
		Sub:    c.GetSubject(),
		Issuer: c.GetIssId(),
	}
}

func (e *ContactEnricher) parseType(t string) model.PeerType {
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
