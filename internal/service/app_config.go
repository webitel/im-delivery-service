package service

import (
	"context"
	"log/slog"
	"slices"
	"time"

	"github.com/hashicorp/golang-lru/v2/expirable"
	"golang.org/x/sync/singleflight"
	"google.golang.org/grpc"

	adminv1 "github.com/webitel/im-delivery-service/gen/go/admin/v1"
)

const (
	// [SUCCESS_CACHE_TTL] Admin-service toggle changes must reach new connections within this window.
	appConfigCacheTTL = 60 * time.Second
	// [FAILURE_CACHE_TTL] Transient failures are cached much shorter to avoid fail-open-caching for too long.
	appConfigFailureCacheTTL = 5 * time.Second
	// [CACHE_SIZE] Distinct application IDs (not users) -- expected to be small.
	appConfigCacheSize = 1024
)

// AppConfigProvider resolves per-application delivery policy for chat "system" messages.
type AppConfigProvider interface {
	// SystemMessageAllowed reports whether a session authenticated through appID
	// should receive a chat system message of the given model.System.Type (e.g.
	// "user_joined", "user_left"). It fails open (returns true) on any error --
	// empty appID, RPC failure, app not found -- so a misbehaving/unreachable
	// admin-service (or a transport with no appID, e.g. long-polling) never
	// globally suppresses chat system messages.
	SystemMessageAllowed(ctx context.Context, appID, systemType string) bool

	// ResolvePolicy resolves the system message policy for an application.
	// Caches the result and can be called multiple times with the same appID
	// to retrieve the cached policy. Fails open (returns allow-all policy) on error.
	// This method is primarily for use at session Attach time; once resolved and stored,
	// the policy should be reused without further RPC calls.
	ResolvePolicy(ctx context.Context, appID string) SystemMessagePolicy
}

// AdminAppSearcher is a narrow interface exposing only SearchApps from the admin client.
// This mirrors the ThreadStatusClient pattern: instead of importing the full *imadmin.Client
// concrete type into the service layer (which would make testing difficult), we define
// a minimal interface here. internal/service/di/module.go provides an adapter that
// wires *imadmin.Client as an AdminAppSearcher via fx.Annotate, keeping the service
// layer decoupled from the concrete client implementation.
type AdminAppSearcher interface {
	SearchApps(ctx context.Context, in *adminv1.SearchAppRequest, opts ...grpc.CallOption) (*adminv1.ApplicationList, error)
}

// SystemMessagePolicy encapsulates the decision logic for whether a system message type is allowed.
// It has three states:
//   - Zero-value (restricted=false, allowed=nil): not configured -> allow all system messages.
//   - Restricted but empty (restricted=true, allowed=[]): block all system messages.
//   - Restricted with allowed types (restricted=true, allowed=[...]): allow only listed types.
type SystemMessagePolicy struct {
	restricted bool
	allowed    []string
}

// Allows reports whether a system message of the given type is allowed by this policy.
func (p SystemMessagePolicy) Allows(systemType string) bool {
	if !p.restricted {
		return true
	}

	return slices.Contains(p.allowed, systemType)
}

// AppConfigService implements AppConfigProvider using a cached, single-flighted admin lookup.
type AppConfigService struct {
	admin        AdminAppSearcher
	successCache *expirable.LRU[string, SystemMessagePolicy]
	failureCache *expirable.LRU[string, struct{}]
	singleflight singleflight.Group
	logger       *slog.Logger
}

var _ AppConfigProvider = (*AppConfigService)(nil)

func NewAppConfigService(admin AdminAppSearcher, logger *slog.Logger) *AppConfigService {
	return &AppConfigService{
		admin:        admin,
		successCache: expirable.NewLRU[string, SystemMessagePolicy](appConfigCacheSize, nil, appConfigCacheTTL),
		failureCache: expirable.NewLRU[string, struct{}](appConfigCacheSize, nil, appConfigFailureCacheTTL),
		logger:       logger.With("component", "app_config_service"),
	}
}

// SystemMessageAllowed is a one-liner delegating to ResolvePolicy.
func (s *AppConfigService) SystemMessageAllowed(ctx context.Context, appID, systemType string) bool {
	return s.ResolvePolicy(ctx, appID).Allows(systemType)
}

// ResolvePolicy resolves the system message policy for an application.
// Returns a zero-value (allow-all) policy immediately for empty appID to fail open
// (e.g., long-polling with no auth context). Otherwise delegates to policyFor.
func (s *AppConfigService) ResolvePolicy(ctx context.Context, appID string) SystemMessagePolicy {
	if appID == "" {
		return SystemMessagePolicy{}
	}

	return s.policyFor(ctx, appID)
}

// policyFor checks the success cache, then the failure cache, then calls resolve via singleflight.
// This ensures at most one concurrent RPC call per appID even under concurrent ResolvePolicy callers.
func (s *AppConfigService) policyFor(ctx context.Context, appID string) SystemMessagePolicy {
	// Check success cache first
	if policy, ok := s.successCache.Get(appID); ok {
		return policy
	}

	// Check failure cache: a recent failure means return allow-all without re-hitting RPC
	if _, ok := s.failureCache.Get(appID); ok {
		return SystemMessagePolicy{}
	}

	// Use singleflight to call resolve exactly once even under concurrent callers
	v, _, _ := s.singleflight.Do(appID, func() (any, error) {
		policy, ok := s.resolve(ctx, appID)
		if !ok {
			s.failureCache.Add(appID, struct{}{})

			return SystemMessagePolicy{}, nil
		}

		s.successCache.Add(appID, policy)

		return policy, nil
	})

	policy, ok := v.(SystemMessagePolicy)
	if !ok {
		return SystemMessagePolicy{}
	}

	return policy
}

// resolve fetches the allow-list for an application from the admin service.
// On any error, nil response, or empty Data, logs a Warn and returns (zero-value policy, false).
// On success, reads AllowSystemMessages field:
//   - If nil: returns (allow-all policy, true).
//   - If present: returns (restricted policy with allowed types, true).
//
// AppID provenance: SessionService.Attach passes the live-session AppID from
// Authorization.AppId (auth.go's AuthService.Inspect), and PushHandler passes each
// device's AppID from the same Authorization.AppId field (device_configuration.go) --
// both paths resolve this appID against the same admin-service client_id.
//
// Note on field masks: No field-mask or Fields selector is set on the SearchAppRequest.
// If admin-service ever honors field masks for this RPC, the allow_system_messages field
// could be silently omitted from the response. This should be flagged if it happens.
func (s *AppConfigService) resolve(ctx context.Context, appID string) (SystemMessagePolicy, bool) {
	res, err := s.admin.SearchApps(ctx, &adminv1.SearchAppRequest{Id: appID})
	if err != nil || res == nil || len(res.GetData()) == 0 {
		s.logger.Warn("APP_CONFIG_LOOKUP_FAILED",
			slog.String("app_id", appID),
			slog.Any("err", err))

		return SystemMessagePolicy{}, false
	}

	app := res.GetData()[0]

	// [IDENTITY_CHECK] Do not blindly trust the first result: without this, a
	// permissive/relaxed Id filter on the admin-service side (or a non-unique
	// client_id) could silently apply a DIFFERENT application's policy here --
	// a cross-tenant delivery-policy decision, not just a wrong log line.
	if app.GetId() != appID {
		s.logger.Warn("APP_CONFIG_LOOKUP_MISMATCH",
			slog.String("app_id", appID),
			slog.String("returned_id", app.GetId()))

		return SystemMessagePolicy{}, false
	}

	allowList := app.GetAllowSystemMessages()

	// Not configured on admin-service: allow all
	if allowList == nil {
		return SystemMessagePolicy{}, true
	}

	// Configured with types: return restricted policy with allowed list
	return SystemMessagePolicy{
		restricted: true,
		allowed:    allowList.GetTypes(),
	}, true
}
