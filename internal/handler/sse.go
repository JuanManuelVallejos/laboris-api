package handler

import (
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// streamSSE writes ch's values out as named SSE events until the channel is
// closed (server-side unsubscribe) or the client disconnects
// (c.Request.Context().Done()). Sends a periodic ping to keep intermediary
// proxies from treating the connection as idle and closing it.
func streamSSE[T any](c *gin.Context, ch <-chan T, eventName string) {
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")
	c.Writer.WriteHeader(http.StatusOK)
	c.Writer.Flush()

	heartbeat := time.NewTicker(20 * time.Second)
	defer heartbeat.Stop()
	ctx := c.Request.Context()

	c.Stream(func(w io.Writer) bool {
		select {
		case payload, ok := <-ch:
			if !ok {
				return false
			}
			c.SSEvent(eventName, payload)
			return true
		case <-heartbeat.C:
			c.SSEvent("ping", nil)
			return true
		case <-ctx.Done():
			return false
		}
	})
}
