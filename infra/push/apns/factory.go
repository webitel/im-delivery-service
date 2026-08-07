package apns

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/sideshow/apns2"
	"github.com/sideshow/apns2/token"
	"golang.org/x/net/http2"
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
//
// proxy selects the APNs connection, mirroring webitel-portal semantics:
//   - ""                                   -> api.push.apple.com (production)
//   - https://api.sandbox.push.apple.com   -> Apple sandbox
//   - http[s]://host[:port][/path]         -> custom webitel-portal proxy
//
// For a custom proxy the native client still writes the /3/device/{token}
// path, the JWT and the payload body; only the host is substituted. proto
// picks the transport for a custom proxy: "h2"/"" (default) or "http/1.1".
func (r *clientRegistry) resolve(appID, proxy, proto string, p8Key []byte, keyID, teamID string) (*apns2.Client, error) {
	host := proxy
	if host == "" {
		host = apns2.HostProduction
	}

	// [KEY] Clients are pinned to a host+transport, so a switch must not reuse
	// a client bound to a different endpoint.
	key := appID + "|" + host + "|" + proto

	r.mu.RLock()
	client, ok := r.clients[key]
	r.mu.RUnlock()

	if ok {
		return client, nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if client, ok = r.clients[key]; ok {
		return client, nil
	}

	// [AUTH_TOKEN] Optional: a custom proxy impersonates APNs and may not need
	// a signing token; a direct Apple connection does.
	tok, err := newToken(p8Key, keyID, teamID)
	if err != nil {
		return nil, err
	}

	switch host {
	case apns2.HostProduction, apns2.HostDevelopment:
		// [APPLE] A direct connection requires a JWT signing token.
		if tok == nil {
			return nil, fmt.Errorf("apns: missing p8 key for app %s", appID)
		}

		client = apns2.NewTokenClient(tok)
	default:
		// [PROXY] Custom endpoint; substitute the host and adapt the transport.
		via, err := url.ParseRequestURI(host)
		if err != nil {
			return nil, fmt.Errorf("apns: invalid proxy url %q: %w", host, err)
		}

		// Sanitize [?query] and [#fragment].
		via.RawQuery = ""
		via.ForceQuery = false
		via.RawFragment = ""
		host = via.String()

		client = apns2.NewTokenClient(tok)

		if err := configureProxyTransport(client, via.Scheme, proto); err != nil {
			return nil, err
		}
	}

	// [ENDPOINT_OVERRIDE] The client appends /3/device/{token} to this host.
	client.Host = host

	r.clients[key] = client

	return client, nil
}

// newToken builds an APNs JWT signing token, or nil when no token material is
// supplied (valid only for a custom proxy).
func newToken(p8Key []byte, keyID, teamID string) (*token.Token, error) {
	if len(p8Key) == 0 && keyID == "" && teamID == "" {
		// No token material: a custom proxy may run unauthenticated.
		return nil, nil //nolint:nilnil // absence of a token is a valid state
	}

	if len(p8Key) == 0 {
		return nil, errors.New("apns: token auth_key required")
	}

	authKey, err := token.AuthKeyFromBytes(p8Key)
	if err != nil {
		return nil, fmt.Errorf("apns: key parse error: %w", err)
	}

	return &token.Token{
		AuthKey: authKey,
		KeyID:   keyID,  // e.g., "ABC123DEFG"
		TeamID:  teamID, // e.g., "DEF890GHIJ"
	}, nil
}

// configureProxyTransport adapts the apns2 HTTP/2 transport to reach a custom
// proxy. apns2 hardcodes an *http2.Transport that only speaks TLS h2, so a
// cleartext http:// proxy needs h2c, and an http/1.1 proxy needs a standard
// transport.
func configureProxyTransport(client *apns2.Client, scheme, proto string) error {
	secure := scheme == "https"

	switch scheme {
	case "http", "https":
	default:
		return fmt.Errorf("apns: proxy scheme %q not supported", scheme)
	}

	switch strings.ToLower(strings.TrimSpace(proto)) {
	case "h2", "http2", "http/2.0", "":
		// Over TLS the hardcoded http2.Transport already works as-is.
		if secure {
			return nil
		}

		// Cleartext h2 (h2c): let the http2.Transport dial plain TCP.
		h2t, ok := client.HTTPClient.Transport.(*http2.Transport)
		if !ok {
			return fmt.Errorf("apns: unexpected transport %T", client.HTTPClient.Transport)
		}

		dialer := &net.Dialer{Timeout: apns2.TLSDialTimeout, KeepAlive: apns2.TCPKeepAlive}

		h2t.AllowHTTP = true
		h2t.DialTLSContext = func(ctx context.Context, network, addr string, _ *tls.Config) (net.Conn, error) {
			return dialer.DialContext(ctx, network, addr)
		}
	case "http", "http/1.1":
		// Force HTTP/1.1 via a standard transport.
		client.HTTPClient.Transport = &http.Transport{
			DialContext: (&net.Dialer{Timeout: apns2.TLSDialTimeout, KeepAlive: apns2.TCPKeepAlive}).DialContext,
		}
	default:
		return fmt.Errorf("apns: proxy proto %q not supported", proto)
	}

	return nil
}
