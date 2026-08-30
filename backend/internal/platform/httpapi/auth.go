package httpapi

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"mypocket/internal/identity"
	"mypocket/internal/platform/config"
)

type currentUserBody struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	DisplayName   string `json:"display_name"`
	AvatarURL     string `json:"avatar_url"`
}

type currentUserResponse struct {
	Status        string          `json:"status"`
	User          currentUserBody `json:"user"`
	CorrelationID string          `json:"correlation_id"`
}

func startGoogleAuth(cfg config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeJSON(w, http.StatusMethodNotAllowed, ErrorEnvelope("VALIDATION_FAILED", "Method not allowed", correlationID(r.Context())))
			return
		}
		if cfg.OAuthFixtureMode {
			http.Redirect(w, r, "/api/v1/auth/google/callback?subject=fixture-user&email=fixture@example.com&email_verified=true&name=Fixture", http.StatusFound)
			return
		}
		writeJSON(w, http.StatusNotImplemented, ErrorEnvelope("PROVIDER_UNAVAILABLE", "Google OAuth is not configured", correlationID(r.Context())))
	}
}

func googleCallback(cfg config.Config, repo IdentityRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeJSON(w, http.StatusMethodNotAllowed, ErrorEnvelope("VALIDATION_FAILED", "Method not allowed", correlationID(r.Context())))
			return
		}
		if !cfg.OAuthFixtureMode {
			writeJSON(w, http.StatusBadGateway, ErrorEnvelope("PROVIDER_UNAVAILABLE", "Google OAuth is not configured", correlationID(r.Context())))
			return
		}
		if repo == nil {
			writeJSON(w, http.StatusServiceUnavailable, ErrorEnvelope("INTERNAL_RETRYABLE", "Identity repository unavailable", correlationID(r.Context())))
			return
		}

		verified, _ := strconv.ParseBool(r.URL.Query().Get("email_verified"))
		user, err := repo.FindOrCreateGoogleUser(r.Context(), identity.GoogleProfile{
			Subject:       r.URL.Query().Get("subject"),
			Email:         r.URL.Query().Get("email"),
			EmailVerified: verified,
			DisplayName:   r.URL.Query().Get("name"),
			AvatarURL:     r.URL.Query().Get("avatar_url"),
		})
		if err != nil {
			writeJSON(w, http.StatusBadRequest, ErrorEnvelope("VALIDATION_FAILED", "Invalid Google profile", correlationID(r.Context())))
			return
		}

		authValue, err := identity.NewCookieSigner([]byte(cfg.CookieSecret)).Sign(user.ID, time.Now().Add(30*24*time.Hour))
		if err != nil {
			writeJSON(w, http.StatusServiceUnavailable, ErrorEnvelope("INTERNAL_RETRYABLE", "Authentication unavailable", correlationID(r.Context())))
			return
		}
		csrfValue, err := identity.NewCSRFToken()
		if err != nil {
			writeJSON(w, http.StatusServiceUnavailable, ErrorEnvelope("INTERNAL_RETRYABLE", "Authentication unavailable", correlationID(r.Context())))
			return
		}

		http.SetCookie(w, authCookie(authValue, time.Now().Add(30*24*time.Hour)))
		http.SetCookie(w, csrfCookie(csrfValue, time.Now().Add(30*24*time.Hour)))
		http.Redirect(w, r, cfg.PublicWebURL, http.StatusFound)
	}
}

func currentUser(cfg config.Config, repo IdentityRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeJSON(w, http.StatusMethodNotAllowed, ErrorEnvelope("VALIDATION_FAILED", "Method not allowed", correlationID(r.Context())))
			return
		}
		if repo == nil {
			writeJSON(w, http.StatusServiceUnavailable, ErrorEnvelope("INTERNAL_RETRYABLE", "Identity repository unavailable", correlationID(r.Context())))
			return
		}

		cookie, err := r.Cookie(identity.AuthCookieName)
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, ErrorEnvelope("AUTH_REQUIRED", "Authentication required", correlationID(r.Context())))
			return
		}
		claims, err := identity.NewCookieSigner([]byte(cfg.CookieSecret)).Verify(cookie.Value)
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, ErrorEnvelope("AUTH_REQUIRED", "Authentication required", correlationID(r.Context())))
			return
		}
		user, err := repo.FindByID(r.Context(), claims.UserID)
		if err != nil {
			status := http.StatusServiceUnavailable
			code := "INTERNAL_RETRYABLE"
			message := "User unavailable"
			if errors.Is(err, identity.ErrUserNotFound) {
				status = http.StatusUnauthorized
				code = "AUTH_REQUIRED"
				message = "Authentication required"
			}
			writeJSON(w, status, ErrorEnvelope(code, message, correlationID(r.Context())))
			return
		}

		writeJSON(w, http.StatusOK, currentUserResponse{
			Status: "ok",
			User: currentUserBody{
				ID:            user.ID,
				Email:         user.Email,
				EmailVerified: user.EmailVerified,
				DisplayName:   user.DisplayName,
				AvatarURL:     user.AvatarURL,
			},
			CorrelationID: correlationID(r.Context()),
		})
	}
}

func logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, ErrorEnvelope("VALIDATION_FAILED", "Method not allowed", correlationID(r.Context())))
		return
	}
	expired := time.Unix(0, 0)
	http.SetCookie(w, authCookie("", expired))
	http.SetCookie(w, csrfCookie("", expired))
	writeJSON(w, http.StatusOK, Envelope{
		Status:        "ok",
		CorrelationID: correlationID(r.Context()),
	})
}

func requireCSRF(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(identity.CSRFCookieName)
		header := r.Header.Get(identity.CSRFHeaderName)
		if err != nil || header == "" || cookie.Value == "" || header != cookie.Value {
			writeJSON(w, http.StatusForbidden, ErrorEnvelope("CSRF_REQUIRED", "CSRF token is required", correlationID(r.Context())))
			return
		}
		next.ServeHTTP(w, r)
	})
}

func authCookie(value string, expires time.Time) *http.Cookie {
	return &http.Cookie{
		Name:     identity.AuthCookieName,
		Value:    value,
		Path:     "/",
		Expires:  expires,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}
}

func csrfCookie(value string, expires time.Time) *http.Cookie {
	return &http.Cookie{
		Name:     identity.CSRFCookieName,
		Value:    value,
		Path:     "/",
		Expires:  expires,
		HttpOnly: false,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}
}
