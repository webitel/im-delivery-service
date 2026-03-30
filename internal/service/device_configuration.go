// internal/service/device_provider.go
package service

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/google/uuid"
	adminv1 "github.com/webitel/im-delivery-service/gen/go/admin/v1"
	authv1 "github.com/webitel/im-delivery-service/gen/go/auth/v1"
	imadmin "github.com/webitel/im-delivery-service/infra/client/im-admin"
	imauth "github.com/webitel/im-delivery-service/infra/client/im-auth"
	"github.com/webitel/im-delivery-service/internal/domain/model"
	"github.com/webitel/im-delivery-service/internal/store"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

const (
	ProviderFCM  = "fcm"
	ProviderAPNS = "apn"
	ProviderWEB  = "web"
)

// DeviceProvider defines the contract for user device discovery and synchronization.
type DeviceProvider interface {
	GetDevices(ctx context.Context, userID uuid.UUID) ([]model.Device, error)
	Sync(ctx context.Context, userID uuid.UUID) ([]model.Device, error)
}

// DeviceService handles orchestration between auth, admin services, and local cache.
type DeviceService struct {
	cache store.PresenceStore
	auth  *imauth.Client
	admin *imadmin.Client
	log   *slog.Logger
}

func NewDeviceService(cache store.PresenceStore, auth *imauth.Client, admin *imadmin.Client, log *slog.Logger) *DeviceService {
	return &DeviceService{
		cache: cache,
		auth:  auth,
		admin: admin,
		log:   log.With("component", "device_service"),
	}
}

// GetDevices returns active devices from cache or performs a full sync on miss.
func (s *DeviceService) GetDevices(ctx context.Context, userID uuid.UUID) ([]model.Device, error) {
	if cached, err := s.cache.UserDevices(ctx, userID); err == nil && cached != nil {
		return *cached, nil
	}

	return s.Sync(ctx, userID)
}

// Sync refreshes the device state from external services and updates the cache.
func (s *DeviceService) Sync(ctx context.Context, userID uuid.UUID) ([]model.Device, error) {
	// 1. Fetch from source
	devices, err := s.pull(ctx, userID)
	if err != nil {
		return nil, err
	}

	if len(devices) == 0 {
		_ = s.cache.SyncDevices(ctx, userID, nil)
		return nil, nil
	}

	// 2. Enrich with app configurations
	enriched := s.enrich(ctx, devices)

	// 3. Update cache
	if err := s.cache.SyncDevices(ctx, userID, enriched); err != nil {
		s.log.Error("cache_sync_failed", slog.String("user_id", userID.String()), slog.Any("err", err))
	}

	return enriched, nil
}

// pull retrieves authorizations and maps them to domain models.
func (s *DeviceService) pull(ctx context.Context, userID uuid.UUID) ([]model.Device, error) {
	resp, err := s.auth.GetAuthorizations(ctx, &authv1.GetAuthorizationRequest{
		Contact: &authv1.InputContact{Input: &authv1.InputContact_Id{Id: userID.String()}},
		Push:    &wrapperspb.BoolValue{Value: true},
	})
	if err != nil {
		return nil, err
	}

	return s.toDomain(resp), nil
}

// enrich fetches push credentials for each unique application ID.
func (s *DeviceService) enrich(ctx context.Context, devices []model.Device) []model.Device {
	memo := make(map[string]*adminv1.Application)

	for i := range devices {
		appID := devices[i].AppID
		if appID == "" {
			continue
		}

		app, ok := memo[appID]
		if !ok {
			res, err := s.admin.SearchApps(ctx, &adminv1.SearchAppRequest{Id: appID})
			if err != nil || res == nil || len(res.Data) == 0 {
				continue
			}
			app = res.Data[0]
			memo[appID] = app
		}

		if app.Service != nil && app.Service.PushService != nil {
			s.apply(&devices[i], app.Service.PushService)
		}
	}
	return devices
}

// apply binds provider-specific credentials to the device model.
func (s *DeviceService) apply(d *model.Device, ps *adminv1.PUSHServiceClient) {
	switch d.PushType {
	case ProviderFCM:
		if fcm := ps.GetFcm(); fcm != nil {
			d.PushConfig.Proxy = fcm.Proxy
			d.PushConfig.Credentials = fcm.Account
		}
	case ProviderAPNS:
		if apn := ps.GetApn(); apn != nil {
			d.PushConfig.Proxy = apn.GetProxy()
			d.PushConfig.Topic = apn.GetTopic()
			if token := apn.GetToken(); token != nil {
				d.PushConfig.Credentials = token.GetAuthKey()
				d.PushConfig.KeyID = token.GetKeyId()
				d.PushConfig.TeamID = token.GetTeamId()
			}
		}
	case ProviderWEB:
		if web := ps.GetWeb(); web != nil {
			d.PushConfig.Proxy = web.Proxy
			d.PushConfig.Credentials = web.Token
		}
	}
}

// map converts proto authorizations into internal domain devices.
func (s *DeviceService) toDomain(src *authv1.AuthorizationList) []model.Device {
	if src == nil {
		return nil
	}

	dst := make([]model.Device, 0, len(src.Data))
	for _, a := range src.Data {
		if a.Device == nil || a.Device.Push == nil {
			continue
		}

		device := model.Device{ID: a.Device.Id, AppID: a.AppId}

		switch t := a.Device.Push.Token.(type) {
		case *authv1.PUSHSubscription_Fcm:
			device.PushType, device.PushToken, device.Platform = ProviderFCM, t.Fcm, model.PlatformAndroid
		case *authv1.PUSHSubscription_Apn:
			device.PushType, device.PushToken, device.Platform = ProviderAPNS, t.Apn, model.PlatformIOS
		case *authv1.PUSHSubscription_Web:
			device.PushType, device.Platform = ProviderWEB, model.PlatformWeb
			if b, err := json.Marshal(t.Web); err == nil {
				device.PushToken = string(b)
			}
		}

		if device.PushToken != "" {
			dst = append(dst, device)
		}
	}
	return dst
}
