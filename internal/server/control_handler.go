// internal/server/control_handler.go
package server

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/watzon/tg-mock/internal/inspector"
	"github.com/watzon/tg-mock/internal/scenario"
	"github.com/watzon/tg-mock/internal/tokens"
	"github.com/watzon/tg-mock/internal/updates"
)

type ControlHandler struct {
	scenarios *scenario.Engine
	tokens    *tokens.Registry
	updates   *updates.Queue
	requests  *inspector.Recorder
}

func NewControlHandler(scenarios *scenario.Engine, tokens *tokens.Registry, updates *updates.Queue, requests *inspector.Recorder) *ControlHandler {
	return &ControlHandler{
		scenarios: scenarios,
		tokens:    tokens,
		updates:   updates,
		requests:  requests,
	}
}

func (h *ControlHandler) Routes() chi.Router {
	r := chi.NewRouter()

	// Scenarios
	r.Route("/scenarios", func(r chi.Router) {
		r.Get("/", h.listScenarios)
		r.Post("/", h.addScenario)
		r.Delete("/", h.clearScenarios)
		r.Delete("/{id}", h.removeScenario)
	})

	// Tokens
	r.Route("/tokens", func(r chi.Router) {
		r.Post("/", h.registerToken)
		r.Delete("/{token}", h.deleteToken)
		r.Patch("/{token}", h.updateToken)
	})

	// Updates
	r.Route("/updates", func(r chi.Router) {
		r.Get("/", h.listUpdates)
		r.Post("/", h.addUpdate)
		r.Delete("/", h.clearUpdates)
	})

	// Requests
	r.Route("/requests", func(r chi.Router) {
		r.Get("/", h.listRequests)
		r.Delete("/", h.clearRequests)
	})

	// State
	r.Post("/reset", h.reset)
	r.Get("/state", h.getState)

	return r
}

// Scenarios handlers

func (h *ControlHandler) listScenarios(w http.ResponseWriter, r *http.Request) {
	scenarios := h.scenarios.List()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"scenarios": scenarios,
	})
}

func (h *ControlHandler) addScenario(w http.ResponseWriter, r *http.Request) {
	var s scenario.Scenario
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id := h.scenarios.Add(&s)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id": id,
	})
}

func (h *ControlHandler) clearScenarios(w http.ResponseWriter, r *http.Request) {
	h.scenarios.Clear()
	w.WriteHeader(http.StatusNoContent)
}

func (h *ControlHandler) removeScenario(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if h.scenarios.Remove(id) {
		w.WriteHeader(http.StatusNoContent)
	} else {
		http.Error(w, "scenario not found", http.StatusNotFound)
	}
}

// Token handlers

func (h *ControlHandler) registerToken(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token   string        `json:"token"`
		Status  tokens.Status `json:"status"`
		BotName string        `json:"bot_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Status == "" {
		req.Status = tokens.StatusActive
	}

	h.tokens.Register(req.Token, tokens.TokenInfo{
		Status:  req.Status,
		BotName: req.BotName,
	})

	w.WriteHeader(http.StatusCreated)
}

func (h *ControlHandler) deleteToken(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	h.tokens.Delete(token)
	w.WriteHeader(http.StatusNoContent)
}

func (h *ControlHandler) updateToken(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")

	var req struct {
		Status tokens.Status `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if h.tokens.UpdateStatus(token, req.Status) {
		w.WriteHeader(http.StatusOK)
	} else {
		http.Error(w, "token not found", http.StatusNotFound)
	}
}

// Updates handlers

func (h *ControlHandler) listUpdates(w http.ResponseWriter, r *http.Request) {
	updates := h.updates.Get(0, 100)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"updates": updates,
		"pending": h.updates.Pending(),
	})
}

func (h *ControlHandler) addUpdate(w http.ResponseWriter, r *http.Request) {
	var update map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id := h.updates.Add(update)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"update_id": id,
	})
}

func (h *ControlHandler) clearUpdates(w http.ResponseWriter, r *http.Request) {
	h.updates.Clear()
	w.WriteHeader(http.StatusNoContent)
}

// Requests handlers

func (h *ControlHandler) listRequests(w http.ResponseWriter, r *http.Request) {
	method := r.URL.Query().Get("method")
	token := r.URL.Query().Get("token")
	limit := 100
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	requests := h.requests.List(method, token, limit)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"requests": requests,
		"count":    h.requests.Count(),
	})
}

func (h *ControlHandler) clearRequests(w http.ResponseWriter, r *http.Request) {
	h.requests.Clear()
	w.WriteHeader(http.StatusNoContent)
}

// State handlers

func (h *ControlHandler) reset(w http.ResponseWriter, r *http.Request) {
	h.scenarios.Clear()
	h.updates.Clear()
	h.requests.Clear()
	w.WriteHeader(http.StatusNoContent)
}

func (h *ControlHandler) getState(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"scenarios_count":   len(h.scenarios.List()),
		"updates_pending":   h.updates.Pending(),
		"requests_recorded": h.requests.Count(),
	})
}
