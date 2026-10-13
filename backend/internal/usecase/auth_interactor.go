package usecase

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
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
var ErrInvalidRefreshToken = errors.New("refresh token không hợp lệ hoặc đã hết hạn")
var ErrInvalidTimezone = errors.New("múi giờ IANA không hợp lệ")

type AuthInteractor struct {
	UserRepo          repository.UserRepository
	PasswordSvc       PasswordService
	TokenSvc          TokenService
	CacheRepo         repository.CacheRepository
	SessionPrefix     string
	GoogleOAuth       GoogleOAuthProvider
	MobileGoogleOAuth MobileGoogleOAuthProvider
	RefreshTTL        time.Duration
	Devices           repository.DeviceCredentialRepository
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
		RefreshTTL:    30 * 24 * time.Hour,
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

func (a *AuthInteractor) LoginWithMobileGoogleCode(ctx context.Context, code, verifier, redirectURI string) (*LoginOutput, error) {
	if a.MobileGoogleOAuth == nil || !a.MobileGoogleOAuth.IsReady() {
		return nil, ErrInvalidGoogleProfile
	}
	profile, err := a.MobileGoogleOAuth.ExchangeCodeWithPKCE(strings.TrimSpace(code), strings.TrimSpace(verifier), strings.TrimSpace(redirectURI))
	if err != nil || !profile.EmailVerified {
		return nil, ErrInvalidGoogleProfile
	}
	return a.LoginWithGoogle(ctx, profile)
}

func (a *AuthInteractor) Refresh(ctx context.Context, rawToken string) (*LoginOutput, error) {
	rawToken = strings.TrimSpace(rawToken)
	if rawToken == "" || a.CacheRepo == nil {
		return nil, ErrInvalidRefreshToken
	}
	key := refreshTokenKey(rawToken)
	value, ok, err := consumeRefreshToken(a.CacheRepo, key)
	if err != nil || !ok {
		return nil, ErrInvalidRefreshToken
	}
	var record refreshRecord
	if json.Unmarshal([]byte(value), &record) != nil || strings.TrimSpace(record.UserID) == "" {
		return nil, ErrInvalidRefreshToken
	}
	if record.FamilyID == "" {
		record.FamilyID = legacySessionFamily(record.UserID)
	}
	if record.FamilyID != "" {
		if _, revoked, cacheErr := a.CacheRepo.Get(familyRevocationKey(record.FamilyID)); cacheErr != nil || revoked {
			return nil, ErrInvalidRefreshToken
		}
	}
	user, err := a.UserRepo.FindByID(record.UserID)
	if err != nil {
		return nil, ErrInvalidRefreshToken
	}
	return a.buildSessionOutput(ctx, user, record.FamilyID)
}

func (a *AuthInteractor) buildLoginOutput(ctx context.Context, user *entity.User) (*LoginOutput, error) {
	return a.buildSessionOutput(ctx, user, "")
}

func (a *AuthInteractor) buildSessionOutput(ctx context.Context, user *entity.User, familyID string) (*LoginOutput, error) {
	_ = ctx
	sessionID := uuid.NewString()
	if familyID == "" {
		familyID = uuid.NewString()
	}
	token, exp, err := a.TokenSvc.Generate(user.ID, user.Email, user.Name, sessionID)
	if err != nil {
		return nil, err
	}
	key := a.sessionKey(sessionID)
	record, err := json.Marshal(sessionRecord{UserID: user.ID, FamilyID: familyID, ExpiresAt: exp})
	if err != nil {
		return nil, err
	}
	if err := a.CacheRepo.Set(key, string(record), time.Until(exp)); err != nil {
		return nil, fmt.Errorf("set session cache: %w", err)
	}
	refreshToken, refreshExp, err := a.issueRefreshToken(user.ID, familyID)
	if err != nil {
		return nil, fmt.Errorf("set refresh token: %w", err)
	}
	// Issuance may overlap logout after Refresh's initial check. Do not return
	// credentials minted after its revocation marker. If logout occurs after
	// this final check, its TTL starts later than these credentials' TTLs.
	_, revoked, cacheErr := a.CacheRepo.Get(familyRevocationKey(familyID))
	if cacheErr != nil || revoked {
		_ = a.CacheRepo.Delete(key)
		_ = a.CacheRepo.Delete(refreshTokenKey(refreshToken))
		if cacheErr != nil {
			return nil, fmt.Errorf("check session revocation: %w", cacheErr)
		}
		return nil, ErrInvalidRefreshToken
	}
	return &LoginOutput{
		Token:            token,
		ExpiresAt:        exp,
		RefreshToken:     refreshToken,
		RefreshExpiresAt: refreshExp,
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

func (a *AuthInteractor) issueRefreshToken(userID, familyID string) (string, time.Time, error) {
	refreshTTL := a.RefreshTTL
	if refreshTTL <= 0 {
		refreshTTL = 30 * 24 * time.Hour
	}
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", time.Time{}, err
	}
	token := "mpr_" + base64.RawURLEncoding.EncodeToString(raw)
	expiresAt := time.Now().Add(refreshTTL)
	record, err := json.Marshal(refreshRecord{UserID: userID, FamilyID: familyID})
	if err != nil {
		return "", time.Time{}, err
	}
	if err := a.CacheRepo.Set(refreshTokenKey(token), string(record), refreshTTL); err != nil {
		return "", time.Time{}, err
	}
	return token, expiresAt, nil
}

func refreshTokenKey(rawToken string) string {
	digest := sha256.Sum256([]byte(rawToken))
	return refreshTokenPrefix + fmt.Sprintf("%x", digest[:])
}

func consumeRefreshToken(cache repository.CacheRepository, key string) (string, bool, error) {
	if atomic, ok := cache.(repository.AtomicCacheRepository); ok {
		return atomic.GetAndDelete(key)
	}
	value, found, err := cache.Get(key)
	if err != nil || !found {
		return value, found, err
	}
	if err := cache.Delete(key); err != nil {
		return "", false, err
	}
	return value, true, nil
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
	value, ok, err := a.CacheRepo.Get(key)
	if err != nil {
		return "", err
	}
	if !ok || value == "" {
		return "", ErrInvalidCredentials
	}
	record, err := decodeSessionRecord(value)
	if err != nil || record.UserID == "" {
		return "", ErrInvalidCredentials
	}
	if record.FamilyID != "" {
		if _, revoked, cacheErr := a.CacheRepo.Get(familyRevocationKey(record.FamilyID)); cacheErr != nil {
			return "", cacheErr
		} else if revoked {
			return "", ErrInvalidCredentials
		}
	}
	return record.UserID, nil
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
