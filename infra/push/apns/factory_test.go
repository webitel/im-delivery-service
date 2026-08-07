package apns

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"testing"

	"github.com/sideshow/apns2"
	"golang.org/x/net/http2"
)

// testP8Key returns a valid PKCS8-encoded EC private key, matching the .p8
// format APNs token auth expects. Generated per-run so the test stays hermetic.
func testP8Key(t *testing.T) []byte {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatalf("marshal key: %v", err)
	}

	return pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})
}

func TestResolveHostSelection(t *testing.T) {
	p8 := testP8Key(t)

	tests := []struct {
		name     string
		proxy    string
		wantHost string
	}{
		{
			name:     "empty proxy defaults to production",
			proxy:    "",
			wantHost: apns2.HostProduction,
		},
		{
			name:     "sandbox host",
			proxy:    apns2.HostDevelopment,
			wantHost: apns2.HostDevelopment,
		},
		{
			name:     "custom https proxy",
			proxy:    "https://proxy.example.com/push/apns",
			wantHost: "https://proxy.example.com/push/apns",
		},
		{
			name:     "custom proxy query and fragment are stripped",
			proxy:    "https://proxy.example.com/push/apns?x=1#frag",
			wantHost: "https://proxy.example.com/push/apns",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := newClientRegistry()

			client, err := r.resolve("app-1", tc.proxy, "", p8, "KEY1234567", "TEAM123456")
			if err != nil {
				t.Fatalf("resolve: %v", err)
			}

			if client.Host != tc.wantHost {
				t.Fatalf("host = %q, want %q", client.Host, tc.wantHost)
			}
		})
	}
}

// A direct Apple connection requires a signing token; a custom proxy does not.
func TestResolveTokenRequirement(t *testing.T) {
	r := newClientRegistry()

	if _, err := r.resolve("app-1", "", "", nil, "", ""); err == nil {
		t.Fatal("expected error for direct Apple connection without a token")
	}

	client, err := r.resolve("app-2", "http://10.10.10.4:8043/push/apns", "", nil, "", "")
	if err != nil {
		t.Fatalf("custom proxy without token must be allowed: %v", err)
	}

	if client.Host != "http://10.10.10.4:8043/push/apns" {
		t.Fatalf("host = %q", client.Host)
	}
}

// A cleartext http proxy must switch the hardcoded http2.Transport to h2c.
func TestResolveClearTextProxyEnablesH2C(t *testing.T) {
	r := newClientRegistry()

	client, err := r.resolve("app-1", "http://10.10.10.4:8043/push/apns", "h2", nil, "", "")
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}

	h2t, ok := client.HTTPClient.Transport.(*http2.Transport)
	if !ok {
		t.Fatalf("transport = %T, want *http2.Transport", client.HTTPClient.Transport)
	}

	if !h2t.AllowHTTP || h2t.DialTLSContext == nil {
		t.Fatal("expected h2c transport (AllowHTTP + DialTLSContext) for cleartext proxy")
	}
}

func TestResolveInvalidProxy(t *testing.T) {
	r := newClientRegistry()

	if _, err := r.resolve("app-1", "ftp://nope", "", nil, "", ""); err == nil {
		t.Fatal("expected error for unsupported proxy scheme")
	}

	if _, err := r.resolve("app-1", "http://ok/apns", "smtp", nil, "", ""); err == nil {
		t.Fatal("expected error for unsupported proto")
	}
}

func TestResolveCachesPerEndpoint(t *testing.T) {
	p8 := testP8Key(t)
	r := newClientRegistry()

	direct, err := r.resolve("app-1", "", "", p8, "KEY1234567", "TEAM123456")
	if err != nil {
		t.Fatalf("resolve direct: %v", err)
	}

	directAgain, err := r.resolve("app-1", "", "", p8, "KEY1234567", "TEAM123456")
	if err != nil {
		t.Fatalf("resolve direct again: %v", err)
	}

	if direct != directAgain {
		t.Fatal("expected cached client for identical endpoint")
	}

	proxied, err := r.resolve("app-1", "https://proxy.example.com", "", p8, "KEY1234567", "TEAM123456")
	if err != nil {
		t.Fatalf("resolve proxied: %v", err)
	}

	if proxied == direct {
		t.Fatal("proxy endpoint must not reuse the direct client")
	}
}
