package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// MaxBodySize rejects any request whose body exceeds n bytes. Wraps the
// request body in an http.MaxBytesReader so oversized bodies fail as soon
// as a handler (or Gin's own binding/multipart parsing) tries to read past
// the limit, rather than being fully buffered into memory first.
func MaxBodySize(n int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, n)
		c.Next()
	}
}
