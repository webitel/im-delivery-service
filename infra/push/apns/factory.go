package apns

import (
	"fmt"
	"sync"

	"github.com/sideshow/apns2"
	"github.com/sideshow/apns2/token"
)

// [CLIENT_REGISTRY] Manages long-lived HTTP/2 connections to Apple APNS.
type clientRegistry struct {
	mu      sync.RWMutex
	clients map[string]*apns2.Client
}

func newClientRegistry() *clientRegistry {
	return &clientRegistry{
		clients: make(map[string]*apns2.Client),
	}
}

// [RESOLVE] Returns an existing APNS client or initializes a new token-based one.
func (r *clientRegistry) resolve(appID string, p8Key []byte, keyID, teamID string) (*apns2.Client, error) {
	r.mu.RLock()
	client, ok := r.clients[appID]
	r.mu.RUnlock()
	if ok {
		return client, nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if client, ok = r.clients[appID]; ok {
		return client, nil
	}

	if len(p8Key) == 0 {
		return nil, fmt.Errorf("apns: missing p8 key for app %s", appID)
	}

	// [AUTH_TOKEN] Apple recommends using JWT tokens (.p8) over certificates.
	authKey, err := token.AuthKeyFromBytes(p8Key)
	if err != nil {
		return nil, fmt.Errorf("apns: key parse error: %w", err)
	}

	t := &token.Token{
		AuthKey: authKey,
		KeyID:   keyID,  // e.g., "ABC123DEFG"
		TeamID:  teamID, // e.g., "DEF890GHIJ"
	}

	// [CLIENT_INIT] Production client uses HTTP/2.
	newClient := apns2.NewTokenClient(t).Production()

	r.clients[appID] = newClient
	return newClient, nil
}
