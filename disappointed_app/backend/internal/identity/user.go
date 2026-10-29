package identity

import "time"

type User struct {
	ID            string
	GoogleSubject string
	Email         string
	EmailVerified bool
	DisplayName   string
	AvatarURL     string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type GoogleProfile struct {
	Subject       string
	Email         string
	EmailVerified bool
	DisplayName   string
	AvatarURL     string
}
