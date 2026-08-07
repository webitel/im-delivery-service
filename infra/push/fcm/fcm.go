package fcm

import (
	"context"
	"log/slog"
	"time"

	"firebase.google.com/go/v4/messaging"

	"github.com/webitel/im-delivery-service/infra/push/webhook"
	"github.com/webitel/im-delivery-service/internal/domain/model"
)

const (
	Name           = "fcm"
	DefaultChannel = "im_messages"
	// [RFC_INTENT] Custom action for Android Intent Filter orchestration.
	// https://developer.android.com/guide/components/intents-filters
	DefaultAction  = "OPEN_CHAT_ACTION"
	AnalyticsLabel = "webitel_delivery_v1"
)

// [PROVIDER] High-performance Firebase Cloud Messaging driver.
// Implements the push.Provider interface with support for Native Android & iOS (via FCM).
type Provider struct {
	log      *slog.Logger
	registry *clientRegistry
}

func NewProvider(log *slog.Logger) *Provider {
	return &Provider{
		log:      log.With("component", "push.fcm"),
		registry: newClientRegistry(),
	}
}

func (p *Provider) Name() string { return Name }

// [SEND] Dispatches standard visual alerts for new incoming messages.
func (p *Provider) Send(ctx context.Context, req *model.PushRequest) error {
	return p.dispatch(ctx, req, false)
}

// [DISMISS] Handles remote notification cancellation via Silent Push / Data messages.
func (p *Provider) Dismiss(ctx context.Context, req *model.PushRequest) error {
	return p.dispatch(ctx, req, true)
}

// [PROCESS] Main orchestration loop for building and dispatching messages.
func (p *Provider) dispatch(ctx context.Context, req *model.PushRequest, isDismiss bool) error {
	for _, dev := range req.Devices {
		if dev.PushType != Name {
			continue
		}

		// [1. PAYLOAD_CONSTRUCTION]
		msg := p.buildFCMMessage(dev, req, isDismiss)

		// [2. PROXY_DELEGATION]
		// If Proxy URL is defined, bypass direct Firebase dispatch.
		if dev.PushConfig.Proxy != "" {
			p.log.Debug("DELEGATING_TO_PROXY", slog.String("url", dev.PushConfig.Proxy))

			proxy := webhook.GetOrCreate[*messaging.Message](dev.PushConfig.Proxy)
			if err := proxy.Send(ctx, msg); err != nil {
				p.log.Error("PROXY_DISPATCH_FAILED", slog.Any("error", err))
			}

			continue
		}

		// [3. CLIENT_RESOLUTION]
		// Resolves an authenticated messaging client from the registry.
		client, err := p.registry.resolve(ctx, dev.AppID, dev.PushConfig.Credentials)
		if err != nil {
			p.log.Error("CLIENT_RESOLUTION_FAILED",
				slog.String("app_id", dev.AppID),
				slog.Any("error", err))

			continue
		}

		// [4. TRANSPORT_EXECUTION]
		// Synchronous dispatch to FCM v1 HTTP API.
		// https://firebase.google.com/docs/reference/fcm/rest/v1/projects.messages/send
		msgID, err := client.Send(ctx, msg)
		if err != nil {
			p.log.Error("FCM_DISPATCH_FAILED",
				slog.String("token", dev.PushToken),
				slog.Any("error", err))

			continue
		}

		p.log.Info("PUSH_SENT", slog.String("msg_id", msgID), slog.String("app", dev.AppID))
	}

	return nil
}

// [BUILDER] Constructs the multi-platform FCM payload.
func (p *Provider) buildFCMMessage(dev model.Device, req *model.PushRequest, isDismiss bool) *messaging.Message {
	ttl := 24 * time.Hour

	m := &messaging.Message{
		Token: dev.PushToken,
		Data:  req.Data,
		// [OBSERVABILITY] Label for FCM aggregate delivery reports.
		FCMOptions: &messaging.FCMOptions{
			AnalyticsLabel: AnalyticsLabel,
		},
	}

	// [PAYLOAD_TYPE] Switch between Notification Alert and Data-only (Silent).
	if isDismiss {
		if m.Data == nil {
			m.Data = make(map[string]string)
		}

		m.Data["action"] = "CANCEL"
		m.Data["collapse_id"] = req.CollapseID
	} else {
		m.Notification = &messaging.Notification{
			Title: req.Title,
			Body:  req.Body,
		}
	}

	// [ANDROID_SPECIFIC]
	// Optimizations for Android OS power management and notification tray behavior.
	m.Android = &messaging.AndroidConfig{
		CollapseKey: req.CollapseID, // [RFC] Replaces existing messages with same key.
		Priority:    "high",         // Forces immediate wake-up via FCM high-priority.
		TTL:         &ttl,
		Notification: &messaging.AndroidNotification{
			Tag:         req.CollapseID,
			Sound:       "default",
			ChannelID:   DefaultChannel,
			Icon:        "ic_notification",
			Color:       "#0052cc",
			ClickAction: DefaultAction,
			// [PRIVACY] Hides sensitive IM content from public lock screens.
			Visibility: messaging.VisibilityPrivate,
		},
	}

	// [APNS_SPECIFIC]
	// Configuration for Apple devices registered through Firebase.
	m.APNS = &messaging.APNSConfig{
		Payload: &messaging.APNSPayload{
			Aps: &messaging.Aps{
				Alert: &messaging.ApsAlert{
					Title: req.Title,
					Body:  req.Body,
				},
				MutableContent: true, // Enables Notification Service Extensions.
				Sound:          "default",
				ThreadID:       req.CollapseID, // Grouping in iOS Notification Center.
			},
			// [IOS_FOCUS] Injects interruption-level via custom map as SDK fields are restricted.
			// https://developer.apple.com/documentation/usernotifications/unnotificationinterruptionlevel
			CustomData: map[string]any{
				"interruption-level": "time-sensitive",
			},
		},
		Headers: map[string]string{
			"apns-collapse-id": req.CollapseID, // APNS-level deduplication.
		},
	}

	return m
}
