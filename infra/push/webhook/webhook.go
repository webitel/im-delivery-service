package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"
)

const Name = "webhook"

// [SHARED_HTTP_CLIENT] Reuse connections to avoid socket exhaustion.
var sharedClient = &http.Client{
	Timeout: 10 * time.Second,
}

// [WEBHOOK_PROVIDER] Generic HTTP transport for push delegation.
// ---------------------------------------------------------------------------------
// [LOGIC]
// - T is a generic native payload (e.g., *messaging.Message or *apns2.Notification).
// - Dispatches JSON to a specified URL via POST.
// ---------------------------------------------------------------------------------
type webhookProvider[T any] struct {
	url    string
	client *http.Client
}

func NewWebhookProvider[T any](url string) *webhookProvider[T] {
	return &webhookProvider[T]{
		url:    url,
		client: sharedClient,
	}
}

func (w *webhookProvider[T]) Name() string { return Name }

// [SEND] Serializes the generic payload and executes the HTTP request.
func (w *webhookProvider[T]) Send(ctx context.Context, payload T) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	hReq, err := http.NewRequestWithContext(ctx, http.MethodPost, w.url, bytes.NewReader(data))
	if err != nil {
		return err
	}

	hReq.Header.Set("Content-Type", "application/json")
	hReq.Header.Set("X-Debug-Proxy", "im-delivery-service")

	resp, err := w.client.Do(hReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}
