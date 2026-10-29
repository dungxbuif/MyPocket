package notification

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	webpush "github.com/SherClockHolmes/webpush-go"
)

type WebPushDelivery struct {
	PublicKey  string
	PrivateKey string
	Subject    string
	Client     *http.Client
}

type DeliveryError struct {
	StatusCode int
	Err        error
}

func (e *DeliveryError) Error() string {
	if e.StatusCode > 0 {
		return fmt.Sprintf("web push returned HTTP %d", e.StatusCode)
	}
	return "web push transport failed"
}

func (e *DeliveryError) Unwrap() error { return e.Err }

func NewWebPushDelivery(publicKey, privateKey, subject string, client *http.Client) (*WebPushDelivery, error) {
	if strings.TrimSpace(publicKey) == "" || strings.TrimSpace(privateKey) == "" || strings.TrimSpace(subject) == "" {
		return nil, fmt.Errorf("complete VAPID configuration is required")
	}
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	return &WebPushDelivery{PublicKey: publicKey, PrivateKey: privateKey, Subject: subject, Client: client}, nil
}

func (d *WebPushDelivery) Send(ctx context.Context, subscription DeliverySubscription, notice Notice) error {
	payload, err := json.Marshal(struct {
		ID         string `json:"id"`
		Kind       string `json:"kind"`
		Title      string `json:"title"`
		Body       string `json:"body"`
		SourceType string `json:"source_type"`
		SourceID   string `json:"source_id"`
	}{notice.ID, notice.Kind, notice.Title, notice.Body, notice.SourceType, notice.SourceID})
	if err != nil {
		return &DeliveryError{Err: err}
	}
	response, err := webpush.SendNotificationWithContext(ctx, payload, &webpush.Subscription{
		Endpoint: subscription.Endpoint,
		Keys: webpush.Keys{
			P256dh: subscription.P256DH,
			Auth:   subscription.Auth,
		},
	}, &webpush.Options{
		HTTPClient:      d.Client,
		Subscriber:      d.Subject,
		VAPIDPublicKey:  d.PublicKey,
		VAPIDPrivateKey: d.PrivateKey,
		TTL:             60,
		Urgency:         webpush.UrgencyNormal,
	})
	if err != nil {
		return &DeliveryError{Err: err}
	}
	defer response.Body.Close()
	_, _ = io.Copy(io.Discard, response.Body)
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return &DeliveryError{StatusCode: response.StatusCode}
	}
	return nil
}
