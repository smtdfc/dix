package helpers

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const configFileName = "dix.config.json"

type Config struct {
	Output string `json:"output"`
}

func ReadConfig() (*Config, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	configPath := filepath.Join(cwd, configFileName)
	body, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(body, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &cfg, nil
}

func WriteConfig(cfg *Config) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	configPath := filepath.Join(cwd, configFileName)
	body, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to encode config file: %w", err)
	}

	if err := os.WriteFile(configPath, body, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
