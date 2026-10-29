package identity

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
)

const (
	CSRFCookieName = "mypocket_csrf"
	CSRFHeaderName = "X-CSRF-Token"
)

func NewCSRFToken() (string, error) {
	var bytes [32]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes[:]), nil
}

func RequireCSRF(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(CSRFCookieName)
		header := r.Header.Get(CSRFHeaderName)
		if err != nil || header == "" || cookie.Value == "" || header != cookie.Value {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"error": map[string]string{
					"code":    "CSRF_REQUIRED",
					"message": "CSRF token is required",
				},
			})
			return
		}
		next.ServeHTTP(w, r)
	})
}
