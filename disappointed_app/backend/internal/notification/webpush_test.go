package notification

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	webpush "github.com/SherClockHolmes/webpush-go"
)

const testP256DH = "BNNL5ZaTfK81qhXOx23-wewhigUeFb632jN6LvRWCFH1ubQr77FE_9qV1FuojuRmHP42zmf34rXgW80OvUVDgTk"
const testPushAuth = "zqbxT6JKstKSY9JKibZLSQ"

func TestWebPushDeliverySendsEncryptedNotice(t *testing.T) {
	privateKey, publicKey, err := webpush.GenerateVAPIDKeys()
	if err != nil {
		t.Fatal(err)
	}
	var request *http.Request
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		request = r.Clone(context.Background())
		body, readErr := io.ReadAll(r.Body)
		if readErr != nil || len(body) == 0 {
			t.Errorf("expected encrypted body, len=%d err=%v", len(body), readErr)
		}
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	delivery, err := NewWebPushDelivery(publicKey, privateKey, "owner@example.com", server.Client())
	if err != nil {
		t.Fatalf("create delivery: %v", err)
	}
	if err := delivery.Send(context.Background(), DeliverySubscription{Endpoint: server.URL, P256DH: testP256DH, Auth: testPushAuth}, Notice{ID: "notice-1", Title: "Ngân sách", Body: "Đã chạm 80%"}); err != nil {
		t.Fatalf("send: %v", err)
	}
	if request == nil || request.Header.Get("Content-Encoding") != "aes128gcm" || !strings.HasPrefix(request.Header.Get("Authorization"), "vapid ") {
		t.Fatalf("missing Web Push headers: %#v", request)
	}
}

func TestWebPushDeliveryClassifiesExpiredWithoutLeakingSubscriptionSecrets(t *testing.T) {
	privateKey, publicKey, err := webpush.GenerateVAPIDKeys()
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusGone) }))
	defer server.Close()
	delivery, err := NewWebPushDelivery(publicKey, privateKey, "owner@example.com", server.Client())
	if err != nil {
		t.Fatal(err)
	}

	err = delivery.Send(context.Background(), DeliverySubscription{Endpoint: server.URL + "/private-endpoint", P256DH: testP256DH, Auth: testPushAuth}, Notice{ID: "notice-2", Title: "Draft", Body: "Review"})
	if !IsExpiredDeliveryError(err) {
		t.Fatalf("expected expired delivery error, got %v", err)
	}
	for _, secret := range []string{"private-endpoint", testP256DH, testPushAuth} {
		if strings.Contains(err.Error(), secret) {
			t.Fatalf("delivery error leaked subscription secret %q: %v", secret, err)
		}
	}
	var deliveryErr *DeliveryError
	if !errors.As(err, &deliveryErr) || deliveryErr.StatusCode != http.StatusGone {
		t.Fatalf("expected typed 410 error, got %#v", err)
	}
}

func TestNewWebPushDeliveryRejectsPartialConfiguration(t *testing.T) {
	if _, err := NewWebPushDelivery("public", "", "owner@example.com", nil); err == nil {
		t.Fatal("expected partial Web Push configuration rejection")
	}
}
