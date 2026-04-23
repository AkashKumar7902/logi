package utils

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRequestIDMiddlewarePreservesIncomingHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(RequestIDMiddleware())
	router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, RequestIDFromContext(c.Request.Context()))
	})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set(RequestIDHeader, "req-incoming")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.Code)
	}

	if got := resp.Header().Get(RequestIDHeader); got != "req-incoming" {
		t.Fatalf("expected response request ID to be preserved, got %q", got)
	}

	if got := strings.TrimSpace(resp.Body.String()); got != "req-incoming" {
		t.Fatalf("expected request context to carry incoming request ID, got %q", got)
	}
}

func TestRequestIDMiddlewareGeneratesHeaderWhenMissing(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(RequestIDMiddleware())
	router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, RequestIDFromContext(c.Request.Context()))
	})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.Code)
	}

	requestID := resp.Header().Get(RequestIDHeader)
	if requestID == "" {
		t.Fatal("expected middleware to generate a request ID")
	}

	if got := strings.TrimSpace(resp.Body.String()); got != requestID {
		t.Fatalf("expected generated request ID to be stored in context, got %q", got)
	}
}

func TestRequestLoggingMiddlewareWritesJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var output bytes.Buffer
	originalLogger := baseLogger()
	setLogger(newLogger(&output, false))
	t.Cleanup(func() {
		setLogger(originalLogger)
	})

	router := gin.New()
	router.Use(RequestIDMiddleware())
	router.Use(RequestLoggingMiddleware())
	router.Use(RecoveryMiddleware())
	router.GET("/ping", func(c *gin.Context) {
		c.Set("userID", "user-123")
		c.Set("role", "driver")
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/ping?trace=true", nil)
	req.Header.Set(RequestIDHeader, "req-logged")
	req.Header.Set("User-Agent", "logger-test")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", resp.Code)
	}

	logLine := strings.TrimSpace(output.String())
	if logLine == "" {
		t.Fatal("expected request logging middleware to write a log line")
	}

	var payload map[string]any
	if err := json.Unmarshal([]byte(logLine), &payload); err != nil {
		t.Fatalf("expected JSON log output, got error: %v", err)
	}

	if got := payload["msg"]; got != "http request completed" {
		t.Fatalf("expected log message to match, got %#v", got)
	}
	if got := payload["request_id"]; got != "req-logged" {
		t.Fatalf("expected request_id in log payload, got %#v", got)
	}
	if got := payload["method"]; got != http.MethodGet {
		t.Fatalf("expected method in log payload, got %#v", got)
	}
	if got := payload["path"]; got != "/ping" {
		t.Fatalf("expected path in log payload, got %#v", got)
	}
	if got := payload["route"]; got != "/ping" {
		t.Fatalf("expected route in log payload, got %#v", got)
	}
	if got := payload["query"]; got != "trace=true" {
		t.Fatalf("expected query in log payload, got %#v", got)
	}
	if got := payload["user_id"]; got != "user-123" {
		t.Fatalf("expected user_id in log payload, got %#v", got)
	}
	if got := payload["role"]; got != "driver" {
		t.Fatalf("expected role in log payload, got %#v", got)
	}
}
