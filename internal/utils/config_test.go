package utils

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadConfigUsesCloudEnvFallbacks(t *testing.T) {
	t.Setenv("PORT", "10000")
	t.Setenv("MONGODB_URI", "mongodb+srv://user:pass@example.mongodb.net/logi")
	t.Setenv("JWT_SECRET", "0123456789abcdef0123456789abcdef")
	t.Setenv("ALLOWED_ORIGINS", "https://frontend.example.com,https://admin.example.com")

	configPath := filepath.Join(t.TempDir(), "config.yaml")
	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}

	if cfg.ServerAddress != ":10000" {
		t.Fatalf("expected server address :10000, got %q", cfg.ServerAddress)
	}
	if cfg.MongoURI != "mongodb+srv://user:pass@example.mongodb.net/logi" {
		t.Fatalf("expected mongo uri from fallback env, got %q", cfg.MongoURI)
	}
	if cfg.JWTSecret != "0123456789abcdef0123456789abcdef" {
		t.Fatalf("expected jwt secret from fallback env, got %q", cfg.JWTSecret)
	}
	if len(cfg.AllowedOrigins) != 2 || cfg.AllowedOrigins[0] != "https://frontend.example.com" || cfg.AllowedOrigins[1] != "https://admin.example.com" {
		t.Fatalf("expected allowed origins from fallback env, got %#v", cfg.AllowedOrigins)
	}
}

func TestLoadConfigRejectsLocalMongoInCloudRuntime(t *testing.T) {
	t.Setenv("PORT", "10000")
	t.Setenv("JWT_SECRET", "0123456789abcdef0123456789abcdef")

	configPath := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(configPath, []byte("mongo_uri: mongodb://localhost:27017/logi\n"), 0o644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	_, err := LoadConfig(configPath)
	if err == nil {
		t.Fatal("expected config validation error for localhost mongo uri in cloud runtime")
	}
	if !strings.Contains(err.Error(), "localhost") {
		t.Fatalf("expected localhost guidance in error, got %v", err)
	}
}

func TestLoadConfigRejectsPlaceholderJWTSecretInCloudRuntime(t *testing.T) {
	t.Setenv("PORT", "10000")
	t.Setenv("MONGODB_URI", "mongodb+srv://user:pass@example.mongodb.net/logi")

	configPath := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(configPath, []byte("jwt_secret: replace-with-a-strong-secret-at-least-32-characters\n"), 0o644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	_, err := LoadConfig(configPath)
	if err == nil {
		t.Fatal("expected config validation error for placeholder jwt secret in cloud runtime")
	}
	if !strings.Contains(err.Error(), "JWT_SECRET") {
		t.Fatalf("expected JWT env guidance in error, got %v", err)
	}
}

func TestLoadConfigReadsAdminBootstrapSettingsFromEnv(t *testing.T) {
	t.Setenv("MONGODB_URI", "mongodb+srv://user:pass@example.mongodb.net/logi")
	t.Setenv("JWT_SECRET", "0123456789abcdef0123456789abcdef")
	t.Setenv("LOGI_ENABLE_ADMIN_BOOTSTRAP", "true")
	t.Setenv("LOGI_ADMIN_BOOTSTRAP_SECRET", "bootstrap-secret-0123456789abcdef")

	configPath := filepath.Join(t.TempDir(), "config.yaml")
	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}

	if !cfg.EnableAdminBootstrap {
		t.Fatal("expected admin bootstrap to be enabled from env")
	}
	if cfg.AdminBootstrapSecret != "bootstrap-secret-0123456789abcdef" {
		t.Fatalf("expected admin bootstrap secret from env, got %q", cfg.AdminBootstrapSecret)
	}
}

func TestLoadConfigRequiresAdminBootstrapSecretWhenEnabled(t *testing.T) {
	t.Setenv("MONGODB_URI", "mongodb+srv://user:pass@example.mongodb.net/logi")
	t.Setenv("JWT_SECRET", "0123456789abcdef0123456789abcdef")
	t.Setenv("LOGI_ENABLE_ADMIN_BOOTSTRAP", "true")

	configPath := filepath.Join(t.TempDir(), "config.yaml")
	_, err := LoadConfig(configPath)
	if err == nil {
		t.Fatal("expected config validation error when admin bootstrap is enabled without a secret")
	}
	if !strings.Contains(err.Error(), "LOGI_ADMIN_BOOTSTRAP_SECRET") {
		t.Fatalf("expected admin bootstrap env guidance in error, got %v", err)
	}
}
