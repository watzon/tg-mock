// internal/webhook/registry.go
package webhook

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"sync"
	"time"
)

// Config represents a registered webhook configuration for a bot token.
type Config struct {
	URL              string   `json:"url"`
	SecretToken      string   `json:"secret_token,omitempty"`
	IPAddress        string   `json:"ip_address,omitempty"`
	MaxConnections   int      `json:"max_connections,omitempty"`
	AllowedUpdates   []string `json:"allowed_updates,omitempty"`
	LastErrorDate    *int64   `json:"last_error_date,omitempty"`
	LastErrorMessage string   `json:"last_error_message,omitempty"`
	CreatedAt        int64    `json:"created_at"`
}

// DeliveryResult captures the result of a webhook delivery attempt.
type DeliveryResult struct {
	Success      bool   `json:"success"`
	StatusCode   int    `json:"status_code"`
	ResponseBody string `json:"response_body,omitempty"`
	Error        string `json:"error,omitempty"`
	DurationMs   int64  `json:"duration_ms"`
}

// Registry manages webhook configurations per bot token.
type Registry struct {
	mu       sync.RWMutex
	webhooks map[string]*Config
	client   *http.Client
}

// NewRegistry creates a new webhook registry.
func NewRegistry() *Registry {
	return &Registry{
		webhooks: make(map[string]*Config),
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Set registers or updates a webhook configuration for a token.
func (r *Registry) Set(token string, cfg *Config) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if cfg.CreatedAt == 0 {
		cfg.CreatedAt = time.Now().Unix()
	}
	r.webhooks[token] = cfg
}

// Get retrieves the webhook configuration for a token.
// Returns nil if no webhook is registered.
func (r *Registry) Get(token string) *Config {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.webhooks[token]
}

// Delete removes the webhook configuration for a token.
// Returns true if a webhook was removed.
func (r *Registry) Delete(token string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.webhooks[token]; ok {
		delete(r.webhooks, token)
		return true
	}
	return false
}

// IsActive returns true if a webhook is registered for the token.
func (r *Registry) IsActive(token string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	cfg := r.webhooks[token]
	return cfg != nil && cfg.URL != ""
}

// List returns all registered webhooks as a map of token -> config.
func (r *Registry) List() map[string]*Config {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make(map[string]*Config, len(r.webhooks))
	for token, cfg := range r.webhooks {
		result[token] = cfg
	}
	return result
}

// Clear removes all webhooks.
func (r *Registry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.webhooks = make(map[string]*Config)
}

// GetInfo returns a WebhookInfo map for the given token,
// suitable for getWebhookInfo response.
func (r *Registry) GetInfo(token string, pendingCount int) map[string]interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()

	cfg := r.webhooks[token]
	if cfg == nil {
		// No webhook configured - return empty webhook info
		return map[string]interface{}{
			"url":                    "",
			"has_custom_certificate": false,
			"pending_update_count":   int64(pendingCount),
		}
	}

	info := map[string]interface{}{
		"url":                    cfg.URL,
		"has_custom_certificate": false, // We don't support custom certificates
		"pending_update_count":   int64(pendingCount),
	}

	if cfg.IPAddress != "" {
		info["ip_address"] = cfg.IPAddress
	}
	if cfg.MaxConnections > 0 {
		info["max_connections"] = int64(cfg.MaxConnections)
	}
	if len(cfg.AllowedUpdates) > 0 {
		info["allowed_updates"] = cfg.AllowedUpdates
	}
	if cfg.LastErrorDate != nil {
		info["last_error_date"] = *cfg.LastErrorDate
	}
	if cfg.LastErrorMessage != "" {
		info["last_error_message"] = cfg.LastErrorMessage
	}

	return info
}

// Deliver sends an update to the registered webhook for a token.
// Returns the delivery result and any error.
func (r *Registry) Deliver(token string, update map[string]interface{}) (*DeliveryResult, error) {
	// Copy config fields under lock to avoid race conditions
	r.mu.RLock()
	cfg := r.webhooks[token]
	if cfg == nil || cfg.URL == "" {
		r.mu.RUnlock()
		return &DeliveryResult{
			Success: false,
			Error:   "no webhook configured",
		}, nil
	}
	webhookURL := cfg.URL
	secretToken := cfg.SecretToken
	r.mu.RUnlock()

	// Marshal the update to JSON
	body, err := json.Marshal(update)
	if err != nil {
		return &DeliveryResult{
			Success: false,
			Error:   err.Error(),
		}, err
	}

	// Create the request
	req, err := http.NewRequest(http.MethodPost, webhookURL, bytes.NewReader(body))
	if err != nil {
		return &DeliveryResult{
			Success: false,
			Error:   err.Error(),
		}, err
	}

	req.Header.Set("Content-Type", "application/json")

	// Add secret token header if configured
	if secretToken != "" {
		req.Header.Set("X-Telegram-Bot-Api-Secret-Token", secretToken)
	}

	// Send the request
	start := time.Now()
	resp, err := r.client.Do(req)
	duration := time.Since(start).Milliseconds()

	if err != nil {
		// Update last error
		r.mu.Lock()
		if c := r.webhooks[token]; c != nil {
			now := time.Now().Unix()
			c.LastErrorDate = &now
			c.LastErrorMessage = err.Error()
		}
		r.mu.Unlock()

		return &DeliveryResult{
			Success:    false,
			Error:      err.Error(),
			DurationMs: duration,
		}, nil
	}
	defer resp.Body.Close()

	// Read response body
	respBody, _ := io.ReadAll(resp.Body)

	result := &DeliveryResult{
		Success:      resp.StatusCode >= 200 && resp.StatusCode < 300,
		StatusCode:   resp.StatusCode,
		ResponseBody: string(respBody),
		DurationMs:   duration,
	}

	if !result.Success {
		result.Error = resp.Status

		// Update last error
		r.mu.Lock()
		if c := r.webhooks[token]; c != nil {
			now := time.Now().Unix()
			c.LastErrorDate = &now
			c.LastErrorMessage = resp.Status
		}
		r.mu.Unlock()
	} else {
		// Clear last error on success
		r.mu.Lock()
		if c := r.webhooks[token]; c != nil {
			c.LastErrorDate = nil
			c.LastErrorMessage = ""
		}
		r.mu.Unlock()
	}

	return result, nil
}
