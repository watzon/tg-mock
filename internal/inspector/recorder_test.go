// internal/inspector/recorder_test.go
package inspector

import (
	"sync"
	"testing"
	"time"
)

func TestRecorder_RecordAndList(t *testing.T) {
	r := NewRecorder()

	// Record requests
	r.Record(RequestRecord{
		Token:      "123:abc",
		Method:     "sendMessage",
		Params:     map[string]interface{}{"chat_id": 123, "text": "hello"},
		Response:   map[string]interface{}{"ok": true},
		IsError:    false,
		StatusCode: 200,
	})
	r.Record(RequestRecord{
		Token:      "123:abc",
		Method:     "getMe",
		Params:     nil,
		Response:   map[string]interface{}{"ok": true},
		IsError:    false,
		StatusCode: 200,
	})

	// List all requests
	requests := r.List("", "", 0)
	if len(requests) != 2 {
		t.Errorf("got %d requests, want 2", len(requests))
	}
}

func TestRecorder_FilterByMethod(t *testing.T) {
	r := NewRecorder()

	// Record different methods
	r.Record(RequestRecord{Method: "sendMessage", Token: "123:abc"})
	r.Record(RequestRecord{Method: "getMe", Token: "123:abc"})
	r.Record(RequestRecord{Method: "sendMessage", Token: "123:abc"})

	// Filter by method
	requests := r.List("sendMessage", "", 0)
	if len(requests) != 2 {
		t.Errorf("got %d sendMessage requests, want 2", len(requests))
	}

	requests = r.List("getMe", "", 0)
	if len(requests) != 1 {
		t.Errorf("got %d getMe requests, want 1", len(requests))
	}

	requests = r.List("unknownMethod", "", 0)
	if len(requests) != 0 {
		t.Errorf("got %d unknown requests, want 0", len(requests))
	}
}

func TestRecorder_FilterByToken(t *testing.T) {
	r := NewRecorder()

	// Record with different tokens
	r.Record(RequestRecord{Method: "sendMessage", Token: "123:abc"})
	r.Record(RequestRecord{Method: "sendMessage", Token: "456:def"})
	r.Record(RequestRecord{Method: "sendMessage", Token: "123:abc"})

	// Filter by token
	requests := r.List("", "123:abc", 0)
	if len(requests) != 2 {
		t.Errorf("got %d requests for token 123:abc, want 2", len(requests))
	}

	requests = r.List("", "456:def", 0)
	if len(requests) != 1 {
		t.Errorf("got %d requests for token 456:def, want 1", len(requests))
	}
}

func TestRecorder_CombinedFilters(t *testing.T) {
	r := NewRecorder()

	// Record mixed requests
	r.Record(RequestRecord{Method: "sendMessage", Token: "123:abc"})
	r.Record(RequestRecord{Method: "sendMessage", Token: "456:def"})
	r.Record(RequestRecord{Method: "getMe", Token: "123:abc"})

	// Filter by both method and token
	requests := r.List("sendMessage", "123:abc", 0)
	if len(requests) != 1 {
		t.Errorf("got %d requests for sendMessage+123:abc, want 1", len(requests))
	}
}

func TestRecorder_Limit(t *testing.T) {
	r := NewRecorder()

	// Record 5 requests
	for i := 0; i < 5; i++ {
		r.Record(RequestRecord{Method: "sendMessage", Token: "123:abc"})
	}

	// Get with limit 3
	requests := r.List("", "", 3)
	if len(requests) != 3 {
		t.Errorf("got %d requests with limit 3, want 3", len(requests))
	}

	// Get all with limit 0
	requests = r.List("", "", 0)
	if len(requests) != 5 {
		t.Errorf("got %d requests with limit 0, want 5", len(requests))
	}
}

func TestRecorder_Clear(t *testing.T) {
	r := NewRecorder()

	// Record requests
	r.Record(RequestRecord{Method: "sendMessage", Token: "123:abc"})
	r.Record(RequestRecord{Method: "getMe", Token: "123:abc"})

	// Verify we have requests
	if r.Count() != 2 {
		t.Errorf("count = %d, want 2", r.Count())
	}

	// Clear
	r.Clear()

	// Verify no requests remain
	if r.Count() != 0 {
		t.Errorf("count after clear = %d, want 0", r.Count())
	}

	requests := r.List("", "", 0)
	if len(requests) != 0 {
		t.Errorf("got %d requests after clear, want 0", len(requests))
	}
}

func TestRecorder_Count(t *testing.T) {
	r := NewRecorder()

	if r.Count() != 0 {
		t.Errorf("initial count = %d, want 0", r.Count())
	}

	r.Record(RequestRecord{Method: "sendMessage", Token: "123:abc"})

	if r.Count() != 1 {
		t.Errorf("count after add = %d, want 1", r.Count())
	}

	r.Record(RequestRecord{Method: "getMe", Token: "123:abc"})

	if r.Count() != 2 {
		t.Errorf("count after second add = %d, want 2", r.Count())
	}
}

func TestRecorder_AutoAssignID(t *testing.T) {
	r := NewRecorder()

	// Record without ID
	id1 := r.Record(RequestRecord{Method: "sendMessage", Token: "123:abc"})
	id2 := r.Record(RequestRecord{Method: "getMe", Token: "123:abc"})

	if id1 != 1 {
		t.Errorf("first auto-assigned id = %d, want 1", id1)
	}
	if id2 != 2 {
		t.Errorf("second auto-assigned id = %d, want 2", id2)
	}

	// Verify the requests have the assigned IDs
	requests := r.List("", "", 0)
	if requests[0].ID != 1 {
		t.Errorf("first request id = %d, want 1", requests[0].ID)
	}
	if requests[1].ID != 2 {
		t.Errorf("second request id = %d, want 2", requests[1].ID)
	}
}

func TestRecorder_TimestampAutoSet(t *testing.T) {
	r := NewRecorder()

	before := time.Now()
	r.Record(RequestRecord{Method: "sendMessage", Token: "123:abc"})
	after := time.Now()

	requests := r.List("", "", 0)
	ts := requests[0].Timestamp

	if ts.Before(before) || ts.After(after) {
		t.Errorf("timestamp %v not between %v and %v", ts, before, after)
	}
}

func TestRecorder_PreservesScenarioID(t *testing.T) {
	r := NewRecorder()

	r.Record(RequestRecord{
		Method:     "sendMessage",
		Token:      "123:abc",
		ScenarioID: "scenario-123",
		IsError:    true,
		StatusCode: 429,
	})

	requests := r.List("", "", 0)
	if requests[0].ScenarioID != "scenario-123" {
		t.Errorf("scenario_id = %q, want %q", requests[0].ScenarioID, "scenario-123")
	}
	if !requests[0].IsError {
		t.Error("is_error = false, want true")
	}
	if requests[0].StatusCode != 429 {
		t.Errorf("status_code = %d, want 429", requests[0].StatusCode)
	}
}

func TestRecorder_ThreadSafety(t *testing.T) {
	r := NewRecorder()
	var wg sync.WaitGroup

	// Concurrent records
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r.Record(RequestRecord{
				Method: "sendMessage",
				Token:  "123:abc",
			})
		}()
	}

	// Concurrent reads
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = r.List("", "", 0)
		}()
	}

	// Concurrent count checks
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = r.Count()
		}()
	}

	wg.Wait()

	// All 100 requests should be recorded
	if r.Count() != 100 {
		t.Errorf("count after concurrent records = %d, want 100", r.Count())
	}
}
