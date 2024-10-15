package utils

import (
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	ServerAddress string `yaml:"server_address"`
	MongoURI      string `yaml:"mongo_uri"`
	JWTSecret     string `yaml:"jwt_secret"`
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
