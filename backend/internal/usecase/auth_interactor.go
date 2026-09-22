package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mypocket/backend/internal/entity"
	"github.com/mypocket/backend/internal/repository"
)

var ErrInvalidCredentials = errors.New("email hoặc mật khẩu không đúng")
var ErrInvalidGoogleProfile = errors.New("thông tin Google không hợp lệ")
var ErrInvalidTimezone = errors.New("múi giờ IANA không hợp lệ")

type AuthInteractor struct {
	UserRepo      repository.UserRepository
	PasswordSvc   PasswordService
	TokenSvc      TokenService
	CacheRepo     repository.CacheRepository
	SessionPrefix string
	GoogleOAuth   GoogleOAuthProvider
}

func NewAuthInteractor(
	userRepo repository.UserRepository,
	passwordSvc PasswordService,
	tokenSvc TokenService,
	cacheRepo repository.CacheRepository,
	googleOAuth GoogleOAuthProvider,
) *AuthInteractor {
	return &AuthInteractor{
		UserRepo:      userRepo,
		PasswordSvc:   passwordSvc,
		TokenSvc:      tokenSvc,
		CacheRepo:     cacheRepo,
		SessionPrefix: sessionCachePrefix,
		GoogleOAuth:   googleOAuth,
	}
}

func (a *AuthInteractor) Login(ctx context.Context, input LoginInput) (*LoginOutput, error) {
	_ = ctx
	user, err := a.UserRepo.FindByEmail(input.Email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}
	if !a.PasswordSvc.Check(input.Password, user.PasswordHash) {
		return nil, ErrInvalidCredentials
	}
	return a.buildLoginOutput(ctx, user)
}

func (a *AuthInteractor) LoginWithGoogle(ctx context.Context, profile entity.GoogleProfile) (*LoginOutput, error) {
	_ = ctx
	if profile.Subject == "" || profile.Email == "" {
		return nil, ErrInvalidGoogleProfile
	}
	user, err := a.UserRepo.FindOrCreateGoogleUser(profile)
	if err != nil {
		return nil, ErrInvalidGoogleProfile
	}
	return a.buildLoginOutput(ctx, user)
}

func (a *AuthInteractor) FetchGoogleProfile(code string) (entity.GoogleProfile, error) {
	if a.GoogleOAuth == nil || !a.GoogleOAuth.IsReady() {
		return entity.GoogleProfile{}, ErrInvalidGoogleProfile
	}
	if strings.TrimSpace(code) == "" {
		return entity.GoogleProfile{}, ErrInvalidGoogleProfile
	}
	return a.GoogleOAuth.ExchangeCode(code)
}

func (a *AuthInteractor) LoginWithGoogleCode(ctx context.Context, code string) (*LoginOutput, error) {
	_ = ctx
	if a.GoogleOAuth == nil || !a.GoogleOAuth.IsReady() {
		return nil, ErrInvalidGoogleProfile
	}
	profile, err := a.GoogleOAuth.ExchangeCode(code)
	if err != nil {
		return nil, ErrInvalidGoogleProfile
	}
	return a.LoginWithGoogle(ctx, profile)
}

func (a *AuthInteractor) buildLoginOutput(ctx context.Context, user *entity.User) (*LoginOutput, error) {
	_ = ctx
	sessionID := uuid.NewString()
	token, exp, err := a.TokenSvc.Generate(user.ID, user.Email, user.Name, sessionID)
	if err != nil {
		return nil, err
	}
	key := a.sessionKey(sessionID)
	if err := a.CacheRepo.Set(key, user.ID, time.Until(exp)); err != nil {
		return nil, fmt.Errorf("set session cache: %w", err)
	}
	return &LoginOutput{
		Token:     token,
		ExpiresAt: exp,
		User: UserProfile{
			ID:                user.ID,
			Name:              user.Name,
			Email:             user.Email,
			Timezone:          user.Timezone,
			TimezoneConfirmed: user.TimezoneConfirmed,
			CreatedAt:         user.CreatedAt,
		},
	}, nil
}

func (a *AuthInteractor) Profile(ctx context.Context, userID string) (*UserProfile, error) {
	_ = ctx
	cacheKey := fmt.Sprintf("%s%s", profileCachePrefix, userID)
	if a.CacheRepo != nil {
		if cached, ok, err := a.CacheRepo.Get(cacheKey); err == nil && ok {
			var profile UserProfile
			if err := json.Unmarshal([]byte(cached), &profile); err == nil && profile.Timezone != "" {
				return &profile, nil
			}
		}
	}
	user, err := a.UserRepo.FindByID(userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}
	profile := &UserProfile{
		ID:                user.ID,
		Name:              user.Name,
		Email:             user.Email,
		Timezone:          user.Timezone,
		TimezoneConfirmed: user.TimezoneConfirmed,
		CreatedAt:         user.CreatedAt,
	}
	if a.CacheRepo != nil {
		raw, _ := json.Marshal(profile)
		_ = a.CacheRepo.Set(cacheKey, string(raw), cacheProfileTTL*time.Second)
	}
	return profile, nil
}

func (a *AuthInteractor) SetTimezone(ctx context.Context, userID, timezone string, initializeOnly bool) (*UserProfile, error) {
	_ = ctx
	timezone = strings.TrimSpace(timezone)
	location, err := time.LoadLocation(timezone)
	if err != nil || timezone == "" || timezone == "Local" {
		return nil, ErrInvalidTimezone
	}
	user, err := a.UserRepo.SetTimezone(userID, location.String(), initializeOnly)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}
	profile := &UserProfile{ID: user.ID, Name: user.Name, Email: user.Email, Timezone: user.Timezone, TimezoneConfirmed: user.TimezoneConfirmed, CreatedAt: user.CreatedAt}
	if a.CacheRepo != nil {
		_ = a.CacheRepo.Delete(fmt.Sprintf("%s%s", profileCachePrefix, userID))
		_ = a.CacheRepo.Delete(fmt.Sprintf("%s%s", homeCachePrefix, userID))
	}
	return profile, nil
}

func (a *AuthInteractor) Home(ctx context.Context, userID string) (*HomeOutput, error) {
	_ = ctx
	cacheKey := fmt.Sprintf("%s%s", homeCachePrefix, userID)
	if cached, ok, err := a.CacheRepo.Get(cacheKey); err == nil && ok {
		var home HomeOutput
		if err := json.Unmarshal([]byte(cached), &home); err == nil {
			return &home, nil
		}
	}
	user, err := a.UserRepo.FindByID(userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}
	home := &HomeOutput{
		Greeting: "Xin chào " + user.Name,
		Stats: struct {
			WalletCount int `json:"wallet_count"`
			NotesCount  int `json:"notes_count"`
		}{
			WalletCount: 0,
			NotesCount:  0,
		},
		Notice: "Chào mừng bạn đến với MyPocket. Hiện tại chưa có dữ liệu giao dịch.",
	}
	raw, err := json.Marshal(home)
	if err == nil {
		_ = a.CacheRepo.Set(cacheKey, string(raw), cacheHomeTTL*time.Second)
	}
	return home, nil
}

func (a *AuthInteractor) sessionKey(sessionID string) string {
	return a.SessionPrefix + sessionID
}

func (a *AuthInteractor) VerifySession(ctx context.Context, sessionID string) (string, error) {
	key := a.sessionKey(sessionID)
	userID, ok, err := a.CacheRepo.Get(key)
	if err != nil {
		return "", err
	}
	if !ok || userID == "" {
		return "", ErrInvalidCredentials
	}
	return userID, nil
}

func (a *AuthInteractor) IsGoogleAuthReady() bool {
	return a.GoogleOAuth != nil && a.GoogleOAuth.IsReady()
}

func (a *AuthInteractor) GoogleAuthURL(state string) (string, bool) {
	if a.GoogleOAuth == nil {
		return "", false
	}
	return a.GoogleOAuth.BuildAuthURL(state), true
}
