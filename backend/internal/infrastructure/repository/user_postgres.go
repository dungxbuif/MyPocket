package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mypocket/backend/internal/entity"
	repository "github.com/mypocket/backend/internal/repository"
	"gorm.io/gorm"
)

type UserPostgresRepository struct {
	db *gorm.DB
}

func NewUserPostgresRepository(db *gorm.DB) repository.UserRepository {
	return &UserPostgresRepository{db: db}
}

func (r *UserPostgresRepository) FindByEmail(email string) (*entity.User, error) {
	var user entity.User
	err := r.db.Where(UserColumnEmail, email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserPostgresRepository) FindByID(id string) (*entity.User, error) {
	var user entity.User
	err := r.db.Where(UserColumnID, id).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserPostgresRepository) FindOrCreateGoogleUser(profile entity.GoogleProfile) (*entity.User, error) {
	if strings.TrimSpace(profile.Subject) == "" {
		return nil, repository.ErrInvalidGoogleProfile
	}
	if strings.TrimSpace(profile.Email) == "" {
		return nil, repository.ErrInvalidGoogleProfile
	}

	ctx := context.Background()
	var user entity.User
	if err := r.db.WithContext(ctx).Where(UserColumnGoogleSubject, profile.Subject).First(&user).Error; err == nil {
		user.Email = profile.Email
		user.EmailVerified = profile.EmailVerified
		user.Name = coalesce(profile.DisplayName, user.Name)
		user.AvatarURL = profile.AvatarURL
		if err := r.db.WithContext(ctx).Save(&user).Error; err != nil {
			return nil, err
		}
		return &user, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if err := r.db.WithContext(ctx).Where(UserColumnEmail, strings.ToLower(strings.TrimSpace(profile.Email))).First(&user).Error; err == nil {
		user.GoogleSubject = profile.Subject
		user.EmailVerified = profile.EmailVerified
		if strings.TrimSpace(profile.DisplayName) != "" {
			user.Name = profile.DisplayName
		}
		if profile.AvatarURL != "" {
			user.AvatarURL = profile.AvatarURL
		}
		if err := r.db.WithContext(ctx).Save(&user).Error; err != nil {
			return nil, err
		}
		return &user, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	now := time.Now().UTC()
	user = entity.User{
		ID:            uuid.NewString(),
		GoogleSubject: profile.Subject,
		Name:          coalesce(profile.DisplayName, "Google User"),
		Email:         strings.ToLower(strings.TrimSpace(profile.Email)),
		EmailVerified: profile.EmailVerified,
		AvatarURL:     profile.AvatarURL,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := r.db.WithContext(ctx).Create(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserPostgresRepository) Create(user *entity.User) error {
	return r.db.Create(user).Error
}

func (r *UserPostgresRepository) SetTimezone(userID, timezone string, initializeOnly bool) (*entity.User, error) {
	query := r.db.Model(&entity.User{}).Where("id = ?", userID)
	if initializeOnly {
		query = query.Where("timezone_confirmed = false")
	}
	result := query.Updates(map[string]any{"timezone": timezone, "timezone_confirmed": true, "updated_at": time.Now().UTC()})
	if result.Error != nil {
		return nil, result.Error
	}
	return r.FindByID(userID)
}

const (
	UserColumnEmail         = "email"
	UserColumnID            = "id"
	UserColumnGoogleSubject = "google_subject"
)

func coalesce(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value != "" {
		return value
	}
	return fallback
}
