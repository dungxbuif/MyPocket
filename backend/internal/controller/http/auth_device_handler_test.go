package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/mypocket/backend/internal/entity"
	"github.com/mypocket/backend/internal/infrastructure/auth"
	repo "github.com/mypocket/backend/internal/infrastructure/repository"
	"github.com/mypocket/backend/internal/usecase"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"
)

type deviceHTTPTestCache struct {
	mu     sync.Mutex
	values map[string]string
}

func (r *deviceHTTPTestCache) Set(k, v string, _ time.Duration) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.values[k] = v
	return nil
}
func (r *deviceHTTPTestCache) Get(k string) (string, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	v, ok := r.values[k]
	return v, ok, nil
}
func (r *deviceHTTPTestCache) Delete(k string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.values, k)
	return nil
}
func (r *deviceHTTPTestCache) GetAndDelete(k string) (string, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	v, ok := r.values[k]
	delete(r.values, k)
	return v, ok, nil
}

func TestDeviceAuthHTTPWithPostgresAndOwnerIsolation(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("requires migrated disposable TEST_DATABASE_URL")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	password := auth.NewPasswordService()
	hashed, _ := password.Hash("fixture-password")
	users := repo.NewUserPostgresRepository(db)
	a := entity.User{ID: uuid.NewString(), GoogleSubject: uuid.NewString(), Email: uuid.NewString() + "@example.test", Name: "Device owner", PasswordHash: hashed}
	b := entity.User{ID: uuid.NewString(), GoogleSubject: uuid.NewString(), Email: uuid.NewString() + "@example.test", Name: "Other owner", PasswordHash: hashed}
	for _, u := range []*entity.User{&a, &b} {
		if err := users.Create(u); err != nil {
			t.Fatal(err)
		}
	}
	t.Cleanup(func() {
		db.Where("owner_id IN ?", []string{a.ID, b.ID}).Delete(&entity.DeviceCredential{})
		db.Where("id IN ?", []string{a.ID, b.ID}).Delete(&entity.User{})
	})
	jwt := auth.NewJWT("synthetic-auth-integration-secret-at-least-32-bytes", time.Hour)
	svc := usecase.NewAuthInteractor(users, password, jwt, &deviceHTTPTestCache{values: map[string]string{}}, nil)
	svc.Devices = repo.NewDeviceCredentialPostgresRepository(db)
	first, err := svc.Login(context.Background(), usecase.LoginInput{Email: a.Email, Password: "fixture-password"})
	if err != nil {
		t.Fatal(err)
	}
	other, err := svc.Login(context.Background(), usecase.LoginInput{Email: b.Email, Password: "fixture-password"})
	if err != nil {
		t.Fatal(err)
	}
	middleware := NewAuthMiddleware(jwt, func(id string) (string, error) { return svc.VerifySession(context.Background(), id) })
	router := NewRouter(NewAuthHandler(svc, false, nil), nil, nil, nil, nil, nil, middleware, nil)
	call := func(method, path, token string, body any) *httptest.ResponseRecorder {
		t.Helper()
		raw, _ := json.Marshal(body)
		req := httptest.NewRequest(method, "/api/v1"+path, bytes.NewReader(raw))
		req.Header.Set("Content-Type", "application/json")
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		rec := httptest.NewRecorder()
		router.Engine.ServeHTTP(rec, req)
		return rec
	}
	payload := usecase.DeviceEnrollmentInput{DeviceID: "fixture-installation", DeviceName: "Fixture iPhone", Password: "fixture-password"}
	for _, token := range []string{"", "mpk_test.invalid"} {
		if r := call("POST", "/auth/devices", token, payload); r.Code != 401 {
			t.Fatalf("unauthorized enrollment: %d", r.Code)
		}
	}
	payload.Password = "wrong"
	if r := call("POST", "/auth/devices", first.Token, payload); r.Code != 403 {
		t.Fatalf("wrong enrollment password: %d", r.Code)
	}
	payload.Password = "fixture-password"
	enrolled := call("POST", "/auth/devices", first.Token, payload)
	if enrolled.Code != 200 {
		t.Fatalf("enroll: %d %s", enrolled.Code, enrolled.Body.String())
	}
	var envelope struct {
		Data usecase.DeviceCredentialOutput `json:"data"`
	}
	if err := json.Unmarshal(enrolled.Body.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	device := envelope.Data
	var record entity.DeviceCredential
	if err := db.Where("id = ?", device.ID).First(&record).Error; err != nil {
		t.Fatal(err)
	}
	if len(record.SecretHash) != 64 || record.SecretHash == device.Secret {
		t.Fatal("raw device secret persisted")
	}
	login := map[string]string{"id": device.ID, "secret": device.Secret}
	if r := call("POST", "/auth/device-login", "", login); r.Code != 200 {
		t.Fatalf("device sign-in: %d", r.Code)
	}
	if r := call("DELETE", "/auth/devices/"+device.ID, other.Token, nil); r.Code != 404 {
		t.Fatalf("cross-owner revoke: %d", r.Code)
	}
	if r := call("POST", "/auth/logout", first.Token, map[string]string{"refresh_token": first.RefreshToken}); r.Code != 204 {
		t.Fatalf("logout: %d %s", r.Code, r.Body.String())
	}
	if r := call("POST", "/auth/devices", first.Token, payload); r.Code != 401 {
		t.Fatal("logged-out access still authorizes enrollment")
	}
	remembered := call("POST", "/auth/device-login", "", login)
	if remembered.Code != 200 {
		t.Fatal("logout destroyed remembered Face ID login")
	}
	var signedIn struct {
		Data usecase.LoginOutput `json:"data"`
	}
	json.Unmarshal(remembered.Body.Bytes(), &signedIn)
	if r := call("DELETE", "/auth/devices/"+device.ID, signedIn.Data.Token, nil); r.Code != 204 {
		t.Fatalf("revoke: %d", r.Code)
	}
	if r := call("POST", "/auth/device-login", "", login); r.Code != 401 {
		t.Fatal("revoked device still signs in")
	}
	if r := call("POST", "/auth/device-login", "", map[string]string{"id": "not-a-uuid", "secret": "mpd_fake"}); r.Code != 401 {
		t.Fatalf("malformed device identifier: %d", r.Code)
	}
}
