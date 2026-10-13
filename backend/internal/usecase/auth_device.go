package usecase

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/mypocket/backend/internal/entity"
	"github.com/mypocket/backend/internal/repository"
	"strings"
	"time"
)

var ErrInvalidDeviceCredential = errors.New("đăng nhập thiết bị đã hết hạn hoặc bị thu hồi; hãy đăng nhập bằng mật khẩu")
var ErrInvalidDeviceEnrollment = errors.New("thông tin đăng ký thiết bị không hợp lệ")

type DeviceEnrollmentInput struct {
	DeviceID   string `json:"device_id"`
	DeviceName string `json:"device_name"`
	Password   string `json:"password"`
}
type DeviceCredentialOutput struct {
	ID        string    `json:"id"`
	Secret    string    `json:"secret"`
	ExpiresAt time.Time `json:"expires_at"`
}
type refreshRecord struct {
	UserID   string `json:"user_id"`
	FamilyID string `json:"family_id,omitempty"`
}
type sessionRecord struct {
	UserID    string    `json:"user_id"`
	FamilyID  string    `json:"family_id"`
	ExpiresAt time.Time `json:"expires_at"`
}

func credentialDigest(secret string) string {
	digest := sha256.Sum256([]byte(secret))
	return fmt.Sprintf("%x", digest[:])
}
func familyRevocationKey(id string) string { return "auth:revoked-family:" + id }

// Old records contain only user ID, so their original device/family cannot be
// recovered. Keep a stable migration group rather than creating an unrelated
// family on refresh that could escape a concurrent logout.
func legacySessionFamily(owner string) string { return "legacy:" + credentialDigest(owner) }

func (a *AuthInteractor) EnrollDevice(ctx context.Context, owner string, in DeviceEnrollmentInput) (*DeviceCredentialOutput, error) {
	in.DeviceID = strings.TrimSpace(in.DeviceID)
	in.DeviceName = strings.TrimSpace(in.DeviceName)
	if a.Devices == nil {
		return nil, errors.New("device credential repository unavailable")
	}
	if owner == "" || in.DeviceID == "" || len(in.DeviceID) > 128 || len(in.DeviceName) > 100 || in.Password == "" {
		return nil, ErrInvalidDeviceEnrollment
	}
	user, err := a.UserRepo.FindByID(owner)
	if err != nil {
		return nil, err
	}
	if !a.PasswordSvc.Check(in.Password, user.PasswordHash) {
		return nil, ErrInvalidCredentials
	}
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return nil, err
	}
	secret := "mpd_" + base64.RawURLEncoding.EncodeToString(raw)
	now := time.Now().UTC()
	device := &entity.DeviceCredential{ID: uuid.NewString(), OwnerID: owner, DeviceID: in.DeviceID, DeviceName: in.DeviceName, SecretHash: credentialDigest(secret), PasswordFingerprint: credentialDigest(user.PasswordHash), ExpiresAt: now.Add(180 * 24 * time.Hour), CreatedAt: now, UpdatedAt: now}
	if err := a.Devices.Save(ctx, device); err != nil {
		return nil, err
	}
	return &DeviceCredentialOutput{ID: device.ID, Secret: secret, ExpiresAt: device.ExpiresAt}, nil
}
func (a *AuthInteractor) LoginWithDevice(ctx context.Context, id, secret string) (*LoginOutput, error) {
	if a.Devices == nil {
		return nil, errors.New("device credential repository unavailable")
	}
	if strings.TrimSpace(id) == "" || len(id) > 128 || !strings.HasPrefix(secret, "mpd_") || len(secret) > 128 {
		return nil, ErrInvalidDeviceCredential
	}
	device, err := a.Devices.FindByID(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrInvalidDeviceCredential
	}
	if err != nil {
		return nil, err
	}
	if device.RevokedAt != nil || !time.Now().Before(device.ExpiresAt) || subtle.ConstantTimeCompare([]byte(device.SecretHash), []byte(credentialDigest(secret))) != 1 {
		return nil, ErrInvalidDeviceCredential
	}
	user, err := a.UserRepo.FindByID(device.OwnerID)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrInvalidDeviceCredential
	}
	if err != nil {
		return nil, err
	}
	if subtle.ConstantTimeCompare([]byte(device.PasswordFingerprint), []byte(credentialDigest(user.PasswordHash))) != 1 {
		return nil, ErrInvalidDeviceCredential
	}
	return a.buildLoginOutput(ctx, user)
}
func (a *AuthInteractor) RevokeDevice(ctx context.Context, owner, id string) error {
	if a.Devices == nil {
		return errors.New("device credential repository unavailable")
	}
	return a.Devices.Revoke(ctx, owner, id)
}

func decodeSessionRecord(value string) (sessionRecord, error) {
	if !strings.HasPrefix(value, "{") {
		return sessionRecord{UserID: value, FamilyID: legacySessionFamily(value)}, nil
	}
	var record sessionRecord
	if err := json.Unmarshal([]byte(value), &record); err != nil {
		return record, err
	}
	return record, nil
}

func (a *AuthInteractor) Logout(ctx context.Context, owner, sessionID, rawRefresh string) error {
	_ = ctx
	value, ok, err := a.CacheRepo.Get(a.sessionKey(sessionID))
	if err != nil {
		return err
	}
	if !ok {
		return ErrInvalidCredentials
	}
	session, err := decodeSessionRecord(value)
	if err != nil || session.UserID != owner || owner == "" {
		return ErrInvalidCredentials
	}
	if rawRefresh != "" {
		raw, found, err := a.CacheRepo.Get(refreshTokenKey(rawRefresh))
		if err != nil {
			return err
		}
		if found {
			var record refreshRecord
			if json.Unmarshal([]byte(raw), &record) != nil || record.UserID != owner || (record.FamilyID != "" && record.FamilyID != session.FamilyID) {
				return ErrInvalidRefreshToken
			}
		}
	}
	if session.FamilyID != "" {
		ttl := a.RefreshTTL
		if ttl <= 0 {
			ttl = 30 * 24 * time.Hour
		}
		if remaining := time.Until(session.ExpiresAt); remaining > ttl {
			ttl = remaining
		}
		if err := a.CacheRepo.Set(familyRevocationKey(session.FamilyID), "revoked", ttl); err != nil {
			return err
		}
	}
	if err := a.CacheRepo.Delete(a.sessionKey(sessionID)); err != nil {
		return err
	}
	if rawRefresh != "" {
		return a.CacheRepo.Delete(refreshTokenKey(rawRefresh))
	}
	return nil
}
