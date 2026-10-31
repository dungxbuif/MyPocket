package push

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
)

const (
	sandboxAPNsEndpoint    = "https://api.sandbox.push.apple.com"
	productionAPNsEndpoint = "https://api.push.apple.com"
	maxAPNsRetryAfter      = 24 * time.Hour
)

// APNsTransportOptions contains test seams. BaseURL is empty in production so
// the configured sandbox/production endpoint is selected by Config.Environment.
type APNsTransportOptions struct {
	Client  *http.Client
	BaseURL string
	Now     func() time.Time
}

type APNsTransport struct {
	config  Config
	client  *http.Client
	baseURL string
	now     func() time.Time
	key     *ecdsa.PrivateKey
}

// NewAPNsTransport creates the token-authenticated HTTP transport. It parses
// the private key once and keeps it only in memory; no secret is persisted or
// included in returned errors.
func NewAPNsTransport(config Config, options APNsTransportOptions) (*APNsTransport, error) {
	if err := config.Validate(); err != nil || !config.Enabled {
		return nil, ErrInvalidConfig
	}
	key, err := parseAPNsPrivateKey(config.PrivateKey)
	if err != nil {
		return nil, ErrInvalidConfig
	}
	baseURL := strings.TrimRight(strings.TrimSpace(options.BaseURL), "/")
	if baseURL == "" {
		baseURL = APNsEndpoint(config.Environment)
	}
	parsed, err := url.Parse(baseURL)
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" {
		return nil, ErrInvalidConfig
	}
	client := options.Client
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second, Transport: &http.Transport{ForceAttemptHTTP2: true}}
	}
	now := options.Now
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	return &APNsTransport{config: config, client: client, baseURL: baseURL, now: now, key: key}, nil
}

func APNsEndpoint(environment string) string {
	switch strings.ToLower(strings.TrimSpace(environment)) {
	case "sandbox":
		return sandboxAPNsEndpoint
	case "production":
		return productionAPNsEndpoint
	default:
		return ""
	}
}

func (t *APNsTransport) Send(ctx context.Context, request Request) (Response, error) {
	if t == nil || t.client == nil || t.key == nil || strings.TrimSpace(request.DeviceToken) == "" || len(request.Payload) == 0 || !json.Valid(request.Payload) {
		return Response{}, ErrTransport
	}
	token := strings.TrimSpace(request.DeviceToken)
	if strings.ContainsAny(token, "\r\n") {
		return Response{}, ErrTransport
	}
	now := t.now().UTC()
	bearer, err := t.signJWT(now)
	if err != nil {
		return Response{}, ErrTransport
	}
	endpoint := t.baseURL + "/3/device/" + url.PathEscape(token)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(string(request.Payload)))
	if err != nil {
		return Response{}, ErrTransport
	}
	req.Header.Set("authorization", "bearer "+bearer)
	req.Header.Set("apns-topic", t.config.Topic)
	req.Header.Set("apns-push-type", "background")
	req.Header.Set("apns-priority", "5")
	req.Header.Set("content-type", "application/json")
	if collapseID := strings.TrimSpace(request.CollapseID); collapseID != "" {
		req.Header.Set("apns-collapse-id", collapseID)
	}
	response, err := t.client.Do(req)
	if err != nil {
		return Response{}, ErrTransport
	}
	defer response.Body.Close()
	_, _ = io.Copy(io.Discard, response.Body)
	result := Response{StatusCode: response.StatusCode, RetryAfter: parseRetryAfter(response.Header.Get("Retry-After"), now)}
	if response.StatusCode >= http.StatusBadRequest {
		result.FailureCode = "apns_http_" + strconv.Itoa(response.StatusCode)
	}
	return result, nil
}

func (t *APNsTransport) signJWT(now time.Time) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
		"iss": t.config.TeamID,
		"iat": now.Unix(),
	})
	token.Header["kid"] = t.config.KeyID
	signed, err := token.SignedString(t.key)
	if err != nil {
		return "", errors.New("APNs JWT signing failed")
	}
	return signed, nil
}

func parseAPNsPrivateKey(value string) (*ecdsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(strings.ReplaceAll(strings.TrimSpace(value), `\n`, "\n")))
	if block == nil {
		return nil, errors.New("APNs private key is not PEM")
	}
	parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		parsedEC, ecErr := x509.ParseECPrivateKey(block.Bytes)
		if ecErr != nil {
			return nil, errors.New("APNs private key is not EC")
		}
		parsed = parsedEC
	}
	key, ok := parsed.(*ecdsa.PrivateKey)
	if !ok || key.Curve != elliptic.P256() {
		return nil, errors.New("APNs private key is not P-256")
	}
	return key, nil
}

func parseRetryAfter(value string, now time.Time) time.Duration {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0
	}
	if seconds, err := strconv.ParseInt(value, 10, 64); err == nil {
		if seconds <= 0 {
			return 0
		}
		delay := time.Duration(seconds) * time.Second
		if delay < 0 || delay > maxAPNsRetryAfter {
			return maxAPNsRetryAfter
		}
		return delay
	}
	when, err := http.ParseTime(value)
	if err != nil {
		return 0
	}
	delay := when.Sub(now)
	if delay <= 0 {
		return 0
	}
	if delay > maxAPNsRetryAfter {
		return maxAPNsRetryAfter
	}
	return delay
}
