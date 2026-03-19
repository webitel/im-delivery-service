package model

// [PUSH_CONFIG] Dynamic routing and authentication settings per Application.
type PushConfig struct {
	Proxy       string `json:"proxy,omitempty"`       // Custom gateway URL (e.g., internal FCM proxy)
	Credentials []byte `json:"credentials,omitempty"` // Service Account JSON (FCM) or .p8 Key (APNs)
	Topic       string `json:"topic,omitempty"`       // Bundle ID / Topic for APNs
}

// [DEVICE] Represents a single notification target with its specific app configuration.
type Device struct {
	ID         string     `json:"id"`
	Platform   string     `json:"platform"`
	AppID      string     `json:"app_id"`    // [CLIENT_ID] Used to lookup Application config
	PushType   string     `json:"push_type"` // fcm, apn, web
	PushToken  string     `json:"push_token"`
	PushConfig PushConfig `json:"push_config"` // [DYNAMIC] Injected from SearchApps
}

const (
	PlatformAndroid = "android"
	PlatformIOS     = "ios"
	PlatformWeb     = "web"
)
