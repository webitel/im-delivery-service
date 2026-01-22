package registry

import "time"

// Option defines a functional configuration type for the Hub.
type Option func(*Hub)

// WithEvictionInterval configures how often the [JANITOR] process runs
// to reclaim memory from inactive users.
func WithEvictionInterval(d time.Duration) Option {
	return func(h *Hub) {
		h.config.evictionInterval = d
	}
}

// WithIdleTimeout defines the [QUIET_PERIOD] after which a user cell
// without active sessions is considered eligible for eviction.
func WithIdleTimeout(d time.Duration) Option {
	return func(h *Hub) {
		h.config.idleTimeout = d
	}
}

// WithMailboxSize sets the [BACKPRESSURE] threshold.
// It defines the buffer capacity for each individual user's actor mailbox.
func WithMailboxSize(size int) Option {
	return func(h *Hub) {
		h.config.mailboxSize = size
	}
}
