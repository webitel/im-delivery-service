package apns

import (
	"context"
	"log/slog"

	"github.com/sideshow/apns2"
	"github.com/sideshow/apns2/payload"
	"github.com/webitel/im-delivery-service/infra/push/webhook"
	"github.com/webitel/im-delivery-service/internal/domain/model"
	"github.com/webitel/webitel-go-kit/pkg/semconv"
)

const Name = "apn"

// [PROVIDER] Implements push.Provider for Apple devices.
type Provider struct {
	log      *slog.Logger
	registry *clientRegistry
}

func NewProvider(log *slog.Logger) *Provider {
	return &Provider{
		log:      log.With(semconv.ComponentKey, "push.apns"),
		registry: newClientRegistry(),
	}
}

func (p *Provider) Name() string { return Name }

func (p *Provider) Send(ctx context.Context, req *model.PushRequest) error {
	return p.dispatch(ctx, req, false)
}

func (p *Provider) Dismiss(ctx context.Context, req *model.PushRequest) error {
	return p.dispatch(ctx, req, true)
}

// [PROCESS] Orchestrates the delivery or proxying of the notification.
func (p *Provider) dispatch(ctx context.Context, req *model.PushRequest, isDismiss bool) error {
	for _, dev := range req.Devices {
		if dev.PushType != Name {
			continue
		}

		// [1. BUILD_NATIVE]
		notification := p.buildAPNSNotification(dev, req, isDismiss)

		// [2. DEBUG_PROXY]
		if dev.PushConfig.Proxy != "" {
			p.log.Debug("PROXY_REDIRECT", slog.String("url", dev.PushConfig.Proxy))
			proxy := webhook.GetOrCreate[*apns2.Notification](dev.PushConfig.Proxy)
			if err := proxy.Send(ctx, notification); err != nil {
				p.log.Error("PROXY_ERROR", slog.Any(semconv.ErrorKey, err))
			}
			continue
		}

		// [3. RESOLVE_TRANSPORT]
		client, err := p.registry.resolve(
			dev.AppID,
			dev.PushConfig.Credentials,
			dev.PushConfig.KeyID, // Assuming these exist in your Config model
			dev.PushConfig.TeamID,
		)
		if err != nil {
			p.log.Error("CLIENT_RESOLUTION_FAILED", slog.Any(semconv.ErrorKey, err))
			continue
		}

		// [4. DISPATCH]
		res, err := client.PushWithContext(ctx, notification)
		if err != nil {
			p.log.Error("TRANSPORT_ERROR", slog.Any(semconv.ErrorKey, err))
			continue
		}

		if !res.Sent() {
			p.log.Warn("APNS_REJECTED",
				slog.Int("status", res.StatusCode),
				slog.String("reason", res.Reason),
				slog.String("id", res.ApnsID),
			)
		} else {
			p.log.Info("PUSH_DELIVERED", slog.String("id", res.ApnsID))
		}
	}
	return nil
}

// [BUILDER] Prepares the final APNS payload.
func (p *Provider) buildAPNSNotification(dev model.Device, req *model.PushRequest, isDismiss bool) *apns2.Notification {
	pl := payload.NewPayload()

	n := &apns2.Notification{
		DeviceToken: dev.PushToken,
		Topic:       dev.PushConfig.Topic, // Bundle ID
		CollapseID:  req.CollapseID,
	}

	// [LOGIC] Visual Alert vs Background/Silent (Dismiss)
	if isDismiss {
		// [DISMISS] Apple handles this via 'content-available: 1' and background push type.
		pl.ContentAvailable()
		pl.Custom("action", "CANCEL")
		pl.Custom("collapse_id", req.CollapseID)

		n.Priority = apns2.PriorityLow
		n.PushType = apns2.PushTypeBackground
	} else {
		pl.AlertTitle(req.Title)
		pl.AlertBody(req.Body)
		pl.Sound("default")
		pl.MutableContent()

		n.Priority = apns2.PriorityHigh
		n.PushType = apns2.PushTypeAlert
	}

	// [DATA] Merge domain-specific data.
	for k, v := range req.Data {
		pl.Custom(k, v)
	}

	n.Payload = pl
	return n
}
