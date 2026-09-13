package auth

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/mypocket/backend/internal/entity"
)

type GoogleOAuth struct {
	client       *http.Client
	clientID     string
	clientSecret string
	redirectURL  string
}

const (
	googleOAuthTokenURL     = "https://oauth2.googleapis.com/token"
	googleUserInfoURL       = "https://openidconnect.googleapis.com/v1/userinfo"
	googleOAuthGrantType    = "authorization_code"
	googleOAuthScopeProfile = "openid email profile"
	googleOAuthRespTypeCode = "code"
)

type GoogleOAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
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

func NewGoogleOAuth(cfg GoogleOAuthConfig) *GoogleOAuth {
	return &GoogleOAuth{
		client:       &http.Client{Timeout: 10 * time.Second},
		clientID:     cfg.ClientID,
		clientSecret: cfg.ClientSecret,
		redirectURL:  cfg.RedirectURL,
	}
}

func (g *GoogleOAuth) IsReady() bool {
	return g.clientID != "" && g.clientSecret != "" && g.redirectURL != ""
}

func (g *GoogleOAuth) BuildAuthURL(state string) string {
	params := url.Values{}
	params.Set("client_id", g.clientID)
	params.Set("redirect_uri", g.redirectURL)
	params.Set("response_type", googleOAuthRespTypeCode)
	params.Set("scope", googleOAuthScopeProfile)
	params.Set("state", state)
	return "https://accounts.google.com/o/oauth2/v2/auth?" + params.Encode()
}

func (g *GoogleOAuth) ExchangeCode(code string) (entity.GoogleProfile, error) {
	if code == "" {
		return entity.GoogleProfile{}, errors.New("oauth code is required")
	}

	form := url.Values{
		"code":          {code},
		"client_id":     {g.clientID},
		"client_secret": {g.clientSecret},
		"redirect_uri":  {g.redirectURL},
		"grant_type":    {googleOAuthGrantType},
	}

	resp, err := g.client.PostForm(googleOAuthTokenURL, form)
	if err != nil {
		return entity.GoogleProfile{}, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		buf := new(bytes.Buffer)
		_, _ = buf.ReadFrom(resp.Body)
		return entity.GoogleProfile{}, fmt.Errorf("google token exchange failed: status %d body %s", resp.StatusCode, buf.String())
	}

	var token googleTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return entity.GoogleProfile{}, err
	}
	if token.AccessToken == "" {
		return entity.GoogleProfile{}, errors.New("google token response missing access_token")
	}

	request, err := http.NewRequest(http.MethodGet, googleUserInfoURL, nil)
	if err != nil {
		return entity.GoogleProfile{}, err
	}
	request.Header.Set("Authorization", "Bearer "+token.AccessToken)
	userinfoResp, err := g.client.Do(request)
	if err != nil {
		return entity.GoogleProfile{}, err
	}
	defer func() { _ = userinfoResp.Body.Close() }()
	if userinfoResp.StatusCode != http.StatusOK {
		buf := new(bytes.Buffer)
		_, _ = buf.ReadFrom(userinfoResp.Body)
		return entity.GoogleProfile{}, fmt.Errorf("google userinfo failed: status %d body %s", userinfoResp.StatusCode, buf.String())
	}

	var profile googleProfileResponse
	if err := json.NewDecoder(userinfoResp.Body).Decode(&profile); err != nil {
		return entity.GoogleProfile{}, err
	}

	return entity.GoogleProfile{
		Subject:       profile.Subject,
		Email:         profile.Email,
		EmailVerified: profile.EmailVerified,
		DisplayName:   profile.Name,
		AvatarURL:     profile.Picture,
	}, nil
}
