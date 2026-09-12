package notification

import (
	"errors"
	"net/http"
	"strings"
)

func IsExpiredDeliveryError(err error) bool {
	if err == nil {
		return false
	}
	var deliveryErr *DeliveryError
	if errors.As(err, &deliveryErr) {
		return deliveryErr.StatusCode == http.StatusNotFound || deliveryErr.StatusCode == http.StatusGone
	}
	return false
}

func RedactEndpoint(endpoint string) string {
	endpoint = strings.TrimSpace(endpoint)
	if len(endpoint) <= 16 {
		return "[redacted]"
	}
	return endpoint[:8] + "..." + endpoint[len(endpoint)-6:]
}
