package fcm

import (
	"context"
	"fmt"
	"sync"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

// [CLIENT_REGISTRY] Encapsulates thread-safe Firebase client management.
type clientRegistry struct {
	mu      sync.RWMutex
	clients map[string]*messaging.Client
}

func newClientRegistry() *clientRegistry {
	return &clientRegistry{
		clients: make(map[string]*messaging.Client),
	}
}

// [RESOLVE] Returns an existing messaging client or initializes a new one.
func (r *clientRegistry) resolve(ctx context.Context, appID string, creds []byte) (*messaging.Client, error) {
	// [1. FAST_PATH] Read-lock for existing clients.
	r.mu.RLock()
	client, ok := r.clients[appID]
	r.mu.RUnlock()
	if ok {
		return client, nil
	}

	// [2. SLOW_PATH] Write-lock for initialization.
	r.mu.Lock()
	defer r.mu.Unlock()

	// Double-check pattern.
	if client, ok = r.clients[appID]; ok {
		return client, nil
	}

	if len(creds) == 0 {
		return nil, fmt.Errorf("fcm: missing credentials for app %s", appID)
	}

	// [3. INIT] Firebase App & Messaging Client.
	app, err := firebase.NewApp(ctx, nil, option.WithCredentialsJSON(creds))
	if err != nil {
		return nil, fmt.Errorf("fcm: firebase init failed: %w", err)
	}

	newClient, err := app.Messaging(ctx)
	if err != nil {
		return nil, fmt.Errorf("fcm: messaging init failed: %w", err)
	}

	r.clients[appID] = newClient
	return newClient, nil
}
