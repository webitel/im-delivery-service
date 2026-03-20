package event

// [OPTION_TYPE] Definition for functional options pattern.
// Matches the exported fields in the Envelope struct.
type Option[T any] func(e *Envelope[T])

// [WITH_PRIORITY] Overrides the default event priority levels.
func WithPriority[T any](p EventPriority) Option[T] {
	return func(e *Envelope[T]) { e.Priority = p }
}

// [WITH_OCCURRED_AT] Explicitly sets the event creation timestamp.
func WithOccurredAt[T any](t int64) Option[T] {
	return func(e *Envelope[T]) { e.OccurredAt = t }
}

// [WITH_TRACE_ID] Injects a correlation ID for cross-service distributed tracing.
func WithTraceID[T any](id string) Option[T] {
	return func(e *Envelope[T]) { e.TraceID = id }
}

// [WITH_ECHO] Marks the event as a sender's device synchronization copy.
func WithEcho[T any](v bool) Option[T] {
	return func(e *Envelope[T]) { e.Echo = v }
}

// [WITH_TRACKED] Manually toggles whether the event triggers push/delivery tracking.
func WithTracked[T any](v bool) Option[T] {
	return func(e *Envelope[T]) { e.CanPush = v }
}
