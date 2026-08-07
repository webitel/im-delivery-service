package model

// [PUSH_CONFIG] Dynamic routing and authentication settings per Application.
type PushConfig struct {
	Proxy       string `json:"proxy,omitempty"`
	Credentials []byte `json:"credentials,omitempty"` // Service Account (FCM) or .p8 (APNs)
	Topic       string `json:"topic,omitempty"`       // Bundle ID
	// [APNS_SPECIFIC]
	KeyID  string `json:"key_id,omitempty"`  // Apple Key ID (10 chars)
	TeamID string `json:"team_id,omitempty"` // Apple Team ID (10 chars)
	// Proto is the APNs proxy transport: "h2" (default) or "http/1.1". Only
	// meaningful for a custom proxy endpoint.
	Proto string `json:"proto,omitempty"`
}

// [DEVICE] Represents a single notification target with its specific app configuration.
type Device struct {
	ID         string     `json:"id"`
	Platform   string     `json:"platform"`
	AppID      string     `json:"app_id"`
	PushType   string     `json:"push_type"` // fcm, apn, web
	PushToken  string     `json:"push_token"`
	PushConfig PushConfig `json:"push_config"`
}

const (
	PlatformAndroid = "android"
	PlatformIOS     = "ios"
	PlatformWeb     = "web"
)
