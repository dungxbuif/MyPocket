package identity

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"
)

const APIKeyPrefix = "mpk_"

type APIKey struct {
	ID         string     `json:"id"`
	UserID     string     `json:"user_id"`
	Name       string     `json:"name"`
	KeyPrefix  string     `json:"key_prefix"`
	KeyHash    string     `json:"-"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	RevokedAt  *time.Time `json:"revoked_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

type CreatedAPIKey struct {
	APIKey
	Plaintext string `json:"plaintext"`
}

func (r *Repository) CreateAPIKey(ctx context.Context, userID string, name string, secret string) (CreatedAPIKey, error) {
	userID = strings.TrimSpace(userID)
	name = strings.TrimSpace(name)
	if userID == "" || name == "" || len(name) > 80 {
		return CreatedAPIKey{}, fmt.Errorf("%w: invalid api key input", ErrValidation)
	}
	plaintext, err := NewAPIKeyToken()
	if err != nil {
		return CreatedAPIKey{}, err
	}
	keyHash, err := HashAPIKey(secret, plaintext)
	if err != nil {
		return CreatedAPIKey{}, err
	}
	keyPrefix := plaintext[:min(len(plaintext), 12)]
	var key APIKey
	err = r.conn.QueryRowContext(ctx, `
		INSERT INTO api_keys (user_id, name, key_prefix, key_hash)
		VALUES ($1, $2, $3, $4)
		RETURNING id::text, user_id::text, name, key_prefix, key_hash, last_used_at, revoked_at, created_at
	`, userID, name, keyPrefix, keyHash).Scan(&key.ID, &key.UserID, &key.Name, &key.KeyPrefix, &key.KeyHash, &key.LastUsedAt, &key.RevokedAt, &key.CreatedAt)
	if err != nil {
		return CreatedAPIKey{}, fmt.Errorf("create api key: %w", err)
	}
	return CreatedAPIKey{APIKey: key, Plaintext: plaintext}, nil
}

func (r *Repository) ListAPIKeys(ctx context.Context, userID string) ([]APIKey, error) {
	rows, err := r.conn.QueryContext(ctx, `
		SELECT id::text, user_id::text, name, key_prefix, last_used_at, revoked_at, created_at
		FROM api_keys
		WHERE user_id = $1
		ORDER BY created_at DESC, id DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("list api keys: %w", err)
	}
	defer rows.Close()
	keys := []APIKey{}
	for rows.Next() {
		var key APIKey
		if err := rows.Scan(&key.ID, &key.UserID, &key.Name, &key.KeyPrefix, &key.LastUsedAt, &key.RevokedAt, &key.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan api key: %w", err)
		}
		keys = append(keys, key)
	}
	return keys, rows.Err()
}

func (r *Repository) RevokeAPIKey(ctx context.Context, userID string, keyID string) (APIKey, error) {
	var key APIKey
	err := r.conn.QueryRowContext(ctx, `
		UPDATE api_keys
		SET revoked_at = now(), updated_at = now()
		WHERE id = $1 AND user_id = $2 AND revoked_at IS NULL
		RETURNING id::text, user_id::text, name, key_prefix, key_hash, last_used_at, revoked_at, created_at
	`, keyID, userID).Scan(&key.ID, &key.UserID, &key.Name, &key.KeyPrefix, &key.KeyHash, &key.LastUsedAt, &key.RevokedAt, &key.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return APIKey{}, ErrUserNotFound
		}
		return APIKey{}, fmt.Errorf("revoke api key: %w", err)
	}
	return key, nil
}

func (r *Repository) AuthenticateAPIKey(ctx context.Context, plaintext string, secret string) (User, APIKey, error) {
	keyHash, err := HashAPIKey(secret, plaintext)
	if err != nil {
		return User{}, APIKey{}, err
	}
	row := r.conn.QueryRowContext(ctx, `
		UPDATE api_keys k
		SET last_used_at = now(), updated_at = now()
		FROM users u
		WHERE k.key_hash = $1 AND k.revoked_at IS NULL AND u.id = k.user_id AND u.disabled_at IS NULL
		RETURNING k.id::text, k.user_id::text, k.name, k.key_prefix, k.key_hash, k.last_used_at, k.revoked_at, k.created_at
	`, keyHash)
	var key APIKey
	if err := row.Scan(&key.ID, &key.UserID, &key.Name, &key.KeyPrefix, &key.KeyHash, &key.LastUsedAt, &key.RevokedAt, &key.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, APIKey{}, ErrUserNotFound
		}
		return User{}, APIKey{}, fmt.Errorf("authenticate api key: %w", err)
	}
	user, err := r.FindByID(ctx, key.UserID)
	if err != nil {
		return User{}, APIKey{}, err
	}
	return user, key, nil
}

func NewAPIKeyToken() (string, error) {
	var b [32]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", fmt.Errorf("generate api key: %w", err)
	}
	return APIKeyPrefix + base64.RawURLEncoding.EncodeToString(b[:]), nil
}

func HashAPIKey(secret string, plaintext string) (string, error) {
	plaintext = strings.TrimSpace(plaintext)
	if !strings.HasPrefix(plaintext, APIKeyPrefix) {
		return "", fmt.Errorf("%w: invalid api key", ErrValidation)
	}
	if len(secret) < 32 {
		return "", fmt.Errorf("%w: api key hash secret is too short", ErrValidation)
	}
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(plaintext))
	return hex.EncodeToString(mac.Sum(nil)), nil
}
