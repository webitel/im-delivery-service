package amqp

import (
	"context"
	"log/slog"

	"github.com/google/uuid"

	"github.com/webitel/im-delivery-service/internal/domain/event"
	"github.com/webitel/im-delivery-service/internal/handler/amqp/payload"
	"github.com/webitel/webitel-go-kit/pkg/semconv"
)

// [ON_DEVICE_REGISTERED_V1] Handles new push token registration.
func (h *MessageHandler) OnDeviceRegisteredV1(ctx context.Context, raw *payload.DeviceConfigurationUpdate) ([]event.Eventer, error) {
	return nil, h.syncUserDevices(ctx, raw, "DEVICE_REGISTERED")
}

// [ON_DEVICE_UNREGISTERED_V1] Handles push token removal.
func (h *MessageHandler) OnDeviceUnregisteredV1(ctx context.Context, raw *payload.DeviceConfigurationUpdate) ([]event.Eventer, error) {
	return nil, h.syncUserDevices(ctx, raw, "DEVICE_UNREGISTERED")
}

// [ON_DEVICE_LOGOUT_V1] Handles user logout.
func (h *MessageHandler) OnDeviceLogoutV1(ctx context.Context, raw *payload.DeviceConfigurationUpdate) ([]event.Eventer, error) {
	return nil, h.syncUserDevices(ctx, raw, "USER_LOGOUT")
}

func (h *MessageHandler) syncUserDevices(ctx context.Context, raw *payload.DeviceConfigurationUpdate, eventType string) error {
	if raw.Authorization == nil || raw.Authorization.Contact == nil {
		h.logger.Debug("AUTH_EVENT_IGNORED_EMPTY_CONTACT", "event", eventType)

		return nil
	}

	uid, err := uuid.Parse(raw.Authorization.Contact.ID)
	if err != nil {
		h.logger.Error("AUTH_EVENT_INVALID_UID", semconv.ErrorKey, err, "event", eventType)
		return nil
	}

	h.logger.Info("AUTH_SYNC_TRIGGERED", slog.String("uid", uid.String()), slog.String("reason", eventType))

	_, err = h.deviceProvider.Sync(ctx, uid)

	return err
}
