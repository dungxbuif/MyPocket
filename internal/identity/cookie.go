package identity

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

type CookieSigner struct {
	secret []byte
}

const AuthCookieName = "mypocket_auth"

type Claims struct {
	UserID    string `json:"uid"`
	IssuedAt  int64  `json:"iat"`
	ExpiresAt int64  `json:"exp"`
}

func NewCookieSigner(secret []byte) CookieSigner {
	copied := make([]byte, len(secret))
	copy(copied, secret)
	return CookieSigner{secret: copied}
}

func (s CookieSigner) Sign(userID string, expiresAt time.Time) (string, error) {
	if strings.TrimSpace(userID) == "" {
		return "", errors.New("user id is required")
	}
	if len(s.secret) < 32 {
		return "", errors.New("cookie secret must be at least 32 bytes")
	}
	claims := Claims{UserID: userID, IssuedAt: time.Now().Unix(), ExpiresAt: expiresAt.Unix()}
	body, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("marshal claims: %w", err)
	}
	payload := base64.RawURLEncoding.EncodeToString(body)
	signature := s.signature(payload)
	return "v1." + payload + "." + signature, nil
}

func (s CookieSigner) Verify(value string) (Claims, error) {
	parts := strings.Split(value, ".")
	if len(parts) != 3 || parts[0] != "v1" {
		return Claims{}, errors.New("invalid auth cookie")
	}
	if !hmac.Equal([]byte(s.signature(parts[1])), []byte(parts[2])) {
		return Claims{}, errors.New("invalid auth cookie")
	}
	body, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return Claims{}, errors.New("invalid auth cookie")
	}
	var claims Claims
	if err := json.Unmarshal(body, &claims); err != nil {
		return Claims{}, errors.New("invalid auth cookie")
	}
	if claims.UserID == "" || time.Now().Unix() >= claims.ExpiresAt {
		return Claims{}, errors.New("invalid auth cookie")
	}
	return claims, nil
}

func (s CookieSigner) signature(payload string) string {
	mac := hmac.New(sha256.New, s.secret)
	_, _ = mac.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
