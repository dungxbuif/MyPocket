package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/mypocket/backend/internal/entity"
	"github.com/mypocket/backend/internal/repository"
)

type deviceRepoStub struct {
	records map[string]entity.DeviceCredential
}

func (r *deviceRepoStub) Save(_ context.Context, d *entity.DeviceCredential) error {
	if r.records == nil {
		r.records = map[string]entity.DeviceCredential{}
	}
	for id, prior := range r.records {
		if prior.OwnerID == d.OwnerID && prior.DeviceID == d.DeviceID {
			d.ID = id
		}
	}
	r.records[d.ID] = *d
	return nil
}
func (r *deviceRepoStub) FindByID(_ context.Context, id string) (*entity.DeviceCredential, error) {
	d, ok := r.records[id]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return &d, nil
}
func (r *deviceRepoStub) Revoke(_ context.Context, owner, id string) error {
	d, ok := r.records[id]
	if !ok || d.OwnerID != owner {
		return repository.ErrNotFound
	}
	now := time.Now()
	d.RevokedAt = &now
	r.records[id] = d
	return nil
}

type enrollmentPasswordStub struct{}

func (enrollmentPasswordStub) Hash(string) (string, error)   { return "hash", nil }
func (enrollmentPasswordStub) Check(password, _ string) bool { return password == "correct-password" }

func deviceAuthService() (*AuthInteractor, *deviceRepoStub, *entity.User) {
	user := &entity.User{ID: "owner", Email: "owner@example.test", Name: "Owner", PasswordHash: "original-password-hash"}
	devices := &deviceRepoStub{}
	svc := NewAuthInteractor(authUserStub{user: user}, enrollmentPasswordStub{}, &authTokenStub{}, &authCacheStub{}, nil)
	svc.Devices = devices
	return svc, devices, user
}
func enrollDevice(t *testing.T, svc *AuthInteractor) *DeviceCredentialOutput {
	t.Helper()
	result, err := svc.EnrollDevice(context.Background(), "owner", DeviceEnrollmentInput{DeviceID: "installation-1", DeviceName: "Test iPhone", Password: "correct-password"})
	if err != nil {
		t.Fatal(err)
	}
	return result
}
func TestDeviceEnrollmentRequiresPasswordAndStoresOnlyDigest(t *testing.T) {
	svc, repo, _ := deviceAuthService()
	if _, err := svc.EnrollDevice(context.Background(), "owner", DeviceEnrollmentInput{DeviceID: "installation-1", Password: "wrong"}); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("wrong password accepted: %v", err)
	}
	if len(repo.records) != 0 {
		t.Fatal("failed enrollment persisted a device")
	}
	out := enrollDevice(t, svc)
	record := repo.records[out.ID]
	if !strings.HasPrefix(out.Secret, "mpd_") || len(record.SecretHash) != 64 || strings.Contains(record.SecretHash, out.Secret) {
		t.Fatal("device credential was not protected by a digest")
	}
	if time.Until(out.ExpiresAt) < 179*24*time.Hour {
		t.Fatal("device credential incorrectly shares refresh-token expiry")
	}
	if _, err := svc.LoginWithDevice(context.Background(), out.ID, out.Secret); err != nil {
		t.Fatalf("enrolled device cannot sign in: %v", err)
	}
	replacement := enrollDevice(t, svc)
	if replacement.ID != out.ID || replacement.Secret == out.Secret {
		t.Fatal("re-enrollment must replace the existing device secret")
	}
	if _, err := svc.LoginWithDevice(context.Background(), out.ID, out.Secret); !errors.Is(err, ErrInvalidDeviceCredential) {
		t.Fatal("replaced secret accepted")
	}
}
func TestDeviceLoginRejectsWrongSecretExpiryPasswordChangeAndRevocation(t *testing.T) {
	for _, scenario := range []string{"wrong-secret", "expired", "password-changed", "revoked", "unknown-device"} {
		t.Run(scenario, func(t *testing.T) {
			svc, repo, user := deviceAuthService()
			out := enrollDevice(t, svc)
			id, secret := out.ID, out.Secret
			record := repo.records[id]
			switch scenario {
			case "wrong-secret":
				secret = "mpd_invalid"
			case "expired":
				record.ExpiresAt = time.Now().Add(-time.Minute)
				repo.records[id] = record
			case "password-changed":
				user.PasswordHash = "new-password-hash"
			case "revoked":
				if err := svc.RevokeDevice(context.Background(), "owner", id); err != nil {
					t.Fatal(err)
				}
			case "unknown-device":
				id = "unknown"
			}
			if _, err := svc.LoginWithDevice(context.Background(), id, secret); !errors.Is(err, ErrInvalidDeviceCredential) {
				t.Fatalf("%s accepted: %v", scenario, err)
			}
		})
	}
}
func TestDeviceRevocationIsOwnerScoped(t *testing.T) {
	svc, _, _ := deviceAuthService()
	out := enrollDevice(t, svc)
	if err := svc.RevokeDevice(context.Background(), "other-owner", out.ID); !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("cross-owner revoke accepted: %v", err)
	}
	if _, err := svc.LoginWithDevice(context.Background(), out.ID, out.Secret); err != nil {
		t.Fatal("cross-owner revoke damaged owner's device")
	}
}
func TestLogoutRevokesRotatedSessionFamilyButKeepsDeviceEnrollment(t *testing.T) {
	svc, _, _ := deviceAuthService()
	device := enrollDevice(t, svc)
	first, err := svc.LoginWithDevice(context.Background(), device.ID, device.Secret)
	if err != nil {
		t.Fatal(err)
	}
	next, err := svc.Refresh(context.Background(), first.RefreshToken)
	if err != nil {
		t.Fatal(err)
	}
	sid := strings.TrimPrefix(next.Token, "access-")
	if err := svc.Logout(context.Background(), "owner", sid, next.RefreshToken); err != nil {
		t.Fatal(err)
	}
	for _, token := range []string{first.Token, next.Token} {
		if _, err := svc.VerifySession(context.Background(), strings.TrimPrefix(token, "access-")); !errors.Is(err, ErrInvalidCredentials) {
			t.Fatal("logged-out family still has usable access")
		}
	}
	if _, err := svc.Refresh(context.Background(), next.RefreshToken); !errors.Is(err, ErrInvalidRefreshToken) {
		t.Fatal("logged-out refresh token still accepted")
	}
	again, err := svc.LoginWithDevice(context.Background(), device.ID, device.Secret)
	if err != nil {
		t.Fatalf("logout removed remembered sign-in: %v", err)
	}
	if _, err := svc.VerifySession(context.Background(), strings.TrimPrefix(again.Token, "access-")); err != nil {
		t.Fatal("fresh device session incorrectly shares revoked family")
	}
}
func TestLogoutCannotRevokeAnotherOwnerOrAnotherSessionFamily(t *testing.T) {
	svc, _, _ := deviceAuthService()
	device := enrollDevice(t, svc)
	first, _ := svc.LoginWithDevice(context.Background(), device.ID, device.Secret)
	second, _ := svc.LoginWithDevice(context.Background(), device.ID, device.Secret)
	sid := strings.TrimPrefix(first.Token, "access-")
	if err := svc.Logout(context.Background(), "other-owner", sid, first.RefreshToken); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatal("cross-owner logout accepted")
	}
	if err := svc.Logout(context.Background(), "owner", sid, second.RefreshToken); !errors.Is(err, ErrInvalidRefreshToken) {
		t.Fatal("unrelated refresh family accepted")
	}
	if _, err := svc.Refresh(context.Background(), second.RefreshToken); err != nil {
		t.Fatal("unrelated family was destroyed")
	}
}

func TestRefreshRotationDoesNotExtendFamilyExpiry(t *testing.T) {

	svc, _, _ := deviceAuthService()
	svc.RefreshTTL = 40 * time.Millisecond
	first, err := svc.buildLoginOutput(context.Background(), &entity.User{ID: "owner"})
	if err != nil {
		t.Fatal(err)
	}
	rotated, err := svc.Refresh(context.Background(), first.RefreshToken)
	if err != nil {
		t.Fatal(err)
	}
	// JSON round-tripping the cache record drops Go's monotonic component; a
	// sub-millisecond wall-clock difference is expected, but rotation must not
	// materially extend the family deadline.
	if rotated.RefreshExpiresAt.After(first.RefreshExpiresAt.Add(time.Millisecond)) {
		t.Fatalf("refresh rotation extended family expiry: first=%s rotated=%s", first.RefreshExpiresAt, rotated.RefreshExpiresAt)
	}
}

func TestLogoutTombstoneCoversAbsoluteFamilyExpiry(t *testing.T) {
	svc, _, _ := deviceAuthService()
	cache := &revocationTTLCache{now: time.Now(), values: map[string]string{}, expires: map[string]time.Time{}}
	svc.CacheRepo = cache
	first, err := svc.buildLoginOutput(context.Background(), &entity.User{ID: "owner"})
	if err != nil {
		t.Fatal(err)
	}
	sessionID := strings.TrimPrefix(first.Token, "access-")
	var session sessionRecord
	raw, _, err := cache.Get(svc.sessionKey(sessionID))
	if err != nil || json.Unmarshal([]byte(raw), &session) != nil {
		t.Fatal("session family record was not persisted")
	}
	if err := svc.Logout(context.Background(), "owner", sessionID, first.RefreshToken); err != nil {
		t.Fatal(err)
	}
	if cache.expires[familyRevocationKey(session.FamilyID)].Before(session.FamilyExpiresAt) {
		t.Fatalf("revocation tombstone expires before family: tombstone=%s family=%s", cache.expires[familyRevocationKey(session.FamilyID)], session.FamilyExpiresAt)
	}
}

func TestLegacyAccessCanLogoutAfterItsRefreshHasRotated(t *testing.T) {
	svc, _, _ := deviceAuthService()
	_ = svc.CacheRepo.Set(svc.sessionKey("legacy-sid"), "owner", time.Hour)
	_ = svc.CacheRepo.Set(refreshTokenKey("legacy-refresh"), `{"user_id":"owner"}`, 30*24*time.Hour)
	rotated, err := svc.Refresh(context.Background(), "legacy-refresh")
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.Logout(context.Background(), "owner", "legacy-sid", rotated.RefreshToken); err != nil {
		t.Fatalf("legacy access cannot revoke its rotated credentials: %v", err)
	}
	if _, err := svc.VerifySession(context.Background(), strings.TrimPrefix(rotated.Token, "access-")); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatal("rotated legacy access survived logout")
	}
	if _, err := svc.Refresh(context.Background(), rotated.RefreshToken); !errors.Is(err, ErrInvalidRefreshToken) {
		t.Fatal("rotated legacy refresh survived logout")
	}
	newLogin, err := svc.buildLoginOutput(context.Background(), &entity.User{ID: "owner"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.VerifySession(context.Background(), strings.TrimPrefix(newLogin.Token, "access-")); err != nil {
		t.Fatal("fresh password/device login must not join the legacy revocation group")
	}
}

// TTL-aware deterministic interleaving: pause issuance at Set, execute logout,
// then advance the cache clock before allowing the replacement to be stored.
// No sleeps or goroutine scheduling assumptions are needed to reproduce it.
type revocationTTLCache struct {
	now       time.Time
	values    map[string]string
	expires   map[string]time.Time
	beforeSet func(string)
}

func (c *revocationTTLCache) Set(key, value string, ttl time.Duration) error {
	if c.beforeSet != nil {
		c.beforeSet(key)
	}
	c.values[key] = value
	c.expires[key] = c.now.Add(ttl)
	return nil
}
func (c *revocationTTLCache) Get(key string) (string, bool, error) {
	v, ok := c.values[key]
	if !ok || !c.now.Before(c.expires[key]) {
		return "", false, nil
	}
	return v, true, nil
}
func (c *revocationTTLCache) Delete(key string) error {
	delete(c.values, key)
	delete(c.expires, key)
	return nil
}
func (c *revocationTTLCache) GetAndDelete(key string) (string, bool, error) {
	v, ok, err := c.Get(key)
	_ = c.Delete(key)
	return v, ok, err
}
func TestConcurrentRefreshCannotResurrectFamilyAfterLogoutMarkerExpires(t *testing.T) {
	for _, legacy := range []bool{false, true} {
		t.Run(map[bool]string{false: "current", true: "legacy"}[legacy], func(t *testing.T) {
			svc, _, _ := deviceAuthService()
			cache := &revocationTTLCache{now: time.Now(), values: map[string]string{}, expires: map[string]time.Time{}}
			svc.CacheRepo = cache
			first, err := svc.buildLoginOutput(context.Background(), &entity.User{ID: "owner"})
			if err != nil {
				t.Fatal(err)
			}
			sid := strings.TrimPrefix(first.Token, "access-")
			if legacy {
				_ = cache.Set(svc.sessionKey(sid), "owner", time.Hour)
				_ = cache.Set(refreshTokenKey(first.RefreshToken), `{"user_id":"owner"}`, 30*24*time.Hour)
			}
			cache.beforeSet = func(key string) {
				if !strings.HasPrefix(key, refreshTokenPrefix) {
					return
				}
				cache.beforeSet = nil
				if err := svc.Logout(context.Background(), "owner", sid, first.RefreshToken); err != nil {
					t.Fatal(err)
				}
				cache.now = cache.now.Add(10 * time.Second)
			}
			delayed, err := svc.Refresh(context.Background(), first.RefreshToken)
			if err == nil {
				t.Error("refresh issued credentials after logout completed")
			}
			cache.now = cache.now.Add(30*24*time.Hour - 5*time.Second)
			if delayed != nil {
				if _, err := svc.Refresh(context.Background(), delayed.RefreshToken); err == nil {
					t.Error("family resurrected after tombstone expiry")
				}
			}
		})
	}
}
