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
		// [1] Extraction
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

		// [2] Decision Logic
		// If there is NO token, we allow the Upgrade regardless of whether ClientID exists.
		// The 'waitAuthFrame' will handle the 5-second timeout for the missing token.
		if token == "" {
			h.log.Debug("ws: no token in headers, allowing upgrade for late-binding auth",
				slog.String("remote", r.RemoteAddr))
			next.ServeHTTP(w, r)

			return
		}

		// [3] Fast-track: Token is present, validate it immediately.
		md := metadata.Pairs("x-webitel-access", token)
		if clientID != "" {
			md.Set("x-webitel-client", clientID)
		}

		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()

		auth, err := h.auther.Inspect(metadata.NewIncomingContext(ctx, md))
		if err != nil {
			st, _ := status.FromError(err)
			h.log.Error("ws: handshake auth failed",
				slog.String("remote", r.RemoteAddr),
				slog.String("err", st.Message()),
			)

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

		// [4] Success
		newCtx := context.WithValue(r.Context(), authInfoKey, auth)
		next.ServeHTTP(w, r.WithContext(newCtx))
	})
}
