package ws

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type contextKey string

const authInfoKey contextKey = "auth_info"

func (h *WSHandler) AuthenticationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// [1] Attempt to extract credentials from headers or query params.
		hToken := r.Header.Get("x-webitel-access")
		qToken := r.URL.Query().Get("x-webitel-access")
		clientID := r.Header.Get("x-webitel-client")
		if clientID == "" {
			clientID = r.URL.Query().Get("x-webitel-client")
		}

		token := hToken
		if token == "" {
			token = qToken
		}

		// [2] If no token is present, decide whether to allow an unauthenticated Upgrade.
		if token == "" {
			if clientID != "" {
				// Client ID without a token is a logic error; reject immediately.
				h.log.Warn("ws: auth failed, client_id without token", slog.String("remote", r.RemoteAddr))
				http.Error(w, "401 Unauthorized: token required", http.StatusUnauthorized)
				return
			}
			// Completely empty request - proceed to late-binding (waitAuthFrame).
			next.ServeHTTP(w, r)
			return
		}

		// [3] Perform identity inspection via gRPC.
		md := metadata.Pairs("x-webitel-access", token, "x-webitel-client", clientID)
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()

		auth, err := h.auther.Inspect(metadata.NewIncomingContext(ctx, md))
		if err != nil {
			st, _ := status.FromError(err)
			h.log.Error("ws: handshake auth failed",
				slog.String("remote", r.RemoteAddr),
				slog.String("err", st.Message()),
			)

			// Map gRPC errors to HTTP statuses.
			switch st.Code() {
			case codes.Unauthenticated, codes.InvalidArgument, codes.Unknown:
				http.Error(w, "401 Unauthorized: "+st.Message(), http.StatusUnauthorized)
			case codes.PermissionDenied:
				http.Error(w, "403 Forbidden", http.StatusForbidden)
			default:
				http.Error(w, "503 Service Unavailable", http.StatusServiceUnavailable)
			}
			return
		}

		// [4] Success: inject auth data and proceed to Upgrade.
		newCtx := context.WithValue(r.Context(), authInfoKey, auth)
		next.ServeHTTP(w, r.WithContext(newCtx))
	})
}
