package registry

import "time"

type Option func(*Hub)

// WithEvictionInterval configures the frequency of the janitor (reclamation) cycle.
func WithEvictionInterval(d time.Duration) Option {
	return func(h *Hub) { h.evictionInterval = d }
}

// WithIdleTimeout defines the duration after which an inactive cell is reclaimed.
func WithIdleTimeout(d time.Duration) Option {
	return func(h *Hub) { h.idleTimeout = d }
}

// WithMailboxSize sets the buffer capacity for individual user actors.
func WithMailboxSize(size int) Option {
	return func(h *Hub) { h.mailboxSize = size }
}
