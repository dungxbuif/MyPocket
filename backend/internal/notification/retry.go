package notification

import "strings"

func IsExpiredDeliveryError(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "404") || strings.Contains(message, "410") || strings.Contains(message, "expired") || strings.Contains(message, "gone")
}

func RedactEndpoint(endpoint string) string {
	endpoint = strings.TrimSpace(endpoint)
	if len(endpoint) <= 16 {
		return "[redacted]"
	}
	return endpoint[:8] + "..." + endpoint[len(endpoint)-6:]
}
