package push

import (
	"context"
	"log/slog"

	"github.com/webitel/im-delivery-service/internal/domain/model"
)

// Provider defines the universal contract for all push delivery channels (FCM, APNs, etc.).
type Provider interface {
	Send(ctx context.Context, req *model.PushRequest) error
	Dismiss(ctx context.Context, req *model.PushRequest) error
	Name() string
}

// MultiProvider acts as an orchestrator that broadcasts requests to all registered drivers.
// It uses the Composite pattern to manage multiple delivery channels transparently.
type MultiProvider struct {
	drivers []Provider
	log     *slog.Logger
}

// NewMultiProvider creates a new instance of the orchestrator.
// It automatically filters out nil providers to prevent runtime panics.
func NewMultiProvider(log *slog.Logger, drivers ...Provider) *MultiProvider {
	active := make([]Provider, 0, len(drivers))
	for _, d := range drivers {
		if d != nil {
			active = append(active, d)
		}
	}

	return &MultiProvider{
		drivers: active,
		log:     log.With("component", "push_multi_provider"),
	}
}

// Name returns the provider identifier.
func (m *MultiProvider) Name() string { return "multi_provider" }

// Send broadcasts the push request to all active drivers.
func (m *MultiProvider) Send(ctx context.Context, req *model.PushRequest) error {
	for _, d := range m.drivers {
		// Safety check to ensure we don't call a nil pointer
		if d == nil {
			continue
		}

		if err := d.Send(ctx, req); err != nil {
			m.log.Error("failed to send push via driver",
				slog.String("driver", d.Name()),
				slog.String("user_id", req.UserID),
				slog.Any("error", err),
			)
			// Continue to the next driver even if one fails
			continue
		}

		m.log.Debug("push sent successfully",
			slog.String("driver", d.Name()),
			slog.String("user_id", req.UserID),
		)
	}

	return nil
}

// Dismiss broadcasts the dismiss/cancel request to all active drivers.
func (m *MultiProvider) Dismiss(ctx context.Context, req *model.PushRequest) error {
	for _, d := range m.drivers {
		if d == nil {
			continue
		}

		if err := d.Dismiss(ctx, req); err != nil {
			m.log.Warn("failed to dismiss push via driver",
				slog.String("driver", d.Name()),
				slog.Any("error", err),
			)
		}
	}

	return nil
}
