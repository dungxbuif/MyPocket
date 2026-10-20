package notification

import (
	"context"
	"time"
)

const MaxPageSize = 50
const MaxDeliveryAttempts = 5

type Notice struct {
	ID         string     `json:"id"`
	UserID     string     `json:"user_id"`
	Kind       string     `json:"kind"`
	Title      string     `json:"title"`
	Body       string     `json:"body"`
	SourceType string     `json:"source_type"`
	SourceID   string     `json:"source_id"`
	ReadAt     *time.Time `json:"read_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

type PushSubscription struct {
	ID        string     `json:"id"`
	Endpoint  string     `json:"endpoint"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

type CreateNoticeInput struct {
	UserID     string
	Kind       string
	Title      string
	Body       string
	SourceType string
	SourceID   string
	DedupeKey  string
}

type CreatePushSubscriptionInput struct {
	Endpoint  string
	P256DH    string
	Auth      string
	ExpiresAt *time.Time
}

type DeliverySubscription struct {
	UserID   string
	ID       string
	Endpoint string
	P256DH   string
	Auth     string
}

type Delivery interface {
	Send(ctx context.Context, subscription DeliverySubscription, notice Notice) error
}
