package httpapi

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Data any `json:"data"`
	Meta any `json:"meta"`
}
type Problem struct {
	Type      string `json:"type,omitempty"`
	Title     string `json:"title"`
	Status    int    `json:"status,omitempty"`
	Detail    string `json:"detail"`
	Instance  string `json:"instance,omitempty"`
	Code      string `json:"code"`
	RequestID string `json:"request_id,omitempty"`
}

func OK(c *gin.Context, data any)      { c.JSON(http.StatusOK, Response{Data: data, Meta: nil}) }
func Created(c *gin.Context, data any) { c.JSON(http.StatusCreated, Response{Data: data, Meta: nil}) }
func NoContent(c *gin.Context)         { c.Status(http.StatusNoContent) }
func Fail(c *gin.Context, status int, problem Problem) {
	problem.Status = status
	problem.Instance = c.Request.URL.Path
	problem.RequestID = requestID(c)
	c.Header("Content-Type", problemContentType)
	c.AbortWithStatusJSON(status, problem)
}

func FailError(c *gin.Context, status int, code string, title string, err error) {
	detail := "request failed"
	if err != nil {
		detail = err.Error()
	}
	Fail(c, status, Problem{Code: code, Title: title, Detail: detail})
}

func requestID(c *gin.Context) string {
	value, _ := c.Get(requestIDContextKey)
	if id, ok := value.(string); ok {
		return id
	}
	return c.GetHeader(requestIDHeader)
}
