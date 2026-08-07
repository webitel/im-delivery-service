package apns

import (
	"context"
	"log/slog"

	"github.com/sideshow/apns2"
	"github.com/sideshow/apns2/payload"

	"github.com/webitel/im-delivery-service/internal/domain/model"
)

const Name = "apn"

// [PROVIDER] Implements push.Provider for Apple devices.
type Provider struct {
	log      *slog.Logger
	registry *clientRegistry
}

func NewProvider(log *slog.Logger) *Provider {
	return &Provider{
		log:      log.With("component", "push.apns"),
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

		// [2. RESOLVE_TRANSPORT]
		// A configured proxy substitutes the APNs host, so the native client
		// still writes the /3/device/{token} path, JWT and body — the proxy
		// receives a request identical to what api.push.apple.com would.
		client, err := p.registry.resolve(
			dev.AppID,
			dev.PushConfig.Proxy,
			dev.PushConfig.Proto,
			dev.PushConfig.Credentials,
			dev.PushConfig.KeyID,
			dev.PushConfig.TeamID,
		)
		if err != nil {
			p.log.Error("CLIENT_RESOLUTION_FAILED", slog.Any("error", err))

			continue
		}

		// [3. DISPATCH]
		res, err := client.PushWithContext(ctx, notification)
		if err != nil {
			p.log.Error("TRANSPORT_ERROR", slog.Any("error", err))

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
