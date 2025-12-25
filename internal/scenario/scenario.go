// Package scenario provides error simulation capabilities for the mock Telegram server.
// It allows registering scenarios that can match API calls and return predetermined responses.
package scenario

import (
	"fmt"
	"sync"
	"sync/atomic"
)

// Scenario represents a single error simulation scenario.
// It can match specific API method calls and return predetermined error responses.
type Scenario struct {
	ID       string                 `json:"id"`
	Method   string                 `json:"method"`              // Method to match, or "*" for any method
	Match    map[string]interface{} `json:"match,omitempty"`     // Parameters to match
	Times    int                    `json:"times"`               // Number of times to trigger (0 = unlimited)
	Response *ErrorResponse         `json:"response,omitempty"`  // Error response to return

	used int32 // atomic counter for number of times this scenario has been used
}

// ErrorResponse represents a Telegram API error response.
type ErrorResponse struct {
	ErrorCode   int    `json:"error_code"`
	Description string `json:"description"`
	RetryAfter  int    `json:"retry_after,omitempty"` // For rate limit errors
}

// Matches checks if this scenario matches the given method and parameters.
// A scenario matches if:
// - The method matches exactly, or the scenario method is "*" (wildcard)
// - All parameters in the Match map exist in params with equal values
func (s *Scenario) Matches(method string, params map[string]interface{}) bool {
	// Check method
	if s.Method != "*" && s.Method != method {
		return false
	}

	// Check match criteria
	for key, expected := range s.Match {
		actual, ok := params[key]
		if !ok {
			return false
		}
		if actual != expected {
			return false
		}
	}

	return true
}

// Use increments the usage counter and returns true if the scenario is still valid.
// For unlimited scenarios (Times=0), it always returns true.
func (s *Scenario) Use() bool {
	if s.Times == 0 {
		return true // Unlimited
	}

	used := atomic.AddInt32(&s.used, 1)
	return int(used) <= s.Times
}

// Exhausted returns true if the scenario has been used the maximum number of times.
// For unlimited scenarios (Times=0), it always returns false.
func (s *Scenario) Exhausted() bool {
	if s.Times == 0 {
		return false
	}
	return int(atomic.LoadInt32(&s.used)) >= s.Times
}

// Engine manages a collection of scenarios.
// It provides thread-safe operations for adding, finding, listing, and removing scenarios.
type Engine struct {
	mu        sync.RWMutex
	scenarios []*Scenario
	idCounter int64
}

// NewEngine creates a new scenario engine.
func NewEngine() *Engine {
	return &Engine{
		scenarios: make([]*Scenario, 0),
	}
}

// Add adds a new scenario to the engine.
// If the scenario doesn't have an ID, one will be generated.
// Returns the scenario's ID.
func (e *Engine) Add(s *Scenario) string {
	e.mu.Lock()
	defer e.mu.Unlock()

	if s.ID == "" {
		s.ID = e.generateID()
	}

	e.scenarios = append(e.scenarios, s)
	return s.ID
}

// generateID creates a unique scenario ID.
func (e *Engine) generateID() string {
	id := atomic.AddInt64(&e.idCounter, 1)
	return fmt.Sprintf("scenario-%d", id)
}

// Find returns the first non-exhausted scenario that matches the given method and parameters.
// Returns nil if no matching scenario is found.
func (e *Engine) Find(method string, params map[string]interface{}) *Scenario {
	e.mu.RLock()
	defer e.mu.RUnlock()

	for _, s := range e.scenarios {
		if s.Matches(method, params) && !s.Exhausted() {
			return s
		}
	}
	return nil
}

// List returns a copy of all scenarios in the engine.
func (e *Engine) List() []*Scenario {
	e.mu.RLock()
	defer e.mu.RUnlock()

	result := make([]*Scenario, len(e.scenarios))
	copy(result, e.scenarios)
	return result
}

// Remove removes a scenario by ID.
// Returns true if the scenario was found and removed, false otherwise.
func (e *Engine) Remove(id string) bool {
	e.mu.Lock()
	defer e.mu.Unlock()

	for i, s := range e.scenarios {
		if s.ID == id {
			e.scenarios = append(e.scenarios[:i], e.scenarios[i+1:]...)
			return true
		}
	}
	return false
}

// Clear removes all scenarios from the engine.
func (e *Engine) Clear() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.scenarios = make([]*Scenario, 0)
}
