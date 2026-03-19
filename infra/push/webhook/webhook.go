package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/webitel/im-delivery-service/infra/push"
	"github.com/webitel/im-delivery-service/internal/domain/model"
)

const (
	Webhook = "webhook"
)

type webhookProvider struct {
	url    string
	client *http.Client
}

// [INTERFACE GUARD]
var _ push.Provider = (*webhookProvider)(nil)

func NewWebhookProvider(url string) push.Provider {
	return &webhookProvider{
		url:    url,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (w *webhookProvider) Name() string { return Webhook }

func (w *webhookProvider) Send(ctx context.Context, req *model.PushRequest) error {
	data, _ := json.Marshal(req)
	hReq, _ := http.NewRequestWithContext(ctx, http.MethodPost, w.url, bytes.NewReader(data))
	hReq.Header.Set("Content-Type", "application/json")

	resp, err := w.client.Do(hReq)
	if err == nil {
		resp.Body.Close()
	}
	return err
}

func (w *webhookProvider) Dismiss(ctx context.Context, req *model.PushRequest) error {
	return w.Send(ctx, req)
}
