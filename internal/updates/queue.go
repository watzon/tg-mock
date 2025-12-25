// Package updates provides a thread-safe queue for managing Telegram updates.
package updates

import (
	"sync"
	"sync/atomic"
)

// Queue is a thread-safe queue for storing and managing Telegram updates.
type Queue struct {
	mu        sync.RWMutex
	updates   []map[string]interface{}
	idCounter int64
}

// NewQueue creates a new empty update queue.
func NewQueue() *Queue {
	return &Queue{
		updates: make([]map[string]interface{}, 0),
	}
}

// Add adds an update to the queue. If the update doesn't have an update_id,
// one will be auto-assigned. Returns the update_id of the added update.
func (q *Queue) Add(update map[string]interface{}) int64 {
	q.mu.Lock()
	defer q.mu.Unlock()

	// Assign update_id if not present
	if _, ok := update["update_id"]; !ok {
		update["update_id"] = atomic.AddInt64(&q.idCounter, 1)
	}

	q.updates = append(q.updates, update)
	return update["update_id"].(int64)
}

// Get returns updates with update_id >= offset, up to limit.
// If offset is 0, all updates are considered.
func (q *Queue) Get(offset int64, limit int) []map[string]interface{} {
	q.mu.RLock()
	defer q.mu.RUnlock()

	result := make([]map[string]interface{}, 0)

	for _, u := range q.updates {
		updateID := u["update_id"].(int64)
		if offset == 0 || updateID >= offset {
			result = append(result, u)
			if len(result) >= limit {
				break
			}
		}
	}

	return result
}

// Acknowledge removes updates with update_id < offset.
// This is used to confirm that updates have been processed.
func (q *Queue) Acknowledge(offset int64) {
	q.mu.Lock()
	defer q.mu.Unlock()

	// Remove updates with update_id < offset
	newUpdates := make([]map[string]interface{}, 0)
	for _, u := range q.updates {
		if u["update_id"].(int64) >= offset {
			newUpdates = append(newUpdates, u)
		}
	}
	q.updates = newUpdates
}

// Clear removes all updates from the queue.
func (q *Queue) Clear() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.updates = make([]map[string]interface{}, 0)
}

// Pending returns the count of pending updates in the queue.
func (q *Queue) Pending() int {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return len(q.updates)
}
