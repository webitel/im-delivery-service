package service

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"sync"
	"testing"

	"google.golang.org/grpc"

	adminv1 "github.com/webitel/im-delivery-service/gen/go/admin/v1"
)

type fakeAdminAppSearcher struct {
	mu       sync.Mutex
	calls    int
	response *adminv1.ApplicationList
	err      error
}

func newFakeAdminAppSearcher() *fakeAdminAppSearcher {
	return &fakeAdminAppSearcher{
		response: nil,
		err:      nil,
	}
}

func (f *fakeAdminAppSearcher) SearchApps(ctx context.Context, in *adminv1.SearchAppRequest, opts ...grpc.CallOption) (*adminv1.ApplicationList, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.calls++

	return f.response, f.err
}

func (f *fakeAdminAppSearcher) callCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.calls
}

func TestResolvePolicy_EmptyAppID_ReturnsAllowAll(t *testing.T) {
	fake := newFakeAdminAppSearcher()
	svc := NewAppConfigService(fake, slog.New(slog.NewTextHandler(io.Discard, nil)))

	policy := svc.ResolvePolicy(context.Background(), "")

	if !policy.Allows("user_joined") || !policy.Allows("user_left") {
		t.Error("empty appID should return allow-all policy")
	}

	if fake.callCount() != 0 {
		t.Errorf("empty appID should not call admin client, got %d calls", fake.callCount())
	}
}

func TestResolvePolicy_RPCError_FailsOpen(t *testing.T) {
	fake := newFakeAdminAppSearcher()
	fake.err = errors.New("admin service unreachable")
	svc := NewAppConfigService(fake, slog.New(slog.NewTextHandler(io.Discard, nil)))

	policy := svc.ResolvePolicy(context.Background(), "myapp")

	if !policy.Allows("user_joined") {
		t.Error("RPC error should fail open (allow-all)")
	}
}

func TestResolvePolicy_NilResponse_FailsOpen(t *testing.T) {
	fake := newFakeAdminAppSearcher()
	fake.response = nil
	svc := NewAppConfigService(fake, slog.New(slog.NewTextHandler(io.Discard, nil)))

	policy := svc.ResolvePolicy(context.Background(), "myapp")

	if !policy.Allows("user_joined") {
		t.Error("nil response should fail open (allow-all)")
	}
}

func TestResolvePolicy_EmptyData_FailsOpen(t *testing.T) {
	fake := newFakeAdminAppSearcher()
	fake.response = &adminv1.ApplicationList{Data: []*adminv1.Application{}}
	svc := NewAppConfigService(fake, slog.New(slog.NewTextHandler(io.Discard, nil)))

	policy := svc.ResolvePolicy(context.Background(), "myapp")

	if !policy.Allows("user_joined") {
		t.Error("empty data should fail open (allow-all)")
	}
}

func TestResolvePolicy_AllowSystemMessagesNil_AllowAll(t *testing.T) {
	fake := newFakeAdminAppSearcher()
	fake.response = &adminv1.ApplicationList{
		Data: []*adminv1.Application{
			{
				Id:                  "myapp",
				AllowSystemMessages: nil,
			},
		},
	}
	svc := NewAppConfigService(fake, slog.New(slog.NewTextHandler(io.Discard, nil)))

	policy := svc.ResolvePolicy(context.Background(), "myapp")

	if !policy.Allows("user_joined") || !policy.Allows("user_left") {
		t.Error("nil AllowSystemMessages should return allow-all policy")
	}
}

func TestResolvePolicy_AllowSystemMessagesEmpty_DenyAll(t *testing.T) {
	fake := newFakeAdminAppSearcher()
	fake.response = &adminv1.ApplicationList{
		Data: []*adminv1.Application{
			{
				Id: "myapp",
				AllowSystemMessages: &adminv1.SystemMessageAllowList{
					Types: []string{},
				},
			},
		},
	}
	svc := NewAppConfigService(fake, slog.New(slog.NewTextHandler(io.Discard, nil)))

	policy := svc.ResolvePolicy(context.Background(), "myapp")

	if policy.Allows("user_joined") || policy.Allows("user_left") {
		t.Error("empty allow-list should deny all system messages")
	}
}

func TestResolvePolicy_AllowSystemMessagesWithTypes_AllowSpecific(t *testing.T) {
	fake := newFakeAdminAppSearcher()
	fake.response = &adminv1.ApplicationList{
		Data: []*adminv1.Application{
			{
				Id: "myapp",
				AllowSystemMessages: &adminv1.SystemMessageAllowList{
					Types: []string{"user_joined"},
				},
			},
		},
	}
	svc := NewAppConfigService(fake, slog.New(slog.NewTextHandler(io.Discard, nil)))

	policy := svc.ResolvePolicy(context.Background(), "myapp")

	if !policy.Allows("user_joined") {
		t.Error("user_joined should be allowed")
	}

	if policy.Allows("user_left") {
		t.Error("user_left should be denied")
	}
}

func TestSystemMessageAllowed_CachesSuccessfulResult(t *testing.T) {
	fake := newFakeAdminAppSearcher()
	fake.response = &adminv1.ApplicationList{
		Data: []*adminv1.Application{
			{
				Id: "myapp",
				AllowSystemMessages: &adminv1.SystemMessageAllowList{
					Types: []string{"user_joined"},
				},
			},
		},
	}
	svc := NewAppConfigService(fake, slog.New(slog.NewTextHandler(io.Discard, nil)))

	result1 := svc.SystemMessageAllowed(context.Background(), "myapp", "user_joined")
	result2 := svc.SystemMessageAllowed(context.Background(), "myapp", "user_joined")

	if !result1 || !result2 {
		t.Error("both calls should allow user_joined")
	}

	if fake.callCount() != 1 {
		t.Errorf("expected exactly 1 admin call (cached), got %d", fake.callCount())
	}
}

func TestResolvePolicy_ResolvesOnce_MultipleAllows(t *testing.T) {
	fake := newFakeAdminAppSearcher()
	fake.response = &adminv1.ApplicationList{
		Data: []*adminv1.Application{
			{
				Id: "myapp",
				AllowSystemMessages: &adminv1.SystemMessageAllowList{
					Types: []string{"user_joined"},
				},
			},
		},
	}
	svc := NewAppConfigService(fake, slog.New(slog.NewTextHandler(io.Discard, nil)))

	policy := svc.ResolvePolicy(context.Background(), "myapp")
	r1 := policy.Allows("user_joined")
	r2 := policy.Allows("user_left")
	r3 := policy.Allows("user_joined")

	if !r1 || r2 || !r3 {
		t.Error("unexpected Allows results")
	}

	if fake.callCount() != 1 {
		t.Errorf("expected exactly 1 admin call total, got %d", fake.callCount())
	}
}
