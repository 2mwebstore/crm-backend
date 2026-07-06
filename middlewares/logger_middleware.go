package middlewares

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// Logger logs method, path, status, and latency for every request.
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		method := c.Request.Method
		clientIP := c.ClientIP()

		log.Printf("[CRM] %s | %3d | %13v | %15s | %s",
			method, status, latency, clientIP, path,
		)
	}
}
