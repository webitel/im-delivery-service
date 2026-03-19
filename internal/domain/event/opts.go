package event

// [WITH_PRIORITY] Overrides the default event priority.
func WithPriority[T any](p EventPriority) Option[T] {
	return func(e *Envelope[T]) { e.priority = p }
}

// [WITH_OCCURRED_AT] Overrides the event creation timestamp.
func WithOccurredAt[T any](t int64) Option[T] {
	return func(e *Envelope[T]) { e.occurredAt = t }
}

// [WITH_TRACE_ID] Injects a correlation ID for distributed tracing.
func WithTraceID[T any](id string) Option[T] {
	return func(e *Envelope[T]) { e.traceID = id }
}

// [WITH_ECHO] Marks the event as a sender's device synchronization copy.
func WithEcho[T any](v bool) Option[T] {
	return func(e *Envelope[T]) { e.echo = v }
}

// [WITH_TRACKED] Determines if the event should trigger delivery tracking.
func WithTracked[T any](v bool) Option[T] {
	return func(e *Envelope[T]) { e.canPush = v }
}
