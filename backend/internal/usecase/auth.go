package usecase

import "time"

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserProfile struct {
	ID                string    `json:"id"`
	Name              string    `json:"name"`
	Email             string    `json:"email"`
	Timezone          string    `json:"timezone"`
	TimezoneConfirmed bool      `json:"timezone_confirmed"`
	CreatedAt         time.Time `json:"created_at"`
}

type LoginOutput struct {
	Token     string      `json:"token"`
	ExpiresAt time.Time   `json:"expires_at"`
	User      UserProfile `json:"user"`
}

type HomeOutput struct {
	Greeting string `json:"greeting"`
	Stats    struct {
		WalletCount int `json:"wallet_count"`
		NotesCount  int `json:"notes_count"`
	} `json:"stats"`
	Notice string `json:"notice"`
}
