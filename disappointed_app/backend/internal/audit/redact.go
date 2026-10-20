package audit

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
)

func HashValue(secret string, value string) string {
	value = strings.TrimSpace(value)
	if secret == "" || value == "" {
		return ""
	}
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(strings.ToLower(value)))
	return hex.EncodeToString(mac.Sum(nil))
}

func SafeMetadata(values map[string]any) json.RawMessage {
	if len(values) == 0 {
		return json.RawMessage(`{}`)
	}
	allowed := map[string]any{}
	for key, value := range values {
		normalized := strings.ToLower(strings.TrimSpace(key))
		if safeMetadataKey(normalized) {
			allowed[normalized] = value
		}
	}
	body, err := json.Marshal(allowed)
	if err != nil || len(body) == 0 {
		return json.RawMessage(`{}`)
	}
	return json.RawMessage(body)
}

func safeMetadataKey(key string) bool {
	switch key {
	case "status", "count", "mutation_id", "entity_type", "operation", "version", "reason", "processed", "refreshed", "has_error":
		return true
	default:
		return false
	}
}
