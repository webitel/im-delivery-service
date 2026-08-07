package webhook

import (
	"sync"
)

// [PROXY_CACHE] Thread-safe global storage for generic providers.
var proxyCache sync.Map

// [GET_OR_CREATE] Type-safe factory for retrieving or initializing proxy clients.
// ---------------------------------------------------------------------------------
// [USAGE]
// - proxy := webhook.GetOrCreate[*messaging.Message](url)
// ---------------------------------------------------------------------------------
func GetOrCreate[T any](url string) *webhookProvider[T] {
	// [1. ATOMIC_LOAD] Check if the provider for this URL already exists.
	if val, ok := proxyCache.Load(url); ok {
		// [WARNING] Ensure T matches the type stored for this URL.
		return val.(*webhookProvider[T])
	}

	// [2. ATOMIC_STORE] LoadOrStore handles potential race conditions.
	actual, _ := proxyCache.LoadOrStore(url, NewWebhookProvider[T](url))

	return actual.(*webhookProvider[T])
}
