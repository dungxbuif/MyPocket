package httpapi

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"mypocket/internal/audit"
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

func startGoogleAuth(cfg config.Config, auditRepo AuditRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeJSON(w, http.StatusMethodNotAllowed, ErrorEnvelope("VALIDATION_FAILED", "Method not allowed", correlationID(r.Context())))
			return
		}
		appendAudit(r.Context(), auditRepo, audit.Event{CorrelationID: correlationID(r.Context()), Action: "auth.google.start", EntityType: "auth", Outcome: audit.OutcomeSuccess, Severity: audit.SeveritySecurity, Source: audit.SourceAuth, RequestMethod: r.Method, RequestPath: safeRequestPath(r), IPHash: audit.HashValue(cfg.AuditHashSecret, clientIP(r)), UserAgentHash: audit.HashValue(cfg.AuditHashSecret, r.UserAgent())})
		if cfg.OAuthFixtureMode {
			http.Redirect(w, r, "/api/v1/auth/google/callback?subject=fixture-user&email=fixture@example.com&email_verified=true&name=Fixture", http.StatusFound)
			return
		}
		if cfg.GoogleClientID == "" || cfg.GoogleClientSecret == "" || cfg.GoogleRedirectURL == "" {
			writeJSON(w, http.StatusNotImplemented, ErrorEnvelope("PROVIDER_UNAVAILABLE", "Google OAuth is not configured", correlationID(r.Context())))
			return
		}
		state, err := randomOAuthState()
		if err != nil {
			writeJSON(w, http.StatusServiceUnavailable, ErrorEnvelope("INTERNAL_RETRYABLE", "Authentication unavailable", correlationID(r.Context())))
			return
		}
		http.SetCookie(w, oauthStateCookie(state, time.Now().Add(10*time.Minute)))
		params := url.Values{}
		params.Set("client_id", cfg.GoogleClientID)
		params.Set("redirect_uri", cfg.GoogleRedirectURL)
		params.Set("response_type", "code")
		params.Set("scope", "openid email profile")
		params.Set("state", state)
		http.Redirect(w, r, "https://accounts.google.com/o/oauth2/v2/auth?"+params.Encode(), http.StatusFound)
	}
}

func googleCallback(cfg config.Config, repo IdentityRepository, auditRepo AuditRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeJSON(w, http.StatusMethodNotAllowed, ErrorEnvelope("VALIDATION_FAILED", "Method not allowed", correlationID(r.Context())))
			return
		}
		if !cfg.OAuthFixtureMode {
			if repo == nil {
				writeJSON(w, http.StatusServiceUnavailable, ErrorEnvelope("INTERNAL_RETRYABLE", "Identity repository unavailable", correlationID(r.Context())))
				auditAuthFailure(r, cfg, auditRepo, "auth.google.callback.failure", "INTERNAL_RETRYABLE", "")
				return
			}
			if r.URL.Query().Get("error") != "" {
				writeJSON(w, http.StatusUnauthorized, ErrorEnvelope("AUTH_REQUIRED", "Google login was cancelled", correlationID(r.Context())))
				auditAuthFailure(r, cfg, auditRepo, "auth.google.callback.cancelled", "AUTH_REQUIRED", "")
				return
			}
			stateCookie, err := r.Cookie("mypocket_oauth_state")
			if err != nil || stateCookie.Value == "" || stateCookie.Value != r.URL.Query().Get("state") {
				writeJSON(w, http.StatusBadRequest, ErrorEnvelope("VALIDATION_FAILED", "Invalid OAuth state", correlationID(r.Context())))
				auditAuthFailure(r, cfg, auditRepo, "auth.google.callback.invalid_state", "VALIDATION_FAILED", "")
				return
			}
			profile, err := exchangeGoogleCode(r, cfg, r.URL.Query().Get("code"))
			if err != nil {
				writeJSON(w, http.StatusBadGateway, ErrorEnvelope("PROVIDER_UNAVAILABLE", "Google login failed", correlationID(r.Context())))
				auditAuthFailure(r, cfg, auditRepo, "auth.google.callback.provider_failure", "PROVIDER_UNAVAILABLE", "")
				return
			}
			if !emailAllowed(cfg, profile.Email) {
				writeJSON(w, http.StatusForbidden, ErrorEnvelope("AUTH_FORBIDDEN", "Email is not allowed", correlationID(r.Context())))
				auditAuthFailure(r, cfg, auditRepo, "auth.google.callback.whitelist_denied", "AUTH_FORBIDDEN", profile.Email)
				return
			}
			user, err := repo.FindOrCreateGoogleUser(r.Context(), profile)
			if err != nil {
				writeJSON(w, http.StatusBadRequest, ErrorEnvelope("VALIDATION_FAILED", "Invalid Google profile", correlationID(r.Context())))
				auditAuthFailure(r, cfg, auditRepo, "auth.google.callback.invalid_profile", "VALIDATION_FAILED", profile.Email)
				return
			}
			appendAudit(r.Context(), auditRepo, audit.Event{CorrelationID: correlationID(r.Context()), ActorUserID: user.ID, ActorEmailHash: audit.HashValue(cfg.AuditHashSecret, profile.Email), Action: "auth.google.callback.success", EntityType: "auth", Outcome: audit.OutcomeSuccess, Severity: audit.SeveritySecurity, Source: audit.SourceAuth, RequestMethod: r.Method, RequestPath: safeRequestPath(r), IPHash: audit.HashValue(cfg.AuditHashSecret, clientIP(r)), UserAgentHash: audit.HashValue(cfg.AuditHashSecret, r.UserAgent())})
			_ = issueAuthRedirect(w, r, cfg, user)
			return
		}
		if repo == nil {
			writeJSON(w, http.StatusServiceUnavailable, ErrorEnvelope("INTERNAL_RETRYABLE", "Identity repository unavailable", correlationID(r.Context())))
			auditAuthFailure(r, cfg, auditRepo, "auth.fixture.callback.failure", "INTERNAL_RETRYABLE", "")
			return
		}

		verified, _ := strconv.ParseBool(r.URL.Query().Get("email_verified"))
		profile := identity.GoogleProfile{
			Subject:       r.URL.Query().Get("subject"),
			Email:         r.URL.Query().Get("email"),
			EmailVerified: verified,
			DisplayName:   r.URL.Query().Get("name"),
			AvatarURL:     r.URL.Query().Get("avatar_url"),
		}
		if !emailAllowed(cfg, profile.Email) {
			writeJSON(w, http.StatusForbidden, ErrorEnvelope("AUTH_FORBIDDEN", "Email is not allowed", correlationID(r.Context())))
			auditAuthFailure(r, cfg, auditRepo, "auth.fixture.callback.whitelist_denied", "AUTH_FORBIDDEN", profile.Email)
			return
		}
		user, err := repo.FindOrCreateGoogleUser(r.Context(), profile)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, ErrorEnvelope("VALIDATION_FAILED", "Invalid Google profile", correlationID(r.Context())))
			auditAuthFailure(r, cfg, auditRepo, "auth.fixture.callback.invalid_profile", "VALIDATION_FAILED", profile.Email)
			return
		}

		appendAudit(r.Context(), auditRepo, audit.Event{CorrelationID: correlationID(r.Context()), ActorUserID: user.ID, ActorEmailHash: audit.HashValue(cfg.AuditHashSecret, profile.Email), Action: "auth.fixture.callback.success", EntityType: "auth", Outcome: audit.OutcomeSuccess, Severity: audit.SeveritySecurity, Source: audit.SourceAuth, RequestMethod: r.Method, RequestPath: safeRequestPath(r), IPHash: audit.HashValue(cfg.AuditHashSecret, clientIP(r)), UserAgentHash: audit.HashValue(cfg.AuditHashSecret, r.UserAgent())})
		_ = issueAuthRedirect(w, r, cfg, user)
		return
	}
}

func auditAuthFailure(r *http.Request, cfg config.Config, auditRepo AuditRepository, action string, code string, email string) {
	appendAudit(r.Context(), auditRepo, audit.Event{
		CorrelationID:  correlationID(r.Context()),
		ActorEmailHash: audit.HashValue(cfg.AuditHashSecret, email),
		Action:         action,
		EntityType:     "auth",
		Outcome:        audit.OutcomeDenied,
		Severity:       audit.SeveritySecurity,
		Source:         audit.SourceAuth,
		ErrorCode:      code,
		RequestMethod:  r.Method,
		RequestPath:    safeRequestPath(r),
		IPHash:         audit.HashValue(cfg.AuditHashSecret, clientIP(r)),
		UserAgentHash:  audit.HashValue(cfg.AuditHashSecret, r.UserAgent()),
	})
}

func emailAllowed(cfg config.Config, email string) bool {
	if len(cfg.AllowedLoginEmails) == 0 {
		return true
	}
	email = strings.ToLower(strings.TrimSpace(email))
	for _, allowed := range cfg.AllowedLoginEmails {
		if email == allowed {
			return true
		}
	}
	return false
}

func issueAuthRedirect(w http.ResponseWriter, r *http.Request, cfg config.Config, user identity.User) error {
	authValue, err := identity.NewCookieSigner([]byte(cfg.CookieSecret)).Sign(user.ID, time.Now().Add(30*24*time.Hour))
	if err != nil {
		writeJSON(w, http.StatusServiceUnavailable, ErrorEnvelope("INTERNAL_RETRYABLE", "Authentication unavailable", correlationID(r.Context())))
		return err
	}
	csrfValue, err := identity.NewCSRFToken()
	if err != nil {
		writeJSON(w, http.StatusServiceUnavailable, ErrorEnvelope("INTERNAL_RETRYABLE", "Authentication unavailable", correlationID(r.Context())))
		return err
	}
	http.SetCookie(w, authCookie(authValue, time.Now().Add(30*24*time.Hour)))
	http.SetCookie(w, csrfCookie(csrfValue, time.Now().Add(30*24*time.Hour)))
	http.Redirect(w, r, cfg.PublicWebURL, http.StatusFound)
	return nil
}

type googleTokenResponse struct {
	AccessToken string `json:"access_token"`
}
type googleProfileResponse struct {
	Subject       string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

func exchangeGoogleCode(r *http.Request, cfg config.Config, code string) (identity.GoogleProfile, error) {
	if strings.TrimSpace(code) == "" {
		return identity.GoogleProfile{}, errors.New("oauth code is required")
	}
	form := url.Values{"code": {code}, "client_id": {cfg.GoogleClientID}, "client_secret": {cfg.GoogleClientSecret}, "redirect_uri": {cfg.GoogleRedirectURL}, "grant_type": {"authorization_code"}}
	resp, err := http.PostForm("https://oauth2.googleapis.com/token", form)
	if err != nil {
		return identity.GoogleProfile{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return identity.GoogleProfile{}, errors.New("google token exchange failed")
	}
	var token googleTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil || token.AccessToken == "" {
		return identity.GoogleProfile{}, errors.New("invalid google token response")
	}
	req, _ := http.NewRequestWithContext(r.Context(), http.MethodGet, "https://openidconnect.googleapis.com/v1/userinfo", nil)
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	profileResp, err := http.DefaultClient.Do(req)
	if err != nil {
		return identity.GoogleProfile{}, err
	}
	defer profileResp.Body.Close()
	if profileResp.StatusCode != http.StatusOK {
		return identity.GoogleProfile{}, errors.New("google userinfo failed")
	}
	var profile googleProfileResponse
	if err := json.NewDecoder(profileResp.Body).Decode(&profile); err != nil {
		return identity.GoogleProfile{}, err
	}
	return identity.GoogleProfile{Subject: profile.Subject, Email: profile.Email, EmailVerified: profile.EmailVerified, DisplayName: profile.Name, AvatarURL: profile.Picture}, nil
}

func randomOAuthState() (string, error) {
	var b [32]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b[:]), nil
}
func oauthStateCookie(value string, expires time.Time) *http.Cookie {
	return &http.Cookie{Name: "mypocket_oauth_state", Value: value, Path: "/", Expires: expires, HttpOnly: true, Secure: true, SameSite: http.SameSiteLaxMode}
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

		userID := authenticatedUserIDFromContext(r.Context())
		if userID == "" {
			writeJSON(w, http.StatusUnauthorized, ErrorEnvelope("AUTH_REQUIRED", "Authentication required", correlationID(r.Context())))
			return
		}
		user, err := repo.FindByID(r.Context(), userID)
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
		if !csrfSatisfiedForMethod(r) {
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
