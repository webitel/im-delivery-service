package cmd

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"errors"
	"io"
	"log/slog"
	"math/big"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/fx"

	healthfx "github.com/webitel/webitel-go-kit/infra/health/fx"

	"github.com/webitel/im-delivery-service/config"
	grpcsrv "github.com/webitel/im-delivery-service/infra/server/grpc"
	httpsrv "github.com/webitel/im-delivery-service/infra/server/http"
	infratls "github.com/webitel/im-delivery-service/infra/tls"
	"github.com/webitel/im-delivery-service/internal/domain/model"
)

type rejectAll struct{}

func (rejectAll) Inspect(context.Context) (*model.AuthContact, error) {
	return nil, errors.New("rejected")
}

func TestHealth_Lifecycle(t *testing.T) {
	cfg := &config.Config{}
	cfg.Service.HTTPAddr = "127.0.0.1:0"

	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	grpcServer, err := grpcsrv.New("127.0.0.1:0", log, rejectAll{}, &infratls.Config{Server: selfSignedTLS(t)}, nil)
	if err != nil {
		t.Fatalf("new grpc server: %v", err)
	}

	rdb := redis.NewClient(&redis.Options{Addr: "127.0.0.1:1"})

	t.Cleanup(func() {
		if err := rdb.Close(); err != nil {
			t.Errorf("close redis: %v", err)
		}
	})

	var httpServer *httpsrv.Server

	app := fx.New(
		fx.NopLogger,
		fx.Supply(cfg, log, grpcServer, rdb),
		fx.Provide(http.NewServeMux, httpsrv.New),
		healthfx.Module(healthfx.Config{}),
		fx.Invoke(func(lc fx.Lifecycle) {
			lc.Append(fx.Hook{
				OnStart: func(context.Context) error {
					go func() {
						if err := grpcServer.Listen(); err != nil && !errors.Is(err, net.ErrClosed) {
							t.Errorf("grpc listen: %v", err)
						}
					}()

					return nil
				},
				OnStop: func(context.Context) error { return grpcServer.Shutdown() },
			})
		}),
		fx.Invoke(registerHealth),
		healthfx.Shutdown(),
		fx.Populate(&httpServer),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := app.Start(ctx); err != nil {
		t.Fatalf("start app: %v", err)
	}

	waitReadyz(t, "http://"+httpServer.Listener().Addr().String()+"/readyz")

	if err := app.Stop(context.Background()); err != nil {
		t.Fatalf("stop app: %v", err)
	}
}

func waitReadyz(t *testing.T, url string) {
	t.Helper()

	deadline := time.Now().Add(10 * time.Second)

	for {
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
		if err != nil {
			t.Fatalf("new request: %v", err)
		}

		resp, err := http.DefaultClient.Do(req)
		if err == nil {
			if cerr := resp.Body.Close(); cerr != nil {
				t.Errorf("close body: %v", cerr)
			}

			if resp.StatusCode == http.StatusOK {
				return
			}
		}

		if time.Now().After(deadline) {
			t.Fatalf("GET %s never returned 200: last err %v", url, err)
		}

		time.Sleep(50 * time.Millisecond)
	}
}

func selfSignedTLS(t *testing.T) *tls.Config {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "localhost"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1")},
	}

	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("create certificate: %v", err)
	}

	return &tls.Config{
		Certificates: []tls.Certificate{{Certificate: [][]byte{der}, PrivateKey: key}},
		MinVersion:   tls.VersionTLS12,
	}
}
