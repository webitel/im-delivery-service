package ws

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"google.golang.org/grpc/metadata"
)

type contextKey string

const authInfoKey contextKey = "auth_info"

func (h *WSHandler) AuthenticationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// [AUDIT] Log the initial handshake request
		h.log.Debug("ws: intercepting handshake request",
			slog.String("remote", r.RemoteAddr),
			slog.String("origin", r.Header.Get("Origin")),
			slog.String("path", r.URL.Path),
		)

		md := metadata.MD{}
		for k, v := range r.Header {
			lowerKey := strings.ToLower(k)
			if lowerKey == "connection" || lowerKey == "upgrade" || strings.HasPrefix(lowerKey, "sec-websocket-") {
				continue
			}
			md.Set(k, v...)
		}

		// [CHANGE] Support x-webitel-client and access tokens in query
		if token := r.URL.Query().Get("x-webitel-access"); token != "" {
			md.Set("x-webitel-access", token)
		} else if token := r.URL.Query().Get("token"); token != "" {
			md.Set("x-webitel-access", token)
		}

		if client := r.URL.Query().Get("x-webitel-client"); client != "" {
			md.Set("x-webitel-client", client)
		}

		ctx := metadata.NewIncomingContext(r.Context(), md)
		auth, err := h.auther.Inspect(ctx)

		if err != nil {
			// [LOG] Log failed pre-auth but don't block (might be late-binding auth)
			h.log.Debug("ws: middleware pre-auth failed",
				slog.String("remote", r.RemoteAddr),
				slog.Any("err", err),
			)
		} else {
			h.log.Info("ws: middleware pre-auth success",
				slog.String("user_id", auth.ContactID),
				slog.String("remote", r.RemoteAddr),
			)
			ctx = context.WithValue(ctx, authInfoKey, auth)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
