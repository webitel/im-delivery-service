package model

type AuthSession struct {
	DC        int64
	ContactID string
	Sub       string
	Iss       string
	Name      string
}
