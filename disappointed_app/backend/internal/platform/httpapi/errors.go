package httpapi

import "strings"

type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type Envelope struct {
	Error         *ErrorBody        `json:"error,omitempty"`
	Status        string            `json:"status,omitempty"`
	Checks        map[string]string `json:"checks,omitempty"`
	CorrelationID string            `json:"correlation_id"`
}

func ErrorEnvelope(code string, message string, correlationID string) Envelope {
	return Envelope{
		Error: &ErrorBody{
			Code:    code,
			Message: safeMessage(code, message),
		},
		CorrelationID: correlationID,
	}
}

func safeMessage(code string, message string) string {
	if strings.HasPrefix(code, "INTERNAL_") {
		return "A temporary internal error occurred"
	}
	message = strings.TrimSpace(message)
	if message == "" {
		return "Request failed"
	}
	return message
}
