// internal/server/bot_handler.go
package server

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/watzon/tg-mock/gen"
	"github.com/watzon/tg-mock/internal/tokens"
)

// BotHandler handles Bot API requests with token validation
type BotHandler struct {
	registry        *tokens.Registry
	registryEnabled bool
}

// NewBotHandler creates a new BotHandler
func NewBotHandler(registry *tokens.Registry, registryEnabled bool) *BotHandler {
	return &BotHandler{
		registry:        registry,
		registryEnabled: registryEnabled,
	}
}

// APIResponse represents the standard Telegram Bot API response format
type APIResponse struct {
	OK          bool        `json:"ok"`
	Result      interface{} `json:"result,omitempty"`
	ErrorCode   int         `json:"error_code,omitempty"`
	Description string      `json:"description,omitempty"`
}

// Handle processes Bot API method requests
func (h *BotHandler) Handle(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	method := chi.URLParam(r, "method")

	w.Header().Set("Content-Type", "application/json")

	// Validate token format
	if !tokens.ValidateFormat(token) {
		h.writeError(w, 401, "Unauthorized: invalid token format")
		return
	}

	// Check registry if enabled
	if h.registryEnabled {
		info, ok := h.registry.Get(token)
		if !ok {
			h.writeError(w, 401, "Unauthorized: token not registered")
			return
		}
		switch info.Status {
		case tokens.StatusBanned:
			h.writeError(w, 403, "Forbidden: bot was banned")
			return
		case tokens.StatusDeactivated:
			h.writeError(w, 401, "Unauthorized: bot was deactivated")
			return
		}
	}

	// Check method exists
	spec, ok := gen.Methods[method]
	if !ok {
		h.writeError(w, 404, "Not Found: method not found")
		return
	}

	// For now, return a success stub
	h.writeSuccess(w, map[string]interface{}{
		"method": spec.Name,
	})
}

func (h *BotHandler) writeError(w http.ResponseWriter, code int, desc string) {
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(APIResponse{
		OK:          false,
		ErrorCode:   code,
		Description: desc,
	})
}

func (h *BotHandler) writeSuccess(w http.ResponseWriter, result interface{}) {
	json.NewEncoder(w).Encode(APIResponse{
		OK:     true,
		Result: result,
	})
}
