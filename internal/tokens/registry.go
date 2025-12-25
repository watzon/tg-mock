// internal/tokens/registry.go
package tokens

import (
	"regexp"
	"sync"
)

type Status string

const (
	StatusActive      Status = "active"
	StatusBanned      Status = "banned"
	StatusDeactivated Status = "deactivated"
)

type TokenInfo struct {
	Status  Status
	BotName string
}

type Registry struct {
	mu     sync.RWMutex
	tokens map[string]TokenInfo
}

var tokenPattern = regexp.MustCompile(`^\d+:[A-Za-z0-9_-]+$`)

func ValidateFormat(token string) bool {
	return tokenPattern.MatchString(token)
}

func NewRegistry() *Registry {
	return &Registry{
		tokens: make(map[string]TokenInfo),
	}
}

func (r *Registry) Register(token string, info TokenInfo) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tokens[token] = info
}

func (r *Registry) Get(token string) (TokenInfo, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	info, ok := r.tokens[token]
	return info, ok
}

func (r *Registry) Delete(token string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.tokens, token)
}

func (r *Registry) UpdateStatus(token string, status Status) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if info, ok := r.tokens[token]; ok {
		info.Status = status
		r.tokens[token] = info
		return true
	}
	return false
}
