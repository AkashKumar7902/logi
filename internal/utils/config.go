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
		EnableTestRoutes:          false,
		DBOperationTimeoutSeconds: 5,
		HTTPReadTimeoutSeconds:    15,
		HTTPWriteTimeoutSeconds:   30,
		HTTPIdleTimeoutSeconds:    60,
		ShutdownTimeoutSeconds:    15,
	}
}

func overrideFromEnv(cfg *Config) {
	applyStringEnv(&cfg.Environment, "LOGI_ENVIRONMENT")
	applyStringEnv(&cfg.ServerAddress, "LOGI_SERVER_ADDRESS")
	applyStringEnv(&cfg.MongoURI, "LOGI_MONGO_URI")
	applyStringEnv(&cfg.JWTSecret, "LOGI_JWT_SECRET")
	applyIntEnv(&cfg.JWTExpirationHours, "LOGI_JWT_EXPIRATION_HOURS")
	applyStringEnv(&cfg.MessagingType, "LOGI_MESSAGING_TYPE")
	applyStringEnv(&cfg.NATSURL, "LOGI_NATS_URL")
	applyStringEnv(&cfg.DistanceCalculatorType, "LOGI_DISTANCE_CALCULATOR_TYPE")
	applyStringEnv(&cfg.GoogleMapsAPIKey, "LOGI_GOOGLE_MAPS_API_KEY")
	applyCSVEnv(&cfg.AllowedOrigins, "LOGI_ALLOWED_ORIGINS")
	applyBoolEnv(&cfg.EnableTestRoutes, "LOGI_ENABLE_TEST_ROUTES")
	applyIntEnv(&cfg.DBOperationTimeoutSeconds, "LOGI_DB_OPERATION_TIMEOUT_SECONDS")
	applyIntEnv(&cfg.HTTPReadTimeoutSeconds, "LOGI_HTTP_READ_TIMEOUT_SECONDS")
	applyIntEnv(&cfg.HTTPWriteTimeoutSeconds, "LOGI_HTTP_WRITE_TIMEOUT_SECONDS")
	applyIntEnv(&cfg.HTTPIdleTimeoutSeconds, "LOGI_HTTP_IDLE_TIMEOUT_SECONDS")
	applyIntEnv(&cfg.ShutdownTimeoutSeconds, "LOGI_SHUTDOWN_TIMEOUT_SECONDS")
}

func validateConfig(cfg *Config) error {
	if strings.TrimSpace(cfg.MongoURI) == "" {
		return fmt.Errorf("mongo_uri is required (or set LOGI_MONGO_URI)")
	}
	if len(strings.TrimSpace(cfg.JWTSecret)) < 32 {
		return fmt.Errorf("jwt_secret must be at least 32 characters for production safety")
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
