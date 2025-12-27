// internal/config/config.go
package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the main configuration structure for tg-mock
type Config struct {
	Server    ServerConfig            `yaml:"server"`
	Storage   StorageConfig           `yaml:"storage"`
	Tokens    map[string]TokenConfig  `yaml:"tokens"`
	Scenarios []ScenarioConfig        `yaml:"scenarios"`
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Port      int   `yaml:"port"`
	Verbose   bool  `yaml:"verbose"`
	Strict    bool  `yaml:"strict"`
	FakerSeed int64 `yaml:"faker_seed"` // Seed for faker (0 = random, >0 = fixed for determinism)
}

// StorageConfig holds file storage configuration
type StorageConfig struct {
	Dir string `yaml:"dir"`
}

// WebhookConfig holds webhook configuration for a bot token
type WebhookConfig struct {
	URL            string   `yaml:"url"`
	SecretToken    string   `yaml:"secret_token,omitempty"`
	IPAddress      string   `yaml:"ip_address,omitempty"`
	MaxConnections int      `yaml:"max_connections,omitempty"`
	AllowedUpdates []string `yaml:"allowed_updates,omitempty"`
}

// TokenConfig holds configuration for a bot token
type TokenConfig struct {
	Status  string         `yaml:"status"`
	BotName string         `yaml:"bot_name"`
	Webhook *WebhookConfig `yaml:"webhook,omitempty"`
}

// ScenarioConfig defines a test scenario for simulating specific responses
type ScenarioConfig struct {
	Method       string                 `yaml:"method"`
	Match        map[string]interface{} `yaml:"match"`
	Times        int                    `yaml:"times"`
	Response     ResponseConfig         `yaml:"response"`                // For error responses
	ResponseData map[string]interface{} `yaml:"response_data,omitempty"` // For success response overrides
}

// ResponseConfig defines the response to return for a scenario
type ResponseConfig struct {
	ErrorCode   int    `yaml:"error_code"`
	Description string `yaml:"description"`
	RetryAfter  int    `yaml:"retry_after"`
}

// DefaultConfig returns a Config with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port:    8081,
			Verbose: false,
			Strict:  false,
		},
		Tokens:    make(map[string]TokenConfig),
		Scenarios: make([]ScenarioConfig, 0),
	}
}

// Load reads a YAML configuration file from the given path and returns a Config
func Load(path string) (*Config, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
