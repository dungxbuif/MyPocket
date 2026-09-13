package usecase

import "github.com/mypocket/backend/internal/entity"

type GoogleOAuthProvider interface {
	ExchangeCode(code string) (entity.GoogleProfile, error)
	IsReady() bool
	BuildAuthURL(state string) string
}
