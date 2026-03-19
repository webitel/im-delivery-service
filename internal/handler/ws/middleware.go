// internal/handler/ws/middleware.go
package ws

import (
	"context"
	"net/http"
	"strings"

	"google.golang.org/grpc/metadata"
)

type contextKey string

const authInfoKey contextKey = "auth_info"

func (h *WSHandler) AuthenticationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		md := metadata.MD{}
		for k, v := range r.Header {
			lowerKey := strings.ToLower(k)
			if lowerKey == "connection" || lowerKey == "upgrade" || strings.HasPrefix(lowerKey, "sec-websocket-") {
				continue
			}
			md.Set(k, v...)
		}

		if token := r.URL.Query().Get("token"); token != "" {
			md.Set("x-webitel-access", token)
		}

		ctx := metadata.NewIncomingContext(r.Context(), md)
		auth, err := h.auther.Inspect(ctx)

		if err == nil {
			ctx = context.WithValue(ctx, authInfoKey, auth)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
