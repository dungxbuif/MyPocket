package httpapi

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"

	"mypocket/internal/platform/config"
)

type Dependencies struct {
	ReadyCheck func() error
}

func NewRouter(cfg config.Config, deps Dependencies) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/health/live", liveHealth)
	mux.HandleFunc("/api/v1/health/ready", readyHealth(deps))

	return correlationMiddleware(mux)
}

func correlationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		correlationID := strings.TrimSpace(r.Header.Get("X-Correlation-ID"))
		if correlationID == "" {
			correlationID = newCorrelationID()
		}
		w.Header().Set("X-Correlation-ID", correlationID)
		next.ServeHTTP(w, r.WithContext(withCorrelationID(r.Context(), correlationID)))
	})
}

func writeJSON(w http.ResponseWriter, status int, body Envelope) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func newCorrelationID() string {
	var bytes [12]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return "req_unavailable"
	}
	return "req_" + hex.EncodeToString(bytes[:])
}
