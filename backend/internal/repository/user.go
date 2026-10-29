package repository

import "github.com/mypocket/backend/internal/entity"

type UserRepository interface {
	FindByEmail(email string) (*entity.User, error)
	FindByID(id string) (*entity.User, error)
	FindOrCreateGoogleUser(profile entity.GoogleProfile) (*entity.User, error)
	SetTimezone(userID, timezone string, initializeOnly bool) (*entity.User, error)
	Create(user *entity.User) error
}
