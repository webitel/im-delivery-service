package fcm

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"sync"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	fcmv1 "google.golang.org/api/fcm/v1"
	"google.golang.org/api/option"
	"google.golang.org/api/transport"
)

// fakeProjectID is the constant project id used when proxying without a real
// service account, mirroring webitel-portal (legacy "invalid").
const fakeProjectID = "proxy"

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
//
// A non-empty proxy substitutes the FCM v1 endpoint: the native SDK still
// builds the /projects/{id}/messages:send path and the {"message":...} body,
// but ships them to the proxy instead of fcm.googleapis.com — a
// webitel-portal-compatible push proxy. Mirroring the portal, proxying does
// NOT mint a Google OAuth token (WithoutAuthentication) and does not require a
// service account: without credentials the project id is faked to "proxy".
func (r *clientRegistry) resolve(ctx context.Context, appID, proxy string, creds []byte) (*messaging.Client, error) {
	// [KEY] Clients are bound to an endpoint, so a proxy switch must not reuse
	// a client pinned to a different base URL.
	key := appID + "|" + proxy

	// [1. FAST_PATH] Read-lock for existing clients.
	r.mu.RLock()
	client, ok := r.clients[key]
	r.mu.RUnlock()

	if ok {
		return client, nil
	}

	// [2. SLOW_PATH] Write-lock for initialization.
	r.mu.Lock()
	defer r.mu.Unlock()

	// Double-check pattern.
	if client, ok = r.clients[key]; ok {
		return client, nil
	}

	// [3. VALIDATE] A direct (non-proxied) client needs a service account.
	if proxy == "" && len(creds) == 0 {
		return nil, fmt.Errorf("fcm: missing credentials for app %s", appID)
	}

	// [4. OPTIONS] Scopes + optional service account.
	opts := []option.ClientOption{
		option.WithScopes(fcmv1.FirebaseMessagingScope),
	}
	if len(creds) > 0 {
		opts = append(opts, option.WithCredentialsJSON(creds))
	}

	// [5. CONFIG] Empty config prevents default ENV credential lookup.
	cfg := &firebase.Config{}

	// [6. PROXY] Substitute the endpoint and drop Google authentication.
	if proxy != "" {
		endpoint, err := normalizeEndpoint(proxy)
		if err != nil {
			return nil, fmt.Errorf("fcm: app %s: %w", appID, err)
		}

		// Resolve the project id for the request path: from the service
		// account when present, otherwise a constant the proxy accepts.
		cfg.ProjectID = fakeProjectID
		if resolved, err := transport.Creds(ctx, opts...); err == nil && resolved.ProjectID != "" {
			cfg.ProjectID = resolved.ProjectID
		}

		opts = append(opts,
			option.WithEndpoint(endpoint),
			option.WithoutAuthentication(),
			option.WithTelemetryDisabled(),
		)
	}

	// [7. INIT] Firebase App & Messaging Client.
	app, err := firebase.NewApp(ctx, cfg, opts...)
	if err != nil {
		return nil, fmt.Errorf("fcm: firebase init failed: %w", err)
	}

	newClient, err := app.Messaging(ctx)
	if err != nil {
		return nil, fmt.Errorf("fcm: messaging init failed: %w", err)
	}

	r.clients[key] = newClient

	return newClient, nil
}

// normalizeEndpoint validates a proxy URL and trims a trailing slash so the
// SDK appends /projects/{id}/messages:send cleanly.
func normalizeEndpoint(raw string) (string, error) {
	via, err := url.ParseRequestURI(raw)
	if err != nil {
		return "", fmt.Errorf("invalid proxy url %q: %w", raw, err)
	}

	if !via.IsAbs() {
		return "", fmt.Errorf("proxy url %q must be absolute", raw)
	}

	switch via.Scheme {
	case "http", "https":
	default:
		return "", fmt.Errorf("proxy url %q: expect http[s] scheme", raw)
	}

	via.Path = strings.TrimRight(via.Path, "/")

	return via.String(), nil
}
