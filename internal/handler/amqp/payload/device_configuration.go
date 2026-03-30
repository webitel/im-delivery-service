package payload

// DeviceConfigurationUpdate represents the root structure for device-related events
// (register, unregister, logout) received via AMQP.
type DeviceConfigurationUpdate struct {
	// Authorization contains the session and device details provided by the identity service.
	Authorization *AuthPayload `json:"authorization"`
}

// AuthPayload describes the session authorization details.
type AuthPayload struct {
	// ID is the unique session (authorization) identifier.
	ID string `json:"id"`
	// DC is the data center identifier (received as string to match JSON input).
	DC string `json:"dc"`
	// Date is the creation timestamp of the session (received as string).
	Date string `json:"date"`
	// Name is the display name of the session/device (e.g., "grpc-go/1.78.0").
	Name string `json:"name"`
	// AppID is the client application identifier.
	AppID string `json:"app_id"`
	// Device contains hardware and push notification metadata.
	Device *DevicePayload `json:"device"`
	// Contact contains information about the authorized end-user.
	Contact *ContactPayload `json:"contact"`
}

// DevicePayload describes the physical or virtual device properties.
type DevicePayload struct {
	// ID is the unique device identifier (e.g., "ws-10").
	ID string `json:"id"`
	// IP is the last known IP address of the device.
	IP string `json:"ip"`
	// App contains specific application versioning info.
	App *AppInfo `json:"app"`
	// Push contains push notification tokens (FCM, etc.).
	Push *PushConfig `json:"push"`
}

// AppInfo describes the client application metadata.
type AppInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	String  string `json:"string"`
}

// PushConfig holds tokens for cloud messaging platforms.
type PushConfig struct {
	// FCM is the Firebase Cloud Messaging token.
	FCM string `json:"fcm"`
}

// ContactPayload describes the user account linked to the session.
type ContactPayload struct {
	// ID is the unique contact/user identifier.
	ID string `json:"id"`
	// DC is the data center where the contact is hosted.
	DC string `json:"dc"`
	// Sub is the subject identifier (often the internal user ID).
	Sub string `json:"sub"`
	// Iss is the token issuer (e.g., "webitel").
	Iss string `json:"iss"`
}
