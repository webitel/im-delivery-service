// internal/infra/push/webhook/factory.go
package webhook

import (
	"sync"
)

// [CACHE] Thread-safe storage for initialized proxy providers.
var proxyCache sync.Map

// [GET_OR_CREATE] Returns an existing provider from cache or initializes a new one.
func GetOrCreate(url string) *webhookProvider {
	if val, ok := proxyCache.Load(url); ok {
		return val.(*webhookProvider)
	}

	// [INITIALIZE] Create a new instance if not found.
	newProvider := NewWebhookProvider(url).(*webhookProvider)

	// [STORE] Atomic update to ensure only one instance per URL exists.
	actual, _ := proxyCache.LoadOrStore(url, newProvider)
	return actual.(*webhookProvider)
}
