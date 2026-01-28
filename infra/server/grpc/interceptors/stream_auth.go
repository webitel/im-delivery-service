package interceptors

import (
	"context"

	"github.com/webitel/im-delivery-service/internal/domain/model"
	"github.com/webitel/im-delivery-service/internal/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type contextKey string

const (
	// AuthContextKey is the key used to store/retrieve AuthContact from context
	AuthContextKey contextKey = "auth_contact"
)

// NewStreamAuthInterceptor creates a middleware for gRPC stream connections.
func NewStreamAuthInterceptor(auther service.Auther) grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		// [PRE_AUTH] Validate identity before allowing the stream to open
		ctx := ss.Context()
		auth, err := auther.Inspect(ctx)
		if err != nil {
			return status.Errorf(codes.Unauthenticated, "authentication failed: %v", err)
		}

		// [ENRICHMENT] Inject the identity into the context for downstream handlers
		newCtx := context.WithValue(ctx, AuthContextKey, auth)

		// [STREAM_WRAPPING] Override the context of the original stream
		wrapped := &wrappedStream{
			ServerStream: ss,
			ctx:          newCtx,
		}

		// Proceed to the actual service method
		return handler(srv, wrapped)
	}
}

// wrappedStream is a thin wrapper to inject a new context into a gRPC stream.
type wrappedStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedStream) Context() context.Context {
	return w.ctx
}

// GetAuthContact is a helper to extract the identity from context safely.
func GetAuthContact(ctx context.Context) (*model.AuthContact, bool) {
	auth, ok := ctx.Value(AuthContextKey).(*model.AuthContact)
	return auth, ok
}
