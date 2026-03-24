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
		token := r.Header.Get("x-webitel-access")
		if token == "" {
			token = r.URL.Query().Get("x-webitel-access")
		}

		if token == "" {
			next.ServeHTTP(w, r)
			return
		}

		md := metadata.Pairs(
			"x-webitel-access", token,
			"x-webitel-client", r.URL.Query().Get("x-webitel-client"),
		)

		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()

		auth, err := h.auther.Inspect(metadata.NewIncomingContext(ctx, md))
		if err != nil {
			st, _ := status.FromError(err)

			// [LOG_HANDSHAKE] Audit failed attempts
			h.log.Error("ws: handshake auth failed",
				slog.String("remote", r.RemoteAddr),
				slog.String("error", st.Message()),
				slog.Int("code", int(st.Code())),
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

		newCtx := context.WithValue(r.Context(), authInfoKey, auth)
		next.ServeHTTP(w, r.WithContext(newCtx))
	})
}
