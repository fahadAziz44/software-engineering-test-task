package middleware

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func NewStructuredLogger() *slog.Logger {
	// This is standard practice for containerized apps.
	return slog.New(slog.NewJSONHandler(os.Stdout, nil))
}

// RequestLogger is a Gin middleware for structured (JSON) logging.
// It logs key information about each request.
func RequestLogger(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// For tracing purposes
		requestID := uuid.New().String()

		reqLogger := logger.With(slog.String("request_id", requestID))

		c.Set("logger", reqLogger)
		c.Set("requestID", requestID)

		// Add requestID to the response header so the client can see it
		c.Writer.Header().Set("X-Request-ID", requestID)

		// Process the request
		c.Next()

		// --- Log *after* the request is handled ---

		latency := time.Since(start)
		status := c.Writer.Status()

		// Log errors specifically
		if len(c.Errors) > 0 {
			// Log the last error
			reqLogger.Error(
				"Request failed",
				slog.String("method", c.Request.Method),
				slog.String("path", c.Request.URL.Path),
				slog.Int("status_code", status),
				slog.Duration("latency", latency),
				slog.String("client_ip", c.ClientIP()),
				slog.String("user_agent", c.Request.UserAgent()),
				slog.String("error", c.Errors.String()),
			)
		} else {
			// Log success with appropriate level based on status code
			msg := "Request completed"
			if status >= http.StatusInternalServerError {
				reqLogger.Error(msg,
					slog.String("method", c.Request.Method),
					slog.String("path", c.Request.URL.Path),
					slog.Int("status_code", status),
					slog.Duration("latency", latency),
					slog.String("client_ip", c.ClientIP()),
					slog.String("user_agent", c.Request.UserAgent()),
				)
			} else if status >= http.StatusBadRequest {
				reqLogger.Warn(msg,
					slog.String("method", c.Request.Method),
					slog.String("path", c.Request.URL.Path),
					slog.Int("status_code", status),
					slog.Duration("latency", latency),
					slog.String("client_ip", c.ClientIP()),
					slog.String("user_agent", c.Request.UserAgent()),
				)
			} else {
				reqLogger.Info(msg,
					slog.String("method", c.Request.Method),
					slog.String("path", c.Request.URL.Path),
					slog.Int("status_code", status),
					slog.Duration("latency", latency),
					slog.String("client_ip", c.ClientIP()),
					slog.String("user_agent", c.Request.UserAgent()),
				)
			}
		}
	}
}
