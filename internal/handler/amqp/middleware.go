package amqp

import (
	"context"
	"log/slog"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/google/uuid"
)

// [TRACE_ID_MIDDLEWARE]
// Ensures TraceID persistence through the call chain.
func TraceIDMiddleware(h message.HandlerFunc) message.HandlerFunc {
	return func(msg *message.Message) ([]*message.Message, error) {
		traceID := msg.Metadata.Get("trace_id")
		if traceID == "" {
			traceID = uuid.NewString()
			msg.Metadata.Set("trace_id", traceID)
		}

		ctx := context.WithValue(msg.Context(), "trace_id", traceID)
		msg.SetContext(ctx)

		return h(msg)
	}
}

// [LOGGING_MIDDLEWARE]
// Structured logging with latency and TraceID.
func LoggingMiddleware(logger *slog.Logger) message.HandlerMiddleware {
	return func(h message.HandlerFunc) message.HandlerFunc {
		return func(msg *message.Message) ([]*message.Message, error) {
			start := time.Now()
			msgs, err := h(msg)

			logger.Debug("MESSAGE_HANDLED",
				"msg_id", msg.UUID,
				"trace_id", msg.Metadata.Get("trace_id"),
				"duration_ms", time.Since(start).Milliseconds(),
				"success", err == nil,
			)
			return msgs, err
		}
	}
}

// [RETRY_MIDDLEWARE]
func NewRetryMiddleware() middleware.Retry {
	return middleware.Retry{
		MaxRetries:      3,
		InitialInterval: time.Second * 2,
		MaxInterval:     time.Second * 15,
		Multiplier:      2.0,
	}
}
