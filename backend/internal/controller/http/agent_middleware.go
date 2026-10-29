package httpapi

import (
	"crypto/sha256"
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	feedbackAgentActorKindKey = "feedbackAgentActorKind"
	feedbackAgentActorKind    = "agent"
)

type FeedbackAgentMiddleware struct{ token []byte }

func NewFeedbackAgentMiddleware(token string) *FeedbackAgentMiddleware {
	return &FeedbackAgentMiddleware{token: []byte(strings.TrimSpace(token))}
}

func (m *FeedbackAgentMiddleware) RequireFeedbackAgent(c *gin.Context) {
	header := strings.TrimSpace(c.GetHeader(authHeader))
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || parts[0] != authScheme {
		Fail(c, http.StatusUnauthorized, Problem{Code: problemCodeAuthInvalidFormat, Title: problemTitleUnauthorized, Detail: invalidAuthFormatMsg})
		c.Abort()
		return
	}
	provided := strings.TrimSpace(parts[1])
	expectedHash := sha256.Sum256(m.token)
	providedHash := sha256.Sum256([]byte(provided))
	if len(m.token) == 0 || subtle.ConstantTimeCompare(expectedHash[:], providedHash[:]) != 1 {
		Fail(c, http.StatusUnauthorized, Problem{Code: problemCodeTokenInvalid, Title: problemTitleUnauthorized, Detail: invalidTokenMsg})
		c.Abort()
		return
	}
	c.Set(feedbackAgentActorKindKey, feedbackAgentActorKind)
	c.Next()
}
