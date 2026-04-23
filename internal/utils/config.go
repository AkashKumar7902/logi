package utils

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Environment               string   `yaml:"environment"`
	ServerAddress             string   `yaml:"server_address"`
	MongoURI                  string   `yaml:"mongo_uri"`
	JWTSecret                 string   `yaml:"jwt_secret"`
	JWTExpirationHours        int      `yaml:"jwt_expiration_hours"`
	MessagingType             string   `yaml:"messaging_type"`
	NATSURL                   string   `yaml:"nats_url"`
	DistanceCalculatorType    string   `yaml:"distance_calculator_type"`
	GoogleMapsAPIKey          string   `yaml:"google_maps_api_key"`
	AllowedOrigins            []string `yaml:"allowed_origins"`
	EnableAdminBootstrap      bool     `yaml:"enable_admin_bootstrap"`
	AdminBootstrapSecret      string   `yaml:"admin_bootstrap_secret"`
	EnableTestRoutes          bool     `yaml:"enable_test_routes"`
	DBOperationTimeoutSeconds int      `yaml:"db_operation_timeout_seconds"`
	HTTPReadTimeoutSeconds    int      `yaml:"http_read_timeout_seconds"`
	HTTPWriteTimeoutSeconds   int      `yaml:"http_write_timeout_seconds"`
	HTTPIdleTimeoutSeconds    int      `yaml:"http_idle_timeout_seconds"`
	ShutdownTimeoutSeconds    int      `yaml:"shutdown_timeout_seconds"`
}

func LoadConfig(path string) (*Config, error) {
	config := defaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
	} else {
		if err := yaml.Unmarshal(data, &config); err != nil {
			return nil, err
		}
	}

	overrideFromEnv(&config)
	if err := validateConfig(&config); err != nil {
		return nil, err
	}
	return &config, nil
}

func (c *Config) AllowedOriginsSet() map[string]struct{} {
	out := make(map[string]struct{}, len(c.AllowedOrigins))
	for _, origin := range c.AllowedOrigins {
		trimmed := strings.TrimSpace(origin)
		if trimmed != "" {
			out[trimmed] = struct{}{}
		}
	}
	return out
}

func defaultConfig() Config {
	return Config{
		Environment:               "development",
		ServerAddress:             ":8080",
		JWTExpirationHours:        72,
		MessagingType:             "websocket",
		DistanceCalculatorType:    "haversine",
		AllowedOrigins:            []string{"http://localhost:3000"},
		EnableAdminBootstrap:      false,
		EnableTestRoutes:          false,
		DBOperationTimeoutSeconds: 5,
		HTTPReadTimeoutSeconds:    15,
		HTTPWriteTimeoutSeconds:   30,
		HTTPIdleTimeoutSeconds:    60,
		ShutdownTimeoutSeconds:    15,
	}
}

func overrideFromEnv(cfg *Config) {
	applyStringEnvWithFallback(&cfg.Environment, "LOGI_ENVIRONMENT", "ENVIRONMENT")
	applyStringEnv(&cfg.ServerAddress, "LOGI_SERVER_ADDRESS")
	if !hasNonEmptyEnv("LOGI_SERVER_ADDRESS") {
		applyPortEnv(&cfg.ServerAddress, "PORT")
	}
	applyStringEnvWithFallback(&cfg.MongoURI, "LOGI_MONGO_URI", "MONGODB_URI", "MONGO_URI")
	applyStringEnvWithFallback(&cfg.JWTSecret, "LOGI_JWT_SECRET", "JWT_SECRET")
	applyIntEnv(&cfg.JWTExpirationHours, "LOGI_JWT_EXPIRATION_HOURS")
	applyStringEnvWithFallback(&cfg.MessagingType, "LOGI_MESSAGING_TYPE")
	applyStringEnvWithFallback(&cfg.NATSURL, "LOGI_NATS_URL", "NATS_URL")
	applyStringEnvWithFallback(&cfg.DistanceCalculatorType, "LOGI_DISTANCE_CALCULATOR_TYPE")
	applyStringEnvWithFallback(&cfg.GoogleMapsAPIKey, "LOGI_GOOGLE_MAPS_API_KEY", "GOOGLE_MAPS_API_KEY")
	applyCSVEnvWithFallback(&cfg.AllowedOrigins, "LOGI_ALLOWED_ORIGINS", "ALLOWED_ORIGINS")
	applyBoolEnv(&cfg.EnableAdminBootstrap, "LOGI_ENABLE_ADMIN_BOOTSTRAP")
	applyStringEnv(&cfg.AdminBootstrapSecret, "LOGI_ADMIN_BOOTSTRAP_SECRET")
	applyBoolEnv(&cfg.EnableTestRoutes, "LOGI_ENABLE_TEST_ROUTES")
	applyIntEnv(&cfg.DBOperationTimeoutSeconds, "LOGI_DB_OPERATION_TIMEOUT_SECONDS")
	applyIntEnv(&cfg.HTTPReadTimeoutSeconds, "LOGI_HTTP_READ_TIMEOUT_SECONDS")
	applyIntEnv(&cfg.HTTPWriteTimeoutSeconds, "LOGI_HTTP_WRITE_TIMEOUT_SECONDS")
	applyIntEnv(&cfg.HTTPIdleTimeoutSeconds, "LOGI_HTTP_IDLE_TIMEOUT_SECONDS")
	applyIntEnv(&cfg.ShutdownTimeoutSeconds, "LOGI_SHUTDOWN_TIMEOUT_SECONDS")
}

func validateConfig(cfg *Config) error {
	if strings.TrimSpace(cfg.MongoURI) == "" {
		return fmt.Errorf("mongo_uri is required (set LOGI_MONGO_URI, MONGODB_URI, or MONGO_URI)")
	}
	if isCloudRuntime() && isLocalMongoURI(cfg.MongoURI) {
		return fmt.Errorf("mongo_uri points to localhost in cloud runtime; set LOGI_MONGO_URI or MONGODB_URI to your external MongoDB connection string")
	}
	if len(strings.TrimSpace(cfg.JWTSecret)) < 32 {
		return fmt.Errorf("jwt_secret must be at least 32 characters for production safety")
	}
	if (cfg.Environment == "production" || isCloudRuntime()) && isPlaceholderSecret(cfg.JWTSecret) {
		return fmt.Errorf("jwt_secret placeholder detected; set LOGI_JWT_SECRET or JWT_SECRET to a strong random value")
	}
	if cfg.EnableAdminBootstrap && len(strings.TrimSpace(cfg.AdminBootstrapSecret)) < 32 {
		return fmt.Errorf("admin_bootstrap_secret must be at least 32 characters when admin bootstrap is enabled (set LOGI_ADMIN_BOOTSTRAP_SECRET)")
	}

	switch cfg.MessagingType {
	case "websocket", "nats":
	default:
		return fmt.Errorf("messaging_type must be one of: websocket, nats")
	}
	if cfg.MessagingType == "nats" && strings.TrimSpace(cfg.NATSURL) == "" {
		return fmt.Errorf("nats_url is required when messaging_type is nats")
	}

	switch cfg.DistanceCalculatorType {
	case "haversine", "google_maps", "":
	default:
		return fmt.Errorf("distance_calculator_type must be one of: haversine, google_maps")
	}
	if cfg.DistanceCalculatorType == "google_maps" && strings.TrimSpace(cfg.GoogleMapsAPIKey) == "" {
		return fmt.Errorf("google_maps_api_key is required when distance_calculator_type is google_maps")
	}

	if cfg.DBOperationTimeoutSeconds <= 0 || cfg.HTTPReadTimeoutSeconds <= 0 || cfg.HTTPWriteTimeoutSeconds <= 0 || cfg.HTTPIdleTimeoutSeconds <= 0 || cfg.ShutdownTimeoutSeconds <= 0 {
		return fmt.Errorf("db/http/shutdown timeout values must be greater than 0")
	}

	return nil
}

func applyStringEnv(target *string, key string) {
	val := strings.TrimSpace(os.Getenv(key))
	if val != "" {
		*target = val
	}
}

func applyStringEnvWithFallback(target *string, keys ...string) {
	for _, key := range keys {
		if val := strings.TrimSpace(os.Getenv(key)); val != "" {
			*target = val
			return
		}
	}
}

func applyIntEnv(target *int, key string) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return
	}
	parsed, err := strconv.Atoi(raw)
	if err == nil {
		*target = parsed
	}
}

func applyBoolEnv(target *bool, key string) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return
	}
	parsed, err := strconv.ParseBool(raw)
	if err == nil {
		*target = parsed
	}
}

func applyCSVEnv(target *[]string, key string) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return
	}
	applyCSVValue(target, raw)
}

func applyCSVEnvWithFallback(target *[]string, keys ...string) {
	for _, key := range keys {
		if raw := strings.TrimSpace(os.Getenv(key)); raw != "" {
			applyCSVValue(target, raw)
			return
		}
	}
}

func applyCSVValue(target *[]string, raw string) {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			out = append(out, trimmed)
		}
	}
	*target = out
}

func applyPortEnv(target *string, key string) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return
	}
	if strings.Contains(raw, ":") {
		*target = raw
		return
	}
	*target = ":" + raw
}

func hasNonEmptyEnv(key string) bool {
	return strings.TrimSpace(os.Getenv(key)) != ""
}

func isCloudRuntime() bool {
	return hasNonEmptyEnv("PORT")
}

func isLocalMongoURI(uri string) bool {
	lower := strings.ToLower(strings.TrimSpace(uri))
	return strings.Contains(lower, "localhost") || strings.Contains(lower, "127.0.0.1") || strings.Contains(lower, "[::1]")
}

func isPlaceholderSecret(secret string) bool {
	return strings.Contains(strings.ToLower(strings.TrimSpace(secret)), "replace-with-a-strong-secret")
}
