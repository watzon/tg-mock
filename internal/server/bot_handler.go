// internal/server/bot_handler.go
package server

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/watzon/tg-mock/gen"
	"github.com/watzon/tg-mock/internal/inspector"
	"github.com/watzon/tg-mock/internal/scenario"
	"github.com/watzon/tg-mock/internal/tokens"
	"github.com/watzon/tg-mock/internal/updates"
	"github.com/watzon/tg-mock/internal/webhook"
)

// BotHandler handles Bot API requests with token validation
type BotHandler struct {
	registry        *tokens.Registry
	registryEnabled bool
	scenarios       *scenario.Engine
	updates         *updates.Queue
	validator       *Validator
	responder       *Responder
	recorder        *inspector.Recorder
	webhooks        *webhook.Registry
}

// NewBotHandler creates a new BotHandler
func NewBotHandler(registry *tokens.Registry, scenarios *scenario.Engine, updates *updates.Queue, responder *Responder, recorder *inspector.Recorder, webhooks *webhook.Registry, registryEnabled bool) *BotHandler {
	return &BotHandler{
		registry:        registry,
		registryEnabled: registryEnabled,
		scenarios:       scenarios,
		updates:         updates,
		validator:       NewValidator(),
		responder:       responder,
		recorder:        recorder,
		webhooks:        webhooks,
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
		h.recordRequest(token, method, nil, "", APIResponse{OK: false, ErrorCode: 401, Description: "Unauthorized: invalid token format"}, true, 401)
		return
	}

	// Check registry if enabled
	if h.registryEnabled {
		info, ok := h.registry.Get(token)
		if !ok {
			h.writeError(w, 401, "Unauthorized: token not registered")
			h.recordRequest(token, method, nil, "", APIResponse{OK: false, ErrorCode: 401, Description: "Unauthorized: token not registered"}, true, 401)
			return
		}
		switch info.Status {
		case tokens.StatusBanned:
			h.writeError(w, 403, "Forbidden: bot was banned")
			h.recordRequest(token, method, nil, "", APIResponse{OK: false, ErrorCode: 403, Description: "Forbidden: bot was banned"}, true, 403)
			return
		case tokens.StatusDeactivated:
			h.writeError(w, 401, "Unauthorized: bot was deactivated")
			h.recordRequest(token, method, nil, "", APIResponse{OK: false, ErrorCode: 401, Description: "Unauthorized: bot was deactivated"}, true, 401)
			return
		}
	}

	// Handle webhook methods before method lookup
	switch method {
	case "setWebhook":
		h.handleSetWebhook(w, token, h.parseParamsOrEmpty(r))
		return
	case "deleteWebhook":
		h.handleDeleteWebhook(w, token, h.parseParamsOrEmpty(r))
		return
	case "getWebhookInfo":
		h.handleGetWebhookInfo(w, token)
		return
	}

	// Check method exists
	spec, ok := gen.Methods[method]
	if !ok {
		h.writeError(w, 404, "Not Found: method not found")
		h.recordRequest(token, method, nil, "", APIResponse{OK: false, ErrorCode: 404, Description: "Not Found: method not found"}, true, 404)
		return
	}

	// Parse parameters
	params, err := h.parseParams(r)
	if err != nil {
		desc := "Bad Request: " + err.Error()
		h.writeError(w, 400, desc)
		h.recordRequest(token, method, nil, "", APIResponse{OK: false, ErrorCode: 400, Description: desc}, true, 400)
		return
	}

	// Check for header-based scenario
	if scenarioName := r.Header.Get("X-TG-Mock-Scenario"); scenarioName != "" {
		if resp := h.handleHeaderScenarioWithRecording(w, r, token, method, params, scenarioName); resp {
			return
		}
	}

	// Check for queued scenarios
	var scenarioOverrides map[string]interface{}
	var matchedScenarioID string
	if s := h.scenarios.Find(method, params); s != nil {
		s.Use()
		matchedScenarioID = s.ID
		if s.IsError() {
			h.writeErrorResponse(w, s.Response)
			h.recordRequest(token, method, params, matchedScenarioID, map[string]interface{}{
				"ok":          false,
				"error_code":  s.Response.ErrorCode,
				"description": s.Response.Description,
			}, true, s.Response.ErrorCode)
			return
		}
		// Store response data overrides for later use
		if s.HasResponseData() {
			scenarioOverrides = s.ResponseData
		}
	}

	// Handle getUpdates specially
	if method == "getUpdates" {
		// Check for webhook conflict
		if h.webhooks.IsActive(token) {
			desc := "Conflict: can't use getUpdates method while webhook is active"
			h.writeError(w, 409, desc)
			h.recordRequest(token, method, params, matchedScenarioID, APIResponse{OK: false, ErrorCode: 409, Description: desc}, true, 409)
			return
		}
		result := h.handleGetUpdates(params)
		h.writeSuccess(w, result)
		h.recordRequest(token, method, params, matchedScenarioID, APIResponse{OK: true, Result: result}, false, 200)
		return
	}

	// Validate request
	if err := h.validator.Validate(spec, params); err != nil {
		desc := "Bad Request: " + err.Error()
		h.writeError(w, 400, desc)
		h.recordRequest(token, method, params, matchedScenarioID, APIResponse{OK: false, ErrorCode: 400, Description: desc}, true, 400)
		return
	}

	// Generate response (with scenario overrides if present)
	result, err := h.responder.GenerateWithOverrides(spec, params, scenarioOverrides)
	if err != nil {
		h.writeError(w, 500, "Internal Server Error")
		h.recordRequest(token, method, params, matchedScenarioID, APIResponse{OK: false, ErrorCode: 500, Description: "Internal Server Error"}, true, 500)
		return
	}

	h.writeSuccess(w, result)
	h.recordRequest(token, method, params, matchedScenarioID, APIResponse{OK: true, Result: result}, false, 200)
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

// handleHeaderScenarioWithRecording handles X-TG-Mock-Scenario header-based error scenarios
// and records the request. It looks up pre-built errors by name and returns the appropriate error response.
func (h *BotHandler) handleHeaderScenarioWithRecording(w http.ResponseWriter, r *http.Request, token, method string, params map[string]interface{}, name string) bool {
	resp := scenario.GetBuiltinError(name)
	if resp == nil {
		return false
	}

	// Allow retry_after override via header
	if retryAfter := r.Header.Get("X-TG-Mock-Retry-After"); retryAfter != "" {
		if val, err := strconv.Atoi(retryAfter); err == nil {
			resp = &scenario.ErrorResponse{
				ErrorCode:   resp.ErrorCode,
				Description: resp.Description,
				RetryAfter:  val,
			}
		}
	}

	h.writeErrorResponse(w, resp)

	// Record with header: prefix for scenario ID
	response := map[string]interface{}{
		"ok":          false,
		"error_code":  resp.ErrorCode,
		"description": resp.Description,
	}
	if resp.RetryAfter > 0 {
		response["parameters"] = map[string]interface{}{
			"retry_after": resp.RetryAfter,
		}
	}
	h.recordRequest(token, method, params, "header:"+name, response, true, resp.ErrorCode)

	return true
}

// writeErrorResponse writes a scenario error response with proper formatting
func (h *BotHandler) writeErrorResponse(w http.ResponseWriter, resp *scenario.ErrorResponse) {
	w.WriteHeader(resp.ErrorCode)
	response := map[string]interface{}{
		"ok":          false,
		"error_code":  resp.ErrorCode,
		"description": resp.Description,
	}
	if resp.RetryAfter > 0 {
		response["parameters"] = map[string]interface{}{
			"retry_after": resp.RetryAfter,
		}
	}
	json.NewEncoder(w).Encode(response)
}

// handleGetUpdates processes the getUpdates method by returning updates from the queue
func (h *BotHandler) handleGetUpdates(params map[string]interface{}) []map[string]interface{} {
	offset := int64(0)
	if o, ok := params["offset"].(float64); ok {
		offset = int64(o)
	} else if o, ok := params["offset"].(string); ok {
		// Handle string offset (from query params)
		if parsed, err := parseInt64(o); err == nil {
			offset = parsed
		}
	}

	limit := 100
	if l, ok := params["limit"].(float64); ok {
		limit = int(l)
	} else if l, ok := params["limit"].(string); ok {
		// Handle string limit (from query params)
		if parsed, err := parseInt(l); err == nil {
			limit = parsed
		}
	}

	// Acknowledge previous updates
	if offset > 0 {
		h.updates.Acknowledge(offset)
	}

	return h.updates.Get(offset, limit)
}

// parseInt64 parses a string to int64
func parseInt64(s string) (int64, error) {
	var n int64
	err := json.Unmarshal([]byte(s), &n)
	return n, err
}

// parseInt parses a string to int
func parseInt(s string) (int, error) {
	var n int
	err := json.Unmarshal([]byte(s), &n)
	return n, err
}

// recordRequest records a request to the inspector
func (h *BotHandler) recordRequest(token, method string, params map[string]interface{}, scenarioID string, response interface{}, isError bool, statusCode int) {
	h.recorder.Record(inspector.RequestRecord{
		Timestamp:  time.Now(),
		Token:      token,
		Method:     method,
		Params:     params,
		ScenarioID: scenarioID,
		Response:   response,
		IsError:    isError,
		StatusCode: statusCode,
	})
}

// parseParamsOrEmpty parses request parameters, returning empty map on error
func (h *BotHandler) parseParamsOrEmpty(r *http.Request) map[string]interface{} {
	params, err := h.parseParams(r)
	if err != nil {
		return make(map[string]interface{})
	}
	return params
}

// handleSetWebhook handles the setWebhook Bot API method
func (h *BotHandler) handleSetWebhook(w http.ResponseWriter, token string, params map[string]interface{}) {
	url, _ := params["url"].(string)

	// Empty URL means delete webhook
	if url == "" {
		h.webhooks.Delete(token)
		if dropPending, _ := params["drop_pending_updates"].(bool); dropPending {
			h.updates.Clear()
		}
		h.writeSuccess(w, true)
		h.recordRequest(token, "setWebhook", params, "", APIResponse{OK: true, Result: true}, false, 200)
		return
	}

	// Build webhook config from params
	cfg := &webhook.Config{
		URL: url,
	}

	if secret, ok := params["secret_token"].(string); ok {
		cfg.SecretToken = secret
	}
	if ip, ok := params["ip_address"].(string); ok {
		cfg.IPAddress = ip
	}
	if maxConn, ok := params["max_connections"].(float64); ok {
		cfg.MaxConnections = int(maxConn)
	}
	if allowed, ok := params["allowed_updates"].([]interface{}); ok {
		for _, v := range allowed {
			if s, ok := v.(string); ok {
				cfg.AllowedUpdates = append(cfg.AllowedUpdates, s)
			}
		}
	}

	h.webhooks.Set(token, cfg)

	// Handle drop_pending_updates
	if dropPending, _ := params["drop_pending_updates"].(bool); dropPending {
		h.updates.Clear()
	}

	h.writeSuccess(w, true)
	h.recordRequest(token, "setWebhook", params, "", APIResponse{OK: true, Result: true}, false, 200)
}

// handleDeleteWebhook handles the deleteWebhook Bot API method
func (h *BotHandler) handleDeleteWebhook(w http.ResponseWriter, token string, params map[string]interface{}) {
	h.webhooks.Delete(token)

	// Handle drop_pending_updates
	if dropPending, _ := params["drop_pending_updates"].(bool); dropPending {
		h.updates.Clear()
	}

	h.writeSuccess(w, true)
	h.recordRequest(token, "deleteWebhook", params, "", APIResponse{OK: true, Result: true}, false, 200)
}

// handleGetWebhookInfo handles the getWebhookInfo Bot API method
func (h *BotHandler) handleGetWebhookInfo(w http.ResponseWriter, token string) {
	pendingCount := h.updates.Pending()
	info := h.webhooks.GetInfo(token, pendingCount)
	h.writeSuccess(w, info)
	h.recordRequest(token, "getWebhookInfo", nil, "", APIResponse{OK: true, Result: info}, false, 200)
}
