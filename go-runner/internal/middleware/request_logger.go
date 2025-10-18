package middleware

import (
	"bytes"
	"io"
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// ResponseWriter wraps gin.ResponseWriter to capture response body
type ResponseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w ResponseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// RequestLogger provides detailed request/response logging
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		
		// Capture request body
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// Wrap response writer to capture response body
		blw := &ResponseWriter{
			ResponseWriter: c.Writer,
			body:          bytes.NewBufferString(""),
		}
		c.Writer = blw

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start)

		// Log request details
		log.Printf("[%s] %s %s %d %v %s %s %s",
			start.Format("2006-01-02 15:04:05"),
			c.Request.Method,
			c.Request.URL.Path,
			c.Writer.Status(),
			duration,
			c.ClientIP(),
			c.Request.UserAgent(),
			c.Request.Referer(),
		)

		// Log request/response bodies in debug mode
		if gin.Mode() == gin.DebugMode {
			if len(requestBody) > 0 {
				log.Printf("Request Body: %s", string(requestBody))
			}
			// if blw.body.Len() > 0 {
			// 	log.Printf("Response Body: %s", blw.body.String())
			// }
		}
	}
}

// ErrorLogger logs errors with context
func ErrorLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Check for errors
		if len(c.Errors) > 0 {
			for _, err := range c.Errors {
				log.Printf("Error: %v | Path: %s | Method: %s | IP: %s",
					err.Error(),
					c.Request.URL.Path,
					c.Request.Method,
					c.ClientIP(),
				)
			}
		}
	}
}

// SecurityHeaders adds security headers
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Next()
	}
}

// RateLimiter provides basic rate limiting
func RateLimiter() gin.HandlerFunc {
	// Simple in-memory rate limiter
	// In production, use Redis or similar
	requests := make(map[string][]time.Time)
	
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		now := time.Now()
		
		// Clean old requests (older than 1 minute)
		if clientRequests, exists := requests[clientIP]; exists {
			var validRequests []time.Time
			for _, reqTime := range clientRequests {
				if now.Sub(reqTime) < time.Minute {
					validRequests = append(validRequests, reqTime)
				}
			}
			requests[clientIP] = validRequests
		}

		// Check rate limit (100 requests per minute)
		if len(requests[clientIP]) >= 100 {
			c.JSON(429, gin.H{
				"error": "Too Many Requests",
				"message": "Rate limit exceeded",
			})
			c.Abort()
			return
		}

		// Add current request
		requests[clientIP] = append(requests[clientIP], now)
		c.Next()
	}
}
