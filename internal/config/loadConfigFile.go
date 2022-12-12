package config

import (
	"os"
)

// Loads the content file at the given path.
func LoadConfigFile(path string) ([]byte, error) {
	b, err := os.ReadFile(path)

	if err != nil {
		return nil, err
	}

	return b, nil
}
