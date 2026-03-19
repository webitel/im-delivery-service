package apns

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/sideshow/apns2"
	"github.com/sideshow/apns2/payload"
	"github.com/sideshow/apns2/token"
	"github.com/webitel/im-delivery-service/infra/push/webhook"
	"github.com/webitel/im-delivery-service/internal/domain/model"
)

const Name = "apn"

// [APNS_PROVIDER] Stateless provider using per-device tokens and topics.
type apnsProvider struct {
	log     *slog.Logger
	mu      sync.RWMutex
	clients map[string]*apns2.Client
}

func NewAPNSProvider(log *slog.Logger) *apnsProvider {
	return &apnsProvider{
		log:     log.With("provider", Name),
		clients: make(map[string]*apns2.Client),
	}
}

func (p *apnsProvider) Name() string { return Name }

func (p *apnsProvider) Send(ctx context.Context, req *model.PushRequest) error {
	for _, dev := range req.Devices {
		if dev.PushType != Name {
			continue
		}

		// 1. [PROXY] Webhook delegation.
		if dev.PushConfig.Proxy != "" {
			p.log.Debug("PROXY_DELEGATION", slog.String("url", dev.PushConfig.Proxy))
			proxy := webhook.GetOrCreate(dev.PushConfig.Proxy)
			_ = proxy.Send(ctx, req)
			continue
		}

		// 2. [CLIENT] Get or create HTTP/2 client with p8 token.
		client, err := p.getOrCreateClient(dev.AppID, dev.PushConfig.Credentials)
		if err != nil {
			p.log.Error("CLIENT_INIT_FAILED", slog.String("app", dev.AppID), slog.Any("err", err))
			continue
		}

		// 3. [NATIVE_SEND] Dispatch with dynamic Topic (Bundle ID).
		pl := payload.NewPayload().
			AlertTitle(req.Title).
			AlertBody(req.Body).
			Sound("default").
			MutableContent().
			Custom("event_id", req.CollapseID)

		notification := &apns2.Notification{
			DeviceToken: dev.PushToken,
			Topic:       dev.PushConfig.Topic, // IMPORTANT: Dynamic bundle ID from Admin service.
			Payload:     pl,
			CollapseID:  req.CollapseID,
		}

		res, err := client.PushWithContext(ctx, notification)
		if err != nil {
			p.log.Error("APNS_TRANSPORT_ERROR", slog.Any("err", err))
			continue
		}

		if !res.Sent() {
			p.log.Warn("APNS_REJECTED",
				slog.Int("status", res.StatusCode),
				slog.String("reason", res.Reason),
			)
		}
	}
	return nil
}

func (p *apnsProvider) getOrCreateClient(appID string, p8Key []byte) (*apns2.Client, error) {
	if len(p8Key) == 0 {
		return nil, fmt.Errorf("missing p8 key for app: %s", appID)
	}

	p.mu.RLock()
	client, ok := p.clients[appID]
	p.mu.RUnlock()
	if ok {
		return client, nil
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if client, ok = p.clients[appID]; ok {
		return client, nil
	}

	// [AUTH_KEY] Parse the p8 bytes.
	// Note: In real setup, you'd also need KeyID and TeamID from PushConfig.
	authKey, err := token.AuthKeyFromBytes(p8Key)
	if err != nil {
		return nil, fmt.Errorf("apns key parse failed: %w", err)
	}

	t := &token.Token{
		AuthKey: authKey,
		// These should ideally be part of your Device.PushConfig too.
		KeyID:  "DYNAMIC_KEY_ID",
		TeamID: "DYNAMIC_TEAM_ID",
	}

	newClient := apns2.NewTokenClient(t).Production()
	p.clients[appID] = newClient

	p.log.Info("NEW_APNS_CLIENT_CREATED", slog.String("app_id", appID))
	return newClient, nil
}

func (p *apnsProvider) Dismiss(ctx context.Context, req *model.PushRequest) error {
	return p.Send(ctx, req)
}
