package push

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
)

func TestAPNsTransportSignsRequestAndSendsOnlyAppPayload(t *testing.T) {
	key, privateKeyPEM := testAPNsKey(t)
	now := time.Date(2026, 11, 1, 3, 4, 5, 0, time.UTC)
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/3/device/device-token" {
			t.Fatalf("unexpected APNs request: %s %s", r.Method, r.URL.Path)
		}
		if got := r.Header.Get("apns-topic"); got != "com.example.mypocket" {
			t.Fatalf("unexpected topic: %q", got)
		}
		if got := r.Header.Get("apns-collapse-id"); got != "notification-1" {
			t.Fatalf("unexpected collapse id: %q", got)
		}
		if got := r.Header.Get("content-type"); got != "application/json" {
			t.Fatalf("unexpected content type: %q", got)
		}
		if got := r.Header.Get("apns-push-type"); got != "background" {
			t.Fatalf("unexpected push type: %q", got)
		}
		if got := r.Header.Get("apns-priority"); got != "5" {
			t.Fatalf("unexpected priority: %q", got)
		}
		encoded := strings.TrimSpace(strings.TrimPrefix(r.Header.Get("authorization"), "bearer "))
		if encoded == "" {
			t.Fatal("missing bearer token")
		}
		claims := jwt.MapClaims{}
		parsed, err := jwt.ParseWithClaims(encoded, claims, func(token *jwt.Token) (any, error) {
			if token.Method != jwt.SigningMethodES256 {
				t.Fatalf("unexpected JWT method: %v", token.Method.Alg())
			}
			if token.Header["kid"] != "key-1" {
				t.Fatalf("unexpected key id: %#v", token.Header["kid"])
			}
			return &key.PublicKey, nil
		})
		if err != nil || !parsed.Valid {
			t.Fatalf("invalid APNs JWT: valid=%v err=%v", parsed.Valid, err)
		}
		if claims["iss"] != "team-1" || claims["iat"] != float64(now.Unix()) {
			t.Fatalf("unexpected JWT claims: %#v", claims)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		var payload map[string]any
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Fatal(err)
		}
		if string(body) != `{"aps":{"content-available":1}}` || payload["alert"] != nil {
			t.Fatalf("unexpected APNs payload: %s", body)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	transport, err := NewAPNsTransport(Config{
		Enabled: true, TeamID: "team-1", KeyID: "key-1", Topic: "com.example.mypocket",
		PrivateKey: string(privateKeyPEM), Environment: "sandbox",
	}, APNsTransportOptions{Client: server.Client(), BaseURL: server.URL, Now: func() time.Time { return now }})
	if err != nil {
		t.Fatal(err)
	}
	response, err := transport.Send(context.Background(), Request{
		NotificationID: "notification-1", TokenID: "token-1", DeviceToken: "device-token",
		CollapseID: "notification-1", Payload: []byte(`{"aps":{"content-available":1}}`),
	})
	if err != nil || response.StatusCode != http.StatusOK {
		t.Fatalf("APNs send failed: response=%+v err=%v", response, err)
	}
}

func TestAPNsTransportParsesRetryAfterWithoutReadingResponseBodyIntoErrors(t *testing.T) {
	_, privateKeyPEM := testAPNsKey(t)
	now := time.Date(2026, 11, 1, 3, 4, 5, 0, time.UTC)
	for _, test := range []struct {
		name       string
		status     int
		header     string
		wantRetry  time.Duration
		wantReason string
	}{
		{name: "seconds", status: http.StatusTooManyRequests, header: "7", wantRetry: 7 * time.Second, wantReason: "apns_http_429"},
		{name: "http-date", status: http.StatusServiceUnavailable, header: now.Add(9 * time.Second).Format(http.TimeFormat), wantRetry: 9 * time.Second, wantReason: "apns_http_503"},
		{name: "bad-device-token", status: http.StatusBadRequest, header: "not-a-duration", wantRetry: 0, wantReason: "apns_http_400"},
		{name: "forbidden", status: http.StatusForbidden, header: "", wantRetry: 0, wantReason: "apns_http_403"},
	} {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if test.header != "" {
					w.Header().Set("Retry-After", test.header)
				}
				w.WriteHeader(test.status)
				_, _ = w.Write([]byte(`{"reason":"BadDeviceToken","secret":"must-not-escape"}`))
			}))
			defer server.Close()
			transport, err := NewAPNsTransport(Config{
				Enabled: true, TeamID: "team-1", KeyID: "key-1", Topic: "com.example.mypocket",
				PrivateKey: string(privateKeyPEM), Environment: "sandbox",
			}, APNsTransportOptions{Client: server.Client(), BaseURL: server.URL, Now: func() time.Time { return now }})
			if err != nil {
				t.Fatal(err)
			}
			response, err := transport.Send(context.Background(), Request{NotificationID: "notice", TokenID: "token", DeviceToken: "device-token", Payload: []byte(`{"aps":{}}`)})
			if err != nil || response.StatusCode != test.status || response.RetryAfter != test.wantRetry || response.FailureCode != test.wantReason {
				t.Fatalf("unexpected response=%+v err=%v", response, err)
			}
			if strings.Contains(errString(err), "must-not-escape") || strings.Contains(response.FailureCode, "secret") {
				t.Fatal("APNs response body leaked into transport error")
			}
		})
	}
}

func TestAPNsTransportUsesEnvironmentEndpointAndRejectsMalformedPrivateKey(t *testing.T) {
	if got := APNsEndpoint("sandbox"); got != "https://api.sandbox.push.apple.com" {
		t.Fatalf("unexpected sandbox endpoint: %q", got)
	}
	if got := APNsEndpoint("production"); got != "https://api.push.apple.com" {
		t.Fatalf("unexpected production endpoint: %q", got)
	}
	if _, err := NewAPNsTransport(Config{Enabled: true, TeamID: "team", KeyID: "key", Topic: "topic", PrivateKey: "not-a-key", Environment: "sandbox"}, APNsTransportOptions{}); err == nil {
		t.Fatal("malformed private key must disable transport construction")
	}
}

func testAPNsKey(t *testing.T) (*ecdsa.PrivateKey, []byte) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	return key, pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
