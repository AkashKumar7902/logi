package api

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"logi/internal/handlers"
	"logi/internal/utils"
	"logi/pkg/auth"
	"logi/pkg/websocket"
)

func newTestRouter(cfg *utils.Config, readinessChecks ...ReadinessCheck) http.Handler {
	return SetupRouter(
		&handlers.UserHandler{},
		&handlers.BookingHandler{},
		&handlers.DriverHandler{},
		&handlers.AdminHandler{},
		auth.NewAuthService(strings.Repeat("a", 32), 72),
		websocket.NewWebSocketHub(),
		&handlers.TestHandler{},
		cfg,
		readinessChecks...,
	)
}

func TestSetupRouterDoesNotExposePublicAdminRegistration(t *testing.T) {
	t.Parallel()

	router := newTestRouter(&utils.Config{
		AllowedOrigins: []string{"http://localhost:3000"},
	})

	req := httptest.NewRequest(http.MethodPost, "/admins/register", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected /admins/register to be unavailable, got status %d", recorder.Code)
	}
}

func TestReadyzReturnsOKWhenReadinessChecksPass(t *testing.T) {
	t.Parallel()

	router := newTestRouter(
		&utils.Config{AllowedOrigins: []string{"http://localhost:3000"}},
		func(ctx context.Context) error { return nil },
	)

	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected ready status 200, got %d", recorder.Code)
	}
}

func TestReadyzReturnsUnavailableWhenReadinessCheckFails(t *testing.T) {
	t.Parallel()

	router := newTestRouter(
		&utils.Config{AllowedOrigins: []string{"http://localhost:3000"}},
		func(ctx context.Context) error { return errors.New("mongo unavailable") },
	)

	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected ready status 503, got %d", recorder.Code)
	}
}

func TestSetupRouterRequiresSecretForAdminBootstrap(t *testing.T) {
	t.Parallel()

	router := newTestRouter(&utils.Config{
		AllowedOrigins:       []string{"http://localhost:3000"},
		EnableAdminBootstrap: true,
		AdminBootstrapSecret: "bootstrap-secret-0123456789abcdef",
	})

	req := httptest.NewRequest(http.MethodPost, "/internal/bootstrap/admin", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected bootstrap route to require a secret, got status %d", recorder.Code)
	}
}
