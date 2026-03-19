package fcm

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"github.com/webitel/im-delivery-service/infra/push/webhook"
	"github.com/webitel/im-delivery-service/internal/domain/model"
	"google.golang.org/api/option"
)

const Name = "fcm"

// [FCM_PROVIDER] Manages multiple Firebase projects dynamically.
type fcmProvider struct {
	log     *slog.Logger
	mu      sync.RWMutex
	clients map[string]*messaging.Client // Key: AppName or Credentials Hash
}

func NewFCMProvider(log *slog.Logger) *fcmProvider {
	return &fcmProvider{
		log:     log.With("provider", Name),
		clients: make(map[string]*messaging.Client),
	}
}

func (p *fcmProvider) Name() string { return Name }

// [SEND] Iterates through devices and uses specific credentials for each.
func (p *fcmProvider) Send(ctx context.Context, req *model.PushRequest) error {
	for _, dev := range req.Devices {
		if dev.PushType != Name {
			continue
		}

		// 1. [PROXY] Check for webhook delegation.
		if dev.PushConfig.Proxy != "" {
			p.log.Debug("PROXY_DELEGATION", slog.String("url", dev.PushConfig.Proxy))
			proxy := webhook.GetOrCreate(dev.PushConfig.Proxy)
			if err := proxy.Send(ctx, req); err != nil {
				p.log.Error("PROXY_FAILED", slog.Any("err", err))
			}
			continue
		}

		// 2. [CLIENT] Get or create a specific FCM client for this app's credentials.
		client, err := p.getOrCreateClient(ctx, dev.AppID, dev.PushConfig.Credentials)
		if err != nil {
			p.log.Error("CLIENT_INIT_FAILED", slog.String("app", dev.AppID), slog.Any("err", err))
			continue
		}

		// 3. [NATIVE_SEND] Single message dispatch.
		msg := &messaging.Message{
			Token: dev.PushToken,
			Data:  req.Data,
			Notification: &messaging.Notification{
				Title: req.Title,
				Body:  req.Body,
			},
			Android: &messaging.AndroidConfig{
				CollapseKey: req.CollapseID,
				Priority:    "high",
				Notification: &messaging.AndroidNotification{
					Tag:   req.CollapseID,
					Sound: "default",
				},
			},
		}

		_, err = client.Send(ctx, msg)
		if err != nil {
			p.log.Error("FCM_DISPATCH_FAILED", slog.String("token", dev.PushToken), slog.Any("err", err))
		}
	}
	return nil
}

func (p *fcmProvider) getOrCreateClient(ctx context.Context, appID string, creds []byte) (*messaging.Client, error) {
	if len(creds) == 0 {
		return nil, fmt.Errorf("empty credentials for app: %s", appID)
	}

	p.mu.RLock()
	client, ok := p.clients[appID]
	p.mu.RUnlock()
	if ok {
		return client, nil
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	// Double check pattern.
	if client, ok = p.clients[appID]; ok {
		return client, nil
	}

	// [INIT] Initialize Firebase app from raw JSON bytes.
	app, err := firebase.NewApp(ctx, nil, option.WithCredentialsJSON(creds))
	if err != nil {
		return nil, fmt.Errorf("firebase app init failed: %w", err)
	}

	newClient, err := app.Messaging(ctx)
	if err != nil {
		return nil, fmt.Errorf("messaging client init failed: %w", err)
	}

	p.clients[appID] = newClient
	p.log.Info("NEW_FCM_CLIENT_CREATED", slog.String("app_id", appID))
	return newClient, nil
}

func (p *fcmProvider) Dismiss(ctx context.Context, req *model.PushRequest) error {
	return p.Send(ctx, req)
}
