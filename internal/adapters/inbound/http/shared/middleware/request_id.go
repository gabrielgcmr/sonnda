package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const requestIDHeader = "X-Request-ID"

// RequestID garante que cada request tenha um request_id e propaga em header.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid, _ := c.Get("request_id")
		requestID := strings.TrimSpace(toString(rid))
		if requestID == "" {
			requestID = strings.TrimSpace(c.GetHeader(requestIDHeader))
		}
		if requestID == "" {
			requestID = uuid.NewString()
		}

		c.Writer.Header().Set(requestIDHeader, requestID)
		c.Set("request_id", requestID)
		c.Next()
	}
}
