// internal/config/config_test.go
package config

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	yaml := `
server:
  port: 9000
  verbose: true

tokens:
  "123:abc":
    status: active
    bot_name: TestBot
  "456:def":
    status: banned
`

	f, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())

	f.WriteString(yaml)
	f.Close()

	cfg, err := Load(f.Name())
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if cfg.Server.Port != 9000 {
		t.Errorf("port = %d, want 9000", cfg.Server.Port)
	}

	if !cfg.Server.Verbose {
		t.Error("verbose should be true")
	}

	if len(cfg.Tokens) != 2 {
		t.Errorf("got %d tokens, want 2", len(cfg.Tokens))
	}

	if cfg.Tokens["123:abc"].Status != "active" {
		t.Error("token status should be active")
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Server.Port != 8081 {
		t.Errorf("default port = %d, want 8081", cfg.Server.Port)
	}

	if cfg.Server.Verbose {
		t.Error("default verbose should be false")
	}

	if cfg.Server.Strict {
		t.Error("default strict should be false")
	}

	if cfg.Tokens == nil {
		t.Error("tokens map should be initialized")
	}

	if cfg.Scenarios == nil {
		t.Error("scenarios slice should be initialized")
	}
}

func TestLoadConfigWithScenarios(t *testing.T) {
	yaml := `
server:
  port: 8081
  strict: true

scenarios:
  - method: sendMessage
    match:
      chat_id: 12345
    times: 2
    response:
      error_code: 429
      description: "Too Many Requests"
      retry_after: 30
`

	f, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())

	f.WriteString(yaml)
	f.Close()

	cfg, err := Load(f.Name())
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if !cfg.Server.Strict {
		t.Error("strict should be true")
	}

	if len(cfg.Scenarios) != 1 {
		t.Fatalf("got %d scenarios, want 1", len(cfg.Scenarios))
	}

	scenario := cfg.Scenarios[0]
	if scenario.Method != "sendMessage" {
		t.Errorf("method = %s, want sendMessage", scenario.Method)
	}

	if scenario.Times != 2 {
		t.Errorf("times = %d, want 2", scenario.Times)
	}

	if scenario.Response.ErrorCode != 429 {
		t.Errorf("error_code = %d, want 429", scenario.Response.ErrorCode)
	}

	if scenario.Response.RetryAfter != 30 {
		t.Errorf("retry_after = %d, want 30", scenario.Response.RetryAfter)
	}
}

func TestLoadConfigWithStorage(t *testing.T) {
	yaml := `
server:
  port: 8081

storage:
  dir: /tmp/tg-mock-files
`

	f, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())

	f.WriteString(yaml)
	f.Close()

	cfg, err := Load(f.Name())
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if cfg.Storage.Dir != "/tmp/tg-mock-files" {
		t.Errorf("storage dir = %s, want /tmp/tg-mock-files", cfg.Storage.Dir)
	}
}

func TestLoadConfigFileNotFound(t *testing.T) {
	_, err := Load("/nonexistent/config.yaml")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestLoadConfigInvalidYAML(t *testing.T) {
	f, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())

	// Invalid YAML
	f.WriteString("invalid: yaml: content: [")
	f.Close()

	_, err = Load(f.Name())
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}
