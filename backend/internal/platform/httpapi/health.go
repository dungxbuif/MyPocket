package httpapi

import "net/http"

func liveHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, ErrorEnvelope("VALIDATION_FAILED", "Method not allowed", correlationID(r.Context())))
		return
	}

	writeJSON(w, http.StatusOK, Envelope{
		Status:        "ok",
		CorrelationID: correlationID(r.Context()),
	})
}

func readyHealth(deps Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeJSON(w, http.StatusMethodNotAllowed, ErrorEnvelope("VALIDATION_FAILED", "Method not allowed", correlationID(r.Context())))
			return
		}
		if deps.ReadyCheck != nil {
			if err := deps.ReadyCheck(); err != nil {
				writeJSON(w, http.StatusServiceUnavailable, ErrorEnvelope("INTERNAL_RETRYABLE", err.Error(), correlationID(r.Context())))
				return
			}
		}

		writeJSON(w, http.StatusOK, Envelope{
			Status:        "ok",
			Checks:        map[string]string{"database": "ok"},
			CorrelationID: correlationID(r.Context()),
		})
	}
}
