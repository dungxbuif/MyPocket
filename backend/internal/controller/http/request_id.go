package httpapi

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	requestIDHeader     = "X-Request-ID"
	requestIDContextKey = "requestID"
)

func requestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader(requestIDHeader)
		if id == "" {
			id = uuid.NewString()
		}
		c.Set(requestIDContextKey, id)
		c.Header(requestIDHeader, id)
		c.Next()
	}
}
