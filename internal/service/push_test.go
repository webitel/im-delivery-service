package service

import (
	"context"
	"testing"

	"github.com/webitel/im-delivery-service/internal/domain/model"
)

// fakeAppConfigProvider implements AppConfigProvider for testing.
type fakeAppConfigProvider struct {
	policies map[string]SystemMessagePolicy
}

func (f *fakeAppConfigProvider) SystemMessageAllowed(ctx context.Context, appID, systemType string) bool {
	policy, ok := f.policies[appID]
	if !ok {
		// Missing key = zero-value = allow-all
		return true
	}

	return policy.Allows(systemType)
}

func (f *fakeAppConfigProvider) ResolvePolicy(ctx context.Context, appID string) SystemMessagePolicy {
	policy, ok := f.policies[appID]
	if !ok {
		return SystemMessagePolicy{}
	}

	return policy
}

func TestFilterSystemMessagePolicy(t *testing.T) {
	t.Run("app not in map passes all devices through", func(t *testing.T) {
		fake := &fakeAppConfigProvider{
			policies: map[string]SystemMessagePolicy{},
		}

		h := &PushHandler{
			appConfig: fake,
		}

		devices := []model.Device{
			{ID: "dev1", AppID: "app1"},
			{ID: "dev2", AppID: "app1"},
		}

		filtered := h.filterSystemMessagePolicy(context.Background(), devices, "user_joined")
		if len(filtered) != 2 {
			t.Errorf("expected 2 devices, got %d", len(filtered))
		}
	})

	t.Run("zero-value policy allows all devices", func(t *testing.T) {
		fake := &fakeAppConfigProvider{
			policies: map[string]SystemMessagePolicy{
				"app1": {},
			},
		}

		h := &PushHandler{
			appConfig: fake,
		}

		devices := []model.Device{
			{ID: "dev1", AppID: "app1"},
			{ID: "dev2", AppID: "app1"},
		}

		filtered := h.filterSystemMessagePolicy(context.Background(), devices, "user_joined")
		if len(filtered) != 2 {
			t.Errorf("expected 2 devices, got %d", len(filtered))
		}
	})

	t.Run("block-all policy removes all devices", func(t *testing.T) {
		fake := &fakeAppConfigProvider{
			policies: map[string]SystemMessagePolicy{
				"app1": {restricted: true},
			},
		}

		h := &PushHandler{
			appConfig: fake,
		}

		devices := []model.Device{
			{ID: "dev1", AppID: "app1"},
			{ID: "dev2", AppID: "app1"},
		}

		filtered := h.filterSystemMessagePolicy(context.Background(), devices, "user_joined")
		if len(filtered) != 0 {
			t.Errorf("expected 0 devices, got %d", len(filtered))
		}
	})

	t.Run("allow-list policy keeps only matching devices", func(t *testing.T) {
		fake := &fakeAppConfigProvider{
			policies: map[string]SystemMessagePolicy{
				"app1": {restricted: true, allowed: []string{"user_joined"}},
				"app2": {restricted: true, allowed: []string{"user_left"}},
			},
		}

		h := &PushHandler{
			appConfig: fake,
		}

		devices := []model.Device{
			{ID: "dev1", AppID: "app1"},
			{ID: "dev2", AppID: "app2"},
			{ID: "dev3", AppID: "app1"},
		}

		filtered := h.filterSystemMessagePolicy(context.Background(), devices, "user_joined")
		if len(filtered) != 2 {
			t.Errorf("expected 2 devices, got %d", len(filtered))
		}

		// Check that only app1 devices remain
		for _, d := range filtered {
			if d.AppID != "app1" {
				t.Errorf("expected all filtered devices to be from app1, got %s", d.AppID)
			}
		}
	})

	t.Run("nil appConfig returns devices unchanged", func(t *testing.T) {
		h := &PushHandler{
			appConfig: nil,
		}

		devices := []model.Device{
			{ID: "dev1", AppID: "app1"},
			{ID: "dev2", AppID: "app2"},
		}

		filtered := h.filterSystemMessagePolicy(context.Background(), devices, "user_joined")
		if len(filtered) != 2 {
			t.Errorf("expected 2 devices, got %d", len(filtered))
		}
	})

	t.Run("multiple apps with different policies filters correctly", func(t *testing.T) {
		fake := &fakeAppConfigProvider{
			policies: map[string]SystemMessagePolicy{
				"app1": {restricted: true, allowed: []string{"user_joined", "user_left"}},
				"app2": {restricted: true, allowed: []string{"user_left"}},
				"app3": {},
			},
		}

		h := &PushHandler{
			appConfig: fake,
		}

		devices := []model.Device{
			{ID: "dev1", AppID: "app1"},
			{ID: "dev2", AppID: "app2"},
			{ID: "dev3", AppID: "app3"},
			{ID: "dev4", AppID: "app1"},
			{ID: "dev5", AppID: "app2"},
		}

		filtered := h.filterSystemMessagePolicy(context.Background(), devices, "user_joined")
		if len(filtered) != 3 {
			t.Errorf("expected 3 devices, got %d", len(filtered))
		}

		// Check that we have app1 (2) and app3 (1) devices
		appCounts := make(map[string]int)
		for _, d := range filtered {
			appCounts[d.AppID]++
		}

		if appCounts["app1"] != 2 {
			t.Errorf("expected 2 devices from app1, got %d", appCounts["app1"])
		}

		if appCounts["app3"] != 1 {
			t.Errorf("expected 1 device from app3, got %d", appCounts["app3"])
		}
	})
}
