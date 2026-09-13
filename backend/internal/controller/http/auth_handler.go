package httpapi

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/mypocket/backend/internal/entity"
	"github.com/mypocket/backend/internal/usecase"

	"github.com/gin-gonic/gin"
)

const (
	httpStatusMessageBadRequest               = "request không hợp lệ"
	httpStatusMessageLoginFailure             = "đăng nhập thất bại"
	httpStatusMessageGoogleLoginUnavailable   = "google login chưa khả dụng"
	httpStatusMessageGoogleLoginCancelled     = "Google login was cancelled"
	httpStatusMessageGoogleLoginStateInvalid  = "trạng thái Google OAuth không hợp lệ"
	httpStatusMessageGoogleProfileInvalid     = "thông tin Google không hợp lệ"
	httpStatusMessageOAuthFixtureMissingEmail = "thiếu email trong OAuth fixture"
	httpStatusMessageOAuthFixtureMissingSubj  = "thiếu subject trong OAuth fixture"
	httpStatusMessageEmailNotAllowed          = "email không được phép đăng nhập"
	authCookieStateName                       = "mypocket_oauth_state"
	authCookieFrontendCallbackName            = "mypocket_oauth_frontend_callback"
	oauthCallbackSeconds                      = 600
	googleAuthCallbackPath                    = "/api/v1/auth/google/callback"
	googleLoginCallbackUserFallback           = "Google User"
	googleAuthFixtureSubject                  = "fixture-user"
	googleAuthFixtureEmail                    = "fixture@example.com"
	googleAuthFixtureName                     = "Fixture"
)

type AuthHandler struct {
	Auth               *usecase.AuthInteractor
	FixtureMode        bool
	AllowedLoginEmails []string
}

func NewAuthHandler(auth *usecase.AuthInteractor, fixtureMode bool, allowedEmails []string) *AuthHandler {
	return &AuthHandler{Auth: auth, FixtureMode: fixtureMode, AllowedLoginEmails: allowedEmails}
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Login godoc
// @Summary Login with email and password
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body loginRequest true "Login credentials"
// @Success 200 {object} usecase.LoginOutput
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/v1/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": httpStatusMessageBadRequest})
		return
	}
	out, err := h.Auth.Login(c.Request.Context(), usecase.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, out)
}

// StartGoogleAuth godoc
// @Summary Start Google OAuth login
// @Tags Authentication
// @Produce json
// @Param frontend_callback query string false "Absolute frontend callback URL"
// @Success 302 "Redirect to Google authorization"
// @Failure 400 {object} map[string]string
// @Failure 501 {object} map[string]string
// @Router /api/v1/auth/google [get]
func (h *AuthHandler) StartGoogleAuth(c *gin.Context) {
	if h.FixtureMode {
		target, err := buildFixtureRedirectURL(c.Query("frontend_callback"))
		if err != nil {
			fixtureURL := googleAuthCallbackPath + "?subject=fixture-user&email=fixture@example.com&email_verified=true&name=Fixture"
			c.Redirect(http.StatusFound, fixtureURL)
			return
		}
		query := target.Query()
		query.Set("subject", googleAuthFixtureSubject)
		query.Set("email", googleAuthFixtureEmail)
		query.Set("email_verified", "true")
		query.Set("name", googleAuthFixtureName)
		target.RawQuery = query.Encode()
		c.Redirect(http.StatusFound, target.String())
		return
	}
	if !h.Auth.IsGoogleAuthReady() {
		c.JSON(http.StatusNotImplemented, gin.H{"error": httpStatusMessageGoogleLoginUnavailable})
		return
	}
	if callbackURL := strings.TrimSpace(c.Query("frontend_callback")); callbackURL != "" {
		target, err := buildFixtureRedirectURL(callbackURL)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.SetCookie(authCookieFrontendCallbackName, target.String(), oauthCallbackSeconds, "/", "", false, true)
	}
	state, err := newOAuthState()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": httpStatusMessageLoginFailure})
		return
	}
	c.SetCookie(authCookieStateName, state, oauthCallbackSeconds, "/", "", false, true)
	authURL, ok := h.Auth.GoogleAuthURL(state)
	if !ok {
		c.JSON(http.StatusNotImplemented, gin.H{"error": httpStatusMessageGoogleLoginUnavailable})
		return
	}
	c.Redirect(http.StatusFound, authURL)
}

func buildFixtureRedirectURL(raw string) (*url.URL, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, errors.New("missing frontend callback")
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return nil, err
	}
	if parsed.Scheme == "" && parsed.Host == "" {
		return nil, errors.New("frontend callback must be absolute URL")
	}
	return parsed, nil
}

// GoogleCallback godoc
// @Summary Complete Google OAuth login
// @Tags Authentication
// @Produce json
// @Param code query string false "Google authorization code"
// @Param state query string false "OAuth state"
// @Param error query string false "OAuth error"
// @Success 200 {object} usecase.LoginOutput
// @Success 302 "Redirect to frontend callback"
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Router /api/v1/auth/google/callback [get]
func (h *AuthHandler) GoogleCallback(c *gin.Context) {
	if h.FixtureMode {
		profile, err := h.loadFixtureProfile(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if !h.isEmailAllowed(profile.Email) {
			c.JSON(http.StatusForbidden, gin.H{"error": httpStatusMessageEmailNotAllowed})
			return
		}
		out, err := h.Auth.LoginWithGoogle(c.Request.Context(), profile)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, out)
		return
	}

	if c.Query("error") != "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": httpStatusMessageGoogleLoginCancelled})
		return
	}
	stateCookie, err := c.Cookie(authCookieStateName)
	if err != nil || strings.TrimSpace(stateCookie) == "" || stateCookie != c.Query("state") {
		c.JSON(http.StatusBadRequest, gin.H{"error": httpStatusMessageGoogleLoginStateInvalid})
		return
	}
	profile, err := h.Auth.FetchGoogleProfile(strings.TrimSpace(c.Query("code")))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": httpStatusMessageGoogleProfileInvalid})
		return
	}
	if !h.isEmailAllowed(profile.Email) {
		c.JSON(http.StatusForbidden, gin.H{"error": httpStatusMessageEmailNotAllowed})
		return
	}
	out, err := h.Auth.LoginWithGoogle(c.Request.Context(), profile)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": httpStatusMessageGoogleProfileInvalid})
		return
	}
	if callbackURL, callbackErr := c.Cookie(authCookieFrontendCallbackName); callbackErr == nil {
		target, err := url.Parse(callbackURL)
		if err == nil {
			query := target.Query()
			query.Set("token", out.Token)
			query.Set("user_id", out.User.ID)
			query.Set("name", out.User.Name)
			query.Set("email", out.User.Email)
			query.Set("created_at", out.User.CreatedAt.Format(time.RFC3339Nano))
			query.Set("expires_at", out.ExpiresAt.Format(time.RFC3339Nano))
			target.RawQuery = query.Encode()
			c.SetCookie(authCookieFrontendCallbackName, "", -1, "/", "", false, true)
			c.Redirect(http.StatusFound, target.String())
			return
		}
	}
	c.JSON(http.StatusOK, out)
}

func (h *AuthHandler) loadFixtureProfile(c *gin.Context) (entity.GoogleProfile, error) {
	subject := strings.TrimSpace(c.Query("subject"))
	email := strings.TrimSpace(c.Query("email"))
	if subject == "" {
		return entity.GoogleProfile{}, errors.New(httpStatusMessageOAuthFixtureMissingSubj)
	}
	if email == "" {
		return entity.GoogleProfile{}, errors.New(httpStatusMessageOAuthFixtureMissingEmail)
	}
	verified, _ := strconv.ParseBool(c.Query("email_verified"))
	displayName := strings.TrimSpace(c.Query("name"))
	if displayName == "" {
		displayName = googleLoginCallbackUserFallback
	}
	return entity.GoogleProfile{
		Subject:       subject,
		Email:         email,
		EmailVerified: verified,
		DisplayName:   displayName,
		AvatarURL:     strings.TrimSpace(c.Query("avatar_url")),
	}, nil
}

func (h *AuthHandler) isEmailAllowed(email string) bool {
	if len(h.AllowedLoginEmails) == 0 {
		return true
	}
	normalized := strings.ToLower(strings.TrimSpace(email))
	for _, allowed := range h.AllowedLoginEmails {
		if normalized == strings.ToLower(strings.TrimSpace(allowed)) {
			return true
		}
	}
	return false
}

func newOAuthState() (string, error) {
	const stateBytes = 32
	buf := make([]byte, stateBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}
