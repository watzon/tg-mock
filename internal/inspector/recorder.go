// Package inspector provides request recording for debugging and testing.
package inspector

import (
	"sync"
	"sync/atomic"
	"time"
)

// RequestRecord captures details of a single Bot API request.
type RequestRecord struct {
	ID         int64                  `json:"id"`
	Timestamp  time.Time              `json:"timestamp"`
	Token      string                 `json:"token"`
	Method     string                 `json:"method"`
	Params     map[string]interface{} `json:"params"`
	ScenarioID string                 `json:"scenario_id,omitempty"`
	Response   interface{}            `json:"response"`
	IsError    bool                   `json:"is_error"`
	StatusCode int                    `json:"status_code"`
}

// Recorder stores and retrieves recorded Bot API requests.
type Recorder struct {
	mu        sync.RWMutex
	requests  []RequestRecord
	idCounter int64
}

// NewRecorder creates a new empty request recorder.
func NewRecorder() *Recorder {
	return &Recorder{
		requests: make([]RequestRecord, 0),
	}
}

// Record adds a request to the recorder and returns its ID.
func (r *Recorder) Record(req RequestRecord) int64 {
	r.mu.Lock()
	defer r.mu.Unlock()

	req.ID = atomic.AddInt64(&r.idCounter, 1)
	if req.Timestamp.IsZero() {
		req.Timestamp = time.Now()
	}

	r.requests = append(r.requests, req)
	return req.ID
}

// List returns recorded requests with optional filtering.
// method: filter by method name (empty = all methods)
// token: filter by token (empty = all tokens)
// limit: maximum number of records to return (0 = all)
func (r *Recorder) List(method, token string, limit int) []RequestRecord {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]RequestRecord, 0)

	for _, req := range r.requests {
		// Apply filters
		if method != "" && req.Method != method {
			continue
		}
		if token != "" && req.Token != token {
			continue
		}

		result = append(result, req)

		if limit > 0 && len(result) >= limit {
			break
		}
	}

	return result
}

// Count returns the total number of recorded requests.
func (r *Recorder) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.requests)
}

// Clear removes all recorded requests.
func (r *Recorder) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.requests = make([]RequestRecord, 0)
}
