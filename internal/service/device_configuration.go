// internal/service/device_resolver.go
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
	FCM  = "fcm"
	APNS = "apn"
	WEB  = "web"
)

// [INTERFACE GUARDJK]
var _ Configurator = (*DeviceResolver)(nil)

// DeviceResolver defines the contract for fetching and enriching user devices.
type Configurator interface {
	GetDevices(ctx context.Context, uid uuid.UUID) ([]model.Device, error)
}

// [RESOLVER] Responsible for fetching, mapping, and enriching device data.
type DeviceResolver struct {
	presence store.PresenceStore
	imauth   *imauth.Client
	imadmin  *imadmin.Client
	log      *slog.Logger
}

func NewDeviceResolver(ps store.PresenceStore, auth *imauth.Client, admin *imadmin.Client, log *slog.Logger) *DeviceResolver {
	return &DeviceResolver{
		presence: ps,
		imauth:   auth,
		imadmin:  admin,
		log:      log.With("component", "device_resolver"),
	}
}

// [GET_ENRICHED_DEVICES] Entry point for device lookup.
func (r *DeviceResolver) GetDevices(ctx context.Context, uid uuid.UUID) ([]model.Device, error) {
	// 1. Try Cache
	if cached, err := r.presence.UserDevices(ctx, uid); err == nil && cached != nil {
		return *cached, nil
	}

	// 2. Fetch Authorizations
	resp, err := r.imauth.GetAuthorizations(ctx, &authv1.GetAuthorizationRequest{
		Contact: &authv1.InputContact{Input: &authv1.InputContact_Id{Id: uid.String()}},
		Push:    &wrapperspb.BoolValue{Value: true},
	})
	if err != nil {
		return nil, err
	}

	devices := r.mapToDomain(resp)
	if len(devices) == 0 {
		return devices, nil
	}

	// 3. Enrich with App Configs
	enriched, err := r.enrichConfigs(ctx, devices)
	if err != nil {
		r.log.Warn("ENRICHMENT_FAILED", slog.Any("err", err))
	}

	// 4. Sync Cache
	_ = r.presence.SyncDevices(ctx, uid, enriched)
	return enriched, nil
}

func (r *DeviceResolver) enrichConfigs(ctx context.Context, devices []model.Device) ([]model.Device, error) {
	appConfigs := make(map[string]*adminv1.Application)
	for _, device := range devices {
		if device.AppID == "" || appConfigs[device.AppID] != nil {
			continue
		}

		res, err := r.imadmin.SearchApps(ctx, &adminv1.SearchAppRequest{Id: device.AppID})
		if err != nil {
			r.log.Warn("APP_LOOKUP_FAILED", slog.String("app", device.AppID), slog.Any("err", err))
			continue
		}

		if res != nil && len(res.Data) > 0 {
			appConfigs[device.AppID] = res.Data[0]
		}
	}

	for i := range devices {
		app, ok := appConfigs[devices[i].AppID]
		if !ok || app.Service == nil || app.Service.PushService == nil {
			continue
		}
		r.applyProviderConfig(&devices[i], app.Service.PushService)
	}
	return devices, nil
}

// [APPLY] Maps generated proto configurations to the internal device model.
func (r *DeviceResolver) applyProviderConfig(d *model.Device, ps *adminv1.PUSHServiceClient) {
	if ps == nil {
		return
	}

	switch d.PushType {
	case FCM:
		// [FCM] Use GetFcm() to safely access the nested Fcm client config.
		if fcm := ps.GetFcm(); fcm != nil {
			d.PushConfig.Proxy = fcm.Proxy
			// Map 'Account' from proto to 'Credentials' in your model.
			d.PushConfig.Credentials = fcm.Account
		}

	case APNS:
		if apn := ps.GetApn(); apn != nil {
			d.PushConfig.Proxy = apn.GetProxy()
			d.PushConfig.Topic = apn.GetTopic()
			// [TOKEN_AUTH] Extract Apple specific metadata
			if token := apn.GetToken(); token != nil {
				d.PushConfig.Credentials = token.GetAuthKey()
				d.PushConfig.KeyID = token.GetKeyId()
				d.PushConfig.TeamID = token.GetTeamId()
			}
		}

	case WEB:
		// [WEB] Webitel push service client.
		if web := ps.GetWeb(); web != nil {
			d.PushConfig.Proxy = web.Proxy
			d.PushConfig.Credentials = web.Token
		}
	}
}

func (r *DeviceResolver) mapToDomain(resp *authv1.AuthorizationList) []model.Device {
	res := make([]model.Device, 0)
	if resp == nil {
		return res
	}
	for _, a := range resp.Data {
		if a.Device == nil || a.Device.Push == nil {
			continue
		}
		d := model.Device{ID: a.Device.Id, AppID: a.AppId}
		switch t := a.Device.Push.Token.(type) {
		case *authv1.PUSHSubscription_Fcm:
			d.PushType, d.PushToken, d.Platform = "fcm", t.Fcm, model.PlatformAndroid
		case *authv1.PUSHSubscription_Apn:
			d.PushType, d.PushToken, d.Platform = "apn", t.Apn, model.PlatformIOS
		case *authv1.PUSHSubscription_Web:
			d.PushType, d.Platform = "web", model.PlatformWeb
			if b, err := json.Marshal(t.Web); err == nil {
				d.PushToken = string(b)
			}
		}
		if d.PushToken != "" {
			res = append(res, d)
		}
	}
	return res
}
