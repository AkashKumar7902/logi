package utils

import (
	"fmt"
	"net/http"
	"net/url"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const requestIDGinKey = "request_id"

func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := strings.TrimSpace(c.GetHeader(RequestIDHeader))
		if requestID == "" {
			requestID = uuid.NewString()
		}

		c.Set(requestIDGinKey, requestID)
		c.Request = c.Request.WithContext(WithRequestID(c.Request.Context(), requestID))
		c.Writer.Header().Set(RequestIDHeader, requestID)

		c.Next()
	}
}

func RequestIDFromGin(c *gin.Context) string {
	if c == nil {
		return ""
	}

	if requestID := c.GetString(requestIDGinKey); requestID != "" {
		return requestID
	}

	return RequestIDFromContext(c.Request.Context())
}

func RequestLoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := redactedRawQuery(c.Request.URL.RawQuery)

		c.Next()

		attrs := []any{
			"method", c.Request.Method,
			"path", path,
			"status", c.Writer.Status(),
			"latency_ms", time.Since(start).Milliseconds(),
			"client_ip", c.ClientIP(),
			"user_agent", c.Request.UserAgent(),
			"response_size", c.Writer.Size(),
		}

		if route := c.FullPath(); route != "" {
			attrs = append(attrs, "route", route)
		}
		if query != "" {
			attrs = append(attrs, "query", query)
		}
		if userID := c.GetString("userID"); userID != "" {
			attrs = append(attrs, "user_id", userID)
		}
		if role := c.GetString("role"); role != "" {
			attrs = append(attrs, "role", role)
		}
		if len(c.Errors) > 0 {
			attrs = append(attrs, "errors", c.Errors.String())
		}

		switch {
		case len(c.Errors) > 0 || c.Writer.Status() >= http.StatusInternalServerError:
			Error(c.Request.Context(), "http request completed", attrs...)
		case c.Writer.Status() >= http.StatusBadRequest:
			Warn(c.Request.Context(), "http request completed", attrs...)
		default:
			Info(c.Request.Context(), "http request completed", attrs...)
		}
	}
}

func redactedRawQuery(rawQuery string) string {
	if rawQuery == "" {
		return ""
	}

	values, err := url.ParseQuery(rawQuery)
	if err != nil {
		return "[unparseable]"
	}

	for key := range values {
		if isSensitiveQueryKey(key) {
			values[key] = []string{"[REDACTED]"}
		}
	}
	return values.Encode()
}

func isSensitiveQueryKey(key string) bool {
	normalized := strings.ToLower(strings.TrimSpace(key))
	return normalized == "token" ||
		normalized == "access_token" ||
		normalized == "id_token" ||
		normalized == "refresh_token" ||
		normalized == "password" ||
		normalized == "secret" ||
		strings.HasSuffix(normalized, "_secret")
}

func RecoveryMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered any) {
		Error(
			c.Request.Context(),
			"panic recovered",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"panic", fmt.Sprintf("%v", recovered),
			"stack", string(debug.Stack()),
		)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
	})
}
