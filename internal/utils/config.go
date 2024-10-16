package utils

import (
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	ServerAddress          string `yaml:"server_address"`
	MongoURI               string `yaml:"mongo_uri"`
	JWTSecret              string `yaml:"jwt_secret"`
	MessagingType          string `yaml:"messaging_type"`
	NATSURL                string `yaml:"nats_url"`
	DistanceCalculatorType string `yaml:"distance_calculator_type"`
	GoogleMapsAPIKey	   string `yaml:"google_maps_api_key"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
