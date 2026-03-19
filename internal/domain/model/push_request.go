package model

type PushRequest struct {
	UserID   string
	Devices  []Device
	Title    string
	Body     string
	Data     map[string]string
	IsSilent bool
	// [COLLAPSE_ID] Used by APNS (apns-collapse-id) and FCM (collapse_key)
	// to identify and replace/remove specific notifications.
	CollapseID string
}
