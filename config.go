package gateway

import (
	"encoding/json"
	"os"
)

const (
	configPath = "config.json"
)

type Config struct {
	Address  int   `json:"address"`
	IsProd   bool  `json:"isProd"`
	SleepMin uint8 `json:"sleepMin"`

	Services *[]Service `json:"services"`
}

// Reads the main config for the Gateway itself.
func GetConfig() (*Config, error) {
	b, err := LoadConfigFile(configPath)

	if err != nil {
		return nil, err
	}

	cfg := new(Config)

	if err := json.Unmarshal(b, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Loads the content file at the given path.
func LoadConfigFile(path string) ([]byte, error) {
	b, err := os.ReadFile(path)

	if err != nil {
		return nil, err
	}

	return b, nil
}
