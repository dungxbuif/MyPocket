package identity

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

type Repository struct {
	conn *sql.DB
}

func NewRepository(conn *sql.DB) *Repository {
	return &Repository{conn: conn}
}

func (r *Repository) FindOrCreateGoogleUser(ctx context.Context, profile GoogleProfile) (User, error) {
	if strings.TrimSpace(profile.Subject) == "" {
		return User{}, fmt.Errorf("google subject is required")
	}
	if strings.TrimSpace(profile.Email) == "" {
		return User{}, fmt.Errorf("email is required")
	}

	row := r.conn.QueryRowContext(ctx, `
		INSERT INTO users (google_subject, email, email_verified, display_name, avatar_url)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (google_subject)
		DO UPDATE SET
			email = EXCLUDED.email,
			email_verified = EXCLUDED.email_verified,
			display_name = EXCLUDED.display_name,
			avatar_url = EXCLUDED.avatar_url,
			updated_at = now()
		WHERE users.disabled_at IS NULL
		RETURNING id::text, google_subject, email, email_verified, display_name, avatar_url, created_at, updated_at
	`, profile.Subject, profile.Email, profile.EmailVerified, profile.DisplayName, profile.AvatarURL)

	var user User
	if err := row.Scan(&user.ID, &user.GoogleSubject, &user.Email, &user.EmailVerified, &user.DisplayName, &user.AvatarURL, &user.CreatedAt, &user.UpdatedAt); err != nil {
		return User{}, fmt.Errorf("find or create google user: %w", err)
	}
	return user, nil
}

func (r *Repository) FindByID(ctx context.Context, id string) (User, error) {
	row := r.conn.QueryRowContext(ctx, `
		SELECT id::text, google_subject, email, email_verified, display_name, avatar_url, created_at, updated_at
		FROM users
		WHERE id = $1 AND disabled_at IS NULL
	`, id)

	var user User
	if err := row.Scan(&user.ID, &user.GoogleSubject, &user.Email, &user.EmailVerified, &user.DisplayName, &user.AvatarURL, &user.CreatedAt, &user.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return User{}, ErrUserNotFound
		}
		return User{}, fmt.Errorf("find user: %w", err)
	}
	return user, nil
}
