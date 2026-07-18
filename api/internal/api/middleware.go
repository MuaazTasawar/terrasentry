package api

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// RequestLogger logs method, path, status, and latency for every request.
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		log.Printf("%s %s -> %d (%s)", c.Request.Method, c.Request.URL.Path, c.Writer.Status(), time.Since(start))
	}
}

// ErrorHandler recovers from panics in handlers and returns a clean 500
// instead of crashing the process.
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic recovered: %v", err)
				c.JSON(500, gin.H{"error": "internal server error"})
				c.Abort()
			}
		}()
		c.Next()
	}
}
