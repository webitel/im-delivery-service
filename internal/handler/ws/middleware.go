package ws

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"google.golang.org/grpc/metadata"
)

// [CONTEXT_BRIDGE] Prepares the context with gRPC-compatible metadata from HTTP.
// ---------------------------------------------------------------------------------
// [LOGIC]
// - Normalizes headers to lowercase.
// - Filters out connection-specific headers that interfere with gRPC calls.
// - Enables seamless usage of service.Auther.Inspect.
// ---------------------------------------------------------------------------------
func ContextBridge(handler http.Handler, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		md := metadata.MD{}
		for k, v := range r.Header {
			lowerKey := strings.ToLower(k)
			// [FILTER] Remove hop-by-hop headers.
			if lowerKey == "connection" || lowerKey == "upgrade" || strings.HasPrefix(lowerKey, "sec-websocket-") {
				continue
			}
			md.Set(k, v...)
		}

		// [FALLBACK] Support token in query for specific client scenarios.
		if token := r.URL.Query().Get("token"); token != "" {
			md.Set("x-webitel-access", token)
		}

		// [TRACE] Handshake lifecycle logging.
		logger.Debug("WS_HANDSHAKE_START", slog.String("remote", r.RemoteAddr))

		handler.ServeHTTP(w, r.WithContext(metadata.NewIncomingContext(r.Context(), md)))

		logger.Debug("WS_HANDSHAKE_END", slog.Duration("took", time.Since(start)))
	})
}
