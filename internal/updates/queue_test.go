// internal/updates/queue_test.go
package updates

import (
	"sync"
	"testing"
)

func TestQueue_AddAndGet(t *testing.T) {
	q := NewQueue()

	// Add updates
	q.Add(map[string]interface{}{
		"update_id": int64(1),
		"message":   map[string]interface{}{"text": "hello"},
	})
	q.Add(map[string]interface{}{
		"update_id": int64(2),
		"message":   map[string]interface{}{"text": "world"},
	})

	// Get updates
	updates := q.Get(0, 100)
	if len(updates) != 2 {
		t.Errorf("got %d updates, want 2", len(updates))
	}
}

func TestQueue_GetWithOffset(t *testing.T) {
	q := NewQueue()

	// Add updates
	q.Add(map[string]interface{}{
		"update_id": int64(1),
		"message":   map[string]interface{}{"text": "hello"},
	})
	q.Add(map[string]interface{}{
		"update_id": int64(2),
		"message":   map[string]interface{}{"text": "world"},
	})
	q.Add(map[string]interface{}{
		"update_id": int64(3),
		"message":   map[string]interface{}{"text": "foo"},
	})

	// Get with offset should return only updates >= offset
	updates := q.Get(2, 100)
	if len(updates) != 2 {
		t.Errorf("got %d updates after offset, want 2", len(updates))
	}

	// First update should have update_id 2
	if updates[0]["update_id"].(int64) != 2 {
		t.Errorf("first update id = %d, want 2", updates[0]["update_id"])
	}
}

func TestQueue_Acknowledge(t *testing.T) {
	q := NewQueue()

	// Add updates
	q.Add(map[string]interface{}{
		"update_id": int64(1),
		"message":   map[string]interface{}{"text": "hello"},
	})
	q.Add(map[string]interface{}{
		"update_id": int64(2),
		"message":   map[string]interface{}{"text": "world"},
	})
	q.Add(map[string]interface{}{
		"update_id": int64(3),
		"message":   map[string]interface{}{"text": "foo"},
	})

	// Acknowledge offset 2 should remove updates with update_id < 2
	q.Acknowledge(2)
	updates := q.Get(0, 100)
	if len(updates) != 2 {
		t.Errorf("got %d updates after ack, want 2", len(updates))
	}

	// First remaining update should have update_id 2
	if updates[0]["update_id"].(int64) != 2 {
		t.Errorf("first update id after ack = %d, want 2", updates[0]["update_id"])
	}
}

func TestQueue_AutoAssignUpdateID(t *testing.T) {
	q := NewQueue()

	// Add update without update_id
	id1 := q.Add(map[string]interface{}{
		"message": map[string]interface{}{"text": "hello"},
	})
	id2 := q.Add(map[string]interface{}{
		"message": map[string]interface{}{"text": "world"},
	})

	if id1 != 1 {
		t.Errorf("first auto-assigned id = %d, want 1", id1)
	}
	if id2 != 2 {
		t.Errorf("second auto-assigned id = %d, want 2", id2)
	}

	// Verify the updates have the assigned IDs
	updates := q.Get(0, 100)
	if updates[0]["update_id"].(int64) != 1 {
		t.Errorf("first update id = %d, want 1", updates[0]["update_id"])
	}
	if updates[1]["update_id"].(int64) != 2 {
		t.Errorf("second update id = %d, want 2", updates[1]["update_id"])
	}
}

func TestQueue_Clear(t *testing.T) {
	q := NewQueue()

	// Add updates
	q.Add(map[string]interface{}{
		"update_id": int64(1),
		"message":   map[string]interface{}{"text": "hello"},
	})
	q.Add(map[string]interface{}{
		"update_id": int64(2),
		"message":   map[string]interface{}{"text": "world"},
	})

	// Verify we have updates
	if q.Pending() != 2 {
		t.Errorf("pending = %d, want 2", q.Pending())
	}

	// Clear
	q.Clear()

	// Verify no updates remain
	if q.Pending() != 0 {
		t.Errorf("pending after clear = %d, want 0", q.Pending())
	}

	updates := q.Get(0, 100)
	if len(updates) != 0 {
		t.Errorf("got %d updates after clear, want 0", len(updates))
	}
}

func TestQueue_Pending(t *testing.T) {
	q := NewQueue()

	if q.Pending() != 0 {
		t.Errorf("initial pending = %d, want 0", q.Pending())
	}

	q.Add(map[string]interface{}{
		"update_id": int64(1),
		"message":   map[string]interface{}{"text": "hello"},
	})

	if q.Pending() != 1 {
		t.Errorf("pending after add = %d, want 1", q.Pending())
	}

	q.Add(map[string]interface{}{
		"update_id": int64(2),
		"message":   map[string]interface{}{"text": "world"},
	})

	if q.Pending() != 2 {
		t.Errorf("pending after second add = %d, want 2", q.Pending())
	}
}

func TestQueue_GetWithLimit(t *testing.T) {
	q := NewQueue()

	// Add 5 updates
	for i := int64(1); i <= 5; i++ {
		q.Add(map[string]interface{}{
			"update_id": i,
			"message":   map[string]interface{}{"text": "msg"},
		})
	}

	// Get with limit 3
	updates := q.Get(0, 3)
	if len(updates) != 3 {
		t.Errorf("got %d updates with limit 3, want 3", len(updates))
	}

	// Should get first 3 updates
	for i, u := range updates {
		expectedID := int64(i + 1)
		if u["update_id"].(int64) != expectedID {
			t.Errorf("update[%d] id = %d, want %d", i, u["update_id"], expectedID)
		}
	}
}

func TestQueue_ThreadSafety(t *testing.T) {
	q := NewQueue()
	var wg sync.WaitGroup

	// Concurrent adds
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			q.Add(map[string]interface{}{
				"message": map[string]interface{}{"text": "concurrent"},
			})
		}()
	}

	// Concurrent reads
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = q.Get(0, 100)
		}()
	}

	// Concurrent pending checks
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = q.Pending()
		}()
	}

	wg.Wait()

	// All 100 updates should be added
	if q.Pending() != 100 {
		t.Errorf("pending after concurrent adds = %d, want 100", q.Pending())
	}
}
