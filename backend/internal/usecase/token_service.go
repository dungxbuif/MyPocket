package usecase

import (
	"errors"
	"time"
)

type TokenClaims struct {
	UserID    string
	Email     string
	Name      string
	SessionID string
	ExpiresAt time.Time
}

var ErrInvalidToken = errors.New("invalid token")

type TokenService interface {
	Generate(userID, email, name, sessionID string) (string, time.Time, error)
	Verify(token string) (*TokenClaims, error)
}
