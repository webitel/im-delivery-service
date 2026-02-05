package ws

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"google.golang.org/grpc/metadata"
)

// ContextBridge wraps an http.Handler to inject gRPC metadata and log handshakes.
func ContextBridge(handler http.Handler, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// [METADATA] Convert HTTP headers to lowercase gRPC metadata.
		md := metadata.MD{}
		for k, v := range r.Header {
			// [FILTER] Skip hop-by-hop and WebSocket-specific headers that break gRPC/HTTP2
			lowerKey := strings.ToLower(k)
			if lowerKey == "connection" ||
				lowerKey == "upgrade" ||
				strings.HasPrefix(lowerKey, "sec-websocket-") {
				continue
			}

			md.Set(k, v...)
		}
		ctx := metadata.NewIncomingContext(r.Context(), md)

		// [OBSERVABILITY] Trace the handshake start.
		logger.Debug("WS_HANDSHAKE_STARTED",
			slog.String("ip", r.RemoteAddr),
			slog.String("path", r.URL.Path),
		)

		handler.ServeHTTP(w, r.WithContext(ctx))

		logger.Debug("WS_HANDSHAKE_FINISHED",
			slog.Duration("duration", time.Since(start)),
		)
	})
}
