package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Host      string            `yaml:"host"`
	Port      int               `yaml:"port"`
	OllamaURL string            `yaml:"ollama_url"`
	APIKeys   map[string]string `yaml:"api_keys"`
	Timeout   int               `yaml:"timeout"`
	LogLevel  string            `yaml:"log_level"`
}

// Load reads the configuration from the specified file
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Set defaults
	if cfg.Host == "" {
		cfg.Host = "0.0.0.0"
	}
	if cfg.Port == 0 {
		cfg.Port = 8080
	}
	if cfg.OllamaURL == "" {
		cfg.OllamaURL = "http://localhost:11434"
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 300
	}
	if cfg.LogLevel == "" {
		cfg.LogLevel = "info"
	}

	return &cfg, nil
}

// GetAddress returns the full address for the server to listen on
func (c *Config) GetAddress() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// GetTimeout returns the timeout as a Duration
func (c *Config) GetTimeout() time.Duration {
	return time.Duration(c.Timeout) * time.Second
}

// GetAlias returns the alias for a given API key, or empty string if not found
func (c *Config) GetAlias(key string) string {
	return c.APIKeys[key]
}
