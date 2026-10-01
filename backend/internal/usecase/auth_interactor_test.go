package usecase

import (
	"errors"
	"testing"
	"time"

	"github.com/mypocket/backend/internal/entity"
)

type authCacheStub struct {
	values map[string]string
}

func (c *authCacheStub) Set(key, value string, _ time.Duration) error {
	if c.values == nil {
		c.values = map[string]string{}
	}
	c.values[key] = value
	return nil
}
func (c *authCacheStub) Get(key string) (string, bool, error) {
	value, ok := c.values[key]
	return value, ok, nil
}
func (c *authCacheStub) Delete(key string) error {
	delete(c.values, key)
	return nil
}
func (c *authCacheStub) GetAndDelete(key string) (string, bool, error) {
	value, ok := c.values[key]
	delete(c.values, key)
	return value, ok, nil
}

type authUserStub struct{ user *entity.User }

func (s authUserStub) FindByEmail(string) (*entity.User, error) { return s.user, nil }
func (s authUserStub) FindByID(id string) (*entity.User, error) {
	if s.user != nil && s.user.ID == id {
		return s.user, nil
	}
	return nil, errors.New("not found")
}
func (s authUserStub) FindOrCreateGoogleUser(entity.GoogleProfile) (*entity.User, error) {
	return s.user, nil
}
func (s authUserStub) SetTimezone(string, string, bool) (*entity.User, error) { return s.user, nil }
func (s authUserStub) Create(*entity.User) error                              { return nil }

type authTokenStub struct{ calls int }

type authPasswordStub struct{}

func (authPasswordStub) Hash(string) (string, error) { return "hash", nil }
func (authPasswordStub) Check(string, string) bool   { return true }

func (s *authTokenStub) Generate(userID, email, name, sessionID string) (string, time.Time, error) {
	s.calls++
	return "access-" + sessionID, time.Now().Add(time.Hour), nil
}
func (*authTokenStub) Verify(string) (*TokenClaims, error) { return nil, ErrInvalidToken }

func TestAuthRefreshRotatesTokenAndRejectsReplay(t *testing.T) {
	cache := &authCacheStub{values: map[string]string{}}
	service := NewAuthInteractor(
		authUserStub{user: &entity.User{ID: "user-1", Email: "user@example.com", Name: "User"}},
		authPasswordStub{},
		&authTokenStub{},
		cache,
		nil,
	)

	first, err := service.Login(nil, LoginInput{Email: "user@example.com"})
	if err != nil {
		t.Fatal(err)
	}
	if first.RefreshToken == "" || first.RefreshExpiresAt.IsZero() {
		t.Fatalf("login did not return refresh credentials: %#v", first)
	}

	rotated, err := service.Refresh(nil, first.RefreshToken)
	if err != nil {
		t.Fatal(err)
	}
	if rotated.RefreshToken == first.RefreshToken || rotated.Token == first.Token {
		t.Fatalf("refresh credentials were not rotated: first=%#v rotated=%#v", first, rotated)
	}
	if _, err := service.Refresh(nil, first.RefreshToken); !errors.Is(err, ErrInvalidRefreshToken) {
		t.Fatalf("replayed refresh token should be rejected, got %v", err)
	}
}
