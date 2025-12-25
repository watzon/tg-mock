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
	validator       *Validator
	responder       *Responder
}

// NewBotHandler creates a new BotHandler
func NewBotHandler(registry *tokens.Registry, registryEnabled bool) *BotHandler {
	return &BotHandler{
		registry:        registry,
		registryEnabled: registryEnabled,
		validator:       NewValidator(),
		responder:       NewResponder(),
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

	// Parse parameters
	params, err := h.parseParams(r)
	if err != nil {
		h.writeError(w, 400, "Bad Request: "+err.Error())
		return
	}

	// Validate request
	if err := h.validator.Validate(spec, params); err != nil {
		h.writeError(w, 400, "Bad Request: "+err.Error())
		return
	}

	// Generate response
	result, err := h.responder.Generate(spec, params)
	if err != nil {
		h.writeError(w, 500, "Internal Server Error")
		return
	}

	h.writeSuccess(w, result)
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

// parseParams extracts parameters from query string, JSON body, and form data
func (h *BotHandler) parseParams(r *http.Request) (map[string]interface{}, error) {
	params := make(map[string]interface{})

	// Parse query params
	for key, values := range r.URL.Query() {
		if len(values) > 0 {
			params[key] = values[0]
		}
	}

	// Parse JSON body if present
	if r.Body != nil && r.ContentLength > 0 {
		contentType := r.Header.Get("Content-Type")
		if contentType == "application/json" || contentType == "" {
			if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
				return nil, err
			}
		} else if contentType == "application/x-www-form-urlencoded" {
			if err := r.ParseForm(); err != nil {
				return nil, err
			}
			for key, values := range r.PostForm {
				if len(values) > 0 {
					params[key] = values[0]
				}
			}
		}
	}

	return params, nil
}
