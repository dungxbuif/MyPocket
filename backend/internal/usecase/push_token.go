package usecase

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mypocket/backend/internal/entity"
	"github.com/mypocket/backend/internal/repository"
)

var ErrInvalidPushToken = errors.New("thông tin đăng ký thông báo đẩy không hợp lệ")

type PushTokenInput struct {
	DeviceID    string `json:"device_id"`
	DeviceName  string `json:"device_name"`
	Platform    string `json:"platform"`
	Environment string `json:"environment"`
	Token       string `json:"token"`
}

type PushTokenInteractor struct {
	Tokens repository.PushTokenRepository
}

func NewPushTokenInteractor(tokens repository.PushTokenRepository) *PushTokenInteractor {
	return &PushTokenInteractor{Tokens: tokens}
}

func (s *PushTokenInteractor) Register(ctx context.Context, owner string, in PushTokenInput) (*entity.PushToken, error) {
	in.DeviceID = strings.TrimSpace(in.DeviceID)
	in.DeviceName = strings.TrimSpace(in.DeviceName)
	in.Platform = strings.ToLower(strings.TrimSpace(in.Platform))
	in.Environment = strings.ToLower(strings.TrimSpace(in.Environment))
	in.Token = strings.TrimSpace(in.Token)
	if s == nil || s.Tokens == nil || strings.TrimSpace(owner) == "" || in.DeviceID == "" || len(in.DeviceID) > 128 || len(in.DeviceName) > 100 || in.Platform != "ios" || (in.Environment != "sandbox" && in.Environment != "production") || len(in.Token) < 32 || len(in.Token) > 4096 {
		return nil, ErrInvalidPushToken
	}
	now := time.Now().UTC()
	item := &entity.PushToken{ID: uuid.NewString(), OwnerID: owner, DeviceID: in.DeviceID, DeviceName: in.DeviceName, Platform: in.Platform, Environment: in.Environment, Token: in.Token, CreatedAt: now, UpdatedAt: now}
	if err := s.Tokens.Save(ctx, item); err != nil {
		return nil, err
	}
	// Never return the provider token even when this method's result is encoded.
	item.Token = ""
	return item, nil
}

func (s *PushTokenInteractor) Revoke(ctx context.Context, owner, id string) error {
	if s == nil || s.Tokens == nil || strings.TrimSpace(owner) == "" || strings.TrimSpace(id) == "" {
		return repository.ErrNotFound
	}
	if _, err := uuid.Parse(id); err != nil {
		return repository.ErrNotFound
	}
	return s.Tokens.Revoke(ctx, owner, id)
}
