// internal/webhook/registry_test.go
package webhook

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/watzon/tg-mock/gen"
)

func TestRegistry_SetGet(t *testing.T) {
	r := NewRegistry(nil)

	cfg := &Config{
		URL:         "https://example.com/webhook",
		SecretToken: "secret123",
	}

	r.Set("123:abc", cfg)

	got := r.Get("123:abc")
	if got == nil {
		t.Fatal("expected config, got nil")
	}
	if got.URL != cfg.URL {
		t.Errorf("URL = %q, want %q", got.URL, cfg.URL)
	}
	if got.SecretToken != cfg.SecretToken {
		t.Errorf("SecretToken = %q, want %q", got.SecretToken, cfg.SecretToken)
	}
	if got.CreatedAt == 0 {
		t.Error("CreatedAt should be auto-set")
	}
}

func TestRegistry_GetNonExistent(t *testing.T) {
	r := NewRegistry(nil)

	got := r.Get("nonexistent")
	if got != nil {
		t.Errorf("expected nil, got %+v", got)
	}
}

func TestRegistry_Delete(t *testing.T) {
	r := NewRegistry(nil)

	r.Set("123:abc", &Config{URL: "https://example.com"})

	if !r.Delete("123:abc") {
		t.Error("Delete should return true for existing webhook")
	}

	if r.Get("123:abc") != nil {
		t.Error("webhook should be deleted")
	}

	if r.Delete("123:abc") {
		t.Error("Delete should return false for non-existent webhook")
	}
}

func TestRegistry_IsActive(t *testing.T) {
	r := NewRegistry(nil)

	// No webhook
	if r.IsActive("123:abc") {
		t.Error("IsActive should return false for non-existent webhook")
	}

	// Webhook with URL
	r.Set("123:abc", &Config{URL: "https://example.com"})
	if !r.IsActive("123:abc") {
		t.Error("IsActive should return true for webhook with URL")
	}

	// Webhook with empty URL
	r.Set("456:def", &Config{URL: ""})
	if r.IsActive("456:def") {
		t.Error("IsActive should return false for webhook with empty URL")
	}
}

func TestRegistry_List(t *testing.T) {
	r := NewRegistry(nil)

	r.Set("123:abc", &Config{URL: "https://example1.com"})
	r.Set("456:def", &Config{URL: "https://example2.com"})

	list := r.List()
	if len(list) != 2 {
		t.Errorf("List length = %d, want 2", len(list))
	}
}

func TestRegistry_Clear(t *testing.T) {
	r := NewRegistry(nil)

	r.Set("123:abc", &Config{URL: "https://example1.com"})
	r.Set("456:def", &Config{URL: "https://example2.com"})

	r.Clear()

	if len(r.List()) != 0 {
		t.Error("Clear should remove all webhooks")
	}
}

func TestRegistry_GetInfo(t *testing.T) {
	r := NewRegistry(nil)

	// No webhook
	info := r.GetInfo("123:abc", 5)
	if info["url"] != "" {
		t.Errorf("url = %q, want empty", info["url"])
	}
	if info["pending_update_count"] != int64(5) {
		t.Errorf("pending_update_count = %v, want 5", info["pending_update_count"])
	}

	// With webhook
	r.Set("123:abc", &Config{
		URL:            "https://example.com",
		IPAddress:      "1.2.3.4",
		MaxConnections: 40,
		AllowedUpdates: []string{"message", "callback_query"},
	})

	info = r.GetInfo("123:abc", 10)
	if info["url"] != "https://example.com" {
		t.Errorf("url = %q, want https://example.com", info["url"])
	}
	if info["ip_address"] != "1.2.3.4" {
		t.Errorf("ip_address = %q, want 1.2.3.4", info["ip_address"])
	}
	if info["max_connections"] != int64(40) {
		t.Errorf("max_connections = %v, want 40", info["max_connections"])
	}
	if allowed, ok := info["allowed_updates"].([]string); !ok || len(allowed) != 2 {
		t.Errorf("allowed_updates = %v, want [message callback_query]", info["allowed_updates"])
	}
}

func TestRegistry_Deliver_Success(t *testing.T) {
	// Create a test server to receive webhook
	var receivedBody map[string]interface{}
	var receivedSecret string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedSecret = r.Header.Get("X-Telegram-Bot-Api-Secret-Token")
		json.NewDecoder(r.Body).Decode(&receivedBody)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok": true}`))
	}))
	defer server.Close()

	r := NewRegistry(nil)
	r.Set("123:abc", &Config{
		URL:         server.URL,
		SecretToken: "mysecret",
	})

	update := map[string]interface{}{
		"update_id": float64(123),
		"message": map[string]interface{}{
			"text": "hello",
		},
	}

	result, err := r.Deliver("123:abc", update)
	if err != nil {
		t.Fatalf("Deliver error: %v", err)
	}

	if !result.Success {
		t.Errorf("Success = false, want true")
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
	if result.DurationMs < 0 {
		t.Errorf("DurationMs = %d, want >= 0", result.DurationMs)
	}

	if receivedSecret != "mysecret" {
		t.Errorf("received secret = %q, want mysecret", receivedSecret)
	}
	if receivedBody["update_id"] != float64(123) {
		t.Errorf("received update_id = %v, want 123", receivedBody["update_id"])
	}
}

func TestRegistry_Deliver_Failure(t *testing.T) {
	// Create a test server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal error"}`))
	}))
	defer server.Close()

	r := NewRegistry(nil)
	r.Set("123:abc", &Config{URL: server.URL})

	result, err := r.Deliver("123:abc", map[string]interface{}{"update_id": float64(1)})
	if err != nil {
		t.Fatalf("Deliver error: %v", err)
	}

	if result.Success {
		t.Error("Success = true, want false")
	}
	if result.StatusCode != 500 {
		t.Errorf("StatusCode = %d, want 500", result.StatusCode)
	}

	// Check that last_error was recorded
	cfg := r.Get("123:abc")
	if cfg.LastErrorDate == nil {
		t.Error("LastErrorDate should be set on failure")
	}
	if cfg.LastErrorMessage == "" {
		t.Error("LastErrorMessage should be set on failure")
	}
}

func TestRegistry_Deliver_NoWebhook(t *testing.T) {
	r := NewRegistry(nil)

	result, err := r.Deliver("nonexistent", map[string]interface{}{})
	if err != nil {
		t.Fatalf("Deliver error: %v", err)
	}

	if result.Success {
		t.Error("Success = true, want false")
	}
	if result.Error != "no webhook configured" {
		t.Errorf("Error = %q, want 'no webhook configured'", result.Error)
	}
}

func TestRegistry_Deliver_ClearsErrorOnSuccess(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	r := NewRegistry(nil)

	// Set up a webhook with a previous error
	now := time.Now().Unix()
	r.Set("123:abc", &Config{
		URL:              server.URL,
		LastErrorDate:    &now,
		LastErrorMessage: "previous error",
	})

	// Deliver successfully
	result, _ := r.Deliver("123:abc", map[string]interface{}{})
	if !result.Success {
		t.Fatal("Deliver should succeed")
	}

	// Error should be cleared
	cfg := r.Get("123:abc")
	if cfg.LastErrorDate != nil {
		t.Error("LastErrorDate should be nil after successful delivery")
	}
	if cfg.LastErrorMessage != "" {
		t.Error("LastErrorMessage should be empty after successful delivery")
	}
}

func TestRegistry_ThreadSafety(t *testing.T) {
	r := NewRegistry(nil)
	var wg sync.WaitGroup

	// Concurrent writes
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			token := "token" + string(rune(i%10))
			r.Set(token, &Config{URL: "https://example.com"})
		}(i)
	}

	// Concurrent reads
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			token := "token" + string(rune(i%10))
			r.Get(token)
			r.IsActive(token)
			r.List()
		}(i)
	}

	// Concurrent deletes
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			token := "token" + string(rune(i%10))
			r.Delete(token)
		}(i)
	}

	wg.Wait()
}

// mockExecutor is a mock implementation of MethodExecutor for testing.
type mockExecutor struct {
	methodCalled string
	paramsCalled map[string]interface{}
	response     interface{}
	responseErr  error
}

func (m *mockExecutor) ExecuteMethod(spec gen.MethodSpec, params map[string]interface{}) (interface{}, error) {
	m.methodCalled = spec.Name
	m.paramsCalled = params
	if m.responseErr != nil {
		return nil, m.responseErr
	}
	if m.response != nil {
		return m.response, nil
	}
	// Default mock response
	return map[string]interface{}{
		"message_id": float64(123),
		"chat": map[string]interface{}{
			"id":   float64(456),
			"type": "private",
		},
	}, nil
}

func TestRegistry_WebhookMethodResponse_SendMessage(t *testing.T) {
	// Create a test server that returns a method call
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"method":"sendMessage","chat_id":123,"text":"Hello!"}`))
	}))
	defer server.Close()

	executor := &mockExecutor{}
	r := NewRegistry(executor)
	r.Set("123:abc", &Config{URL: server.URL})

	update := map[string]interface{}{
		"update_id": float64(1),
		"message": map[string]interface{}{
			"text": "test",
		},
	}

	result, err := r.Deliver("123:abc", update)
	if err != nil {
		t.Fatalf("Deliver error: %v", err)
	}

	if !result.Success {
		t.Errorf("Success = false, want true")
	}

	// Check method was executed
	if result.MethodResult == nil {
		t.Fatal("MethodResult should not be nil")
	}

	if result.MethodResult.Method != "sendMessage" {
		t.Errorf("Method = %q, want 'sendMessage'", result.MethodResult.Method)
	}

	if result.MethodResult.Params["chat_id"] != float64(123) {
		t.Errorf("chat_id = %v, want 123", result.MethodResult.Params["chat_id"])
	}

	if result.MethodResult.Params["text"] != "Hello!" {
		t.Errorf("text = %q, want 'Hello!'", result.MethodResult.Params["text"])
	}

	if result.MethodResult.Response == nil {
		t.Error("Response should not be nil")
	}

	// Verify executor was called
	if executor.methodCalled != "sendMessage" {
		t.Errorf("executor method = %q, want 'sendMessage'", executor.methodCalled)
	}
}

func TestRegistry_WebhookMethodResponse_EmptyResponse(t *testing.T) {
	// Create a test server that returns an empty response (just an ack)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	executor := &mockExecutor{}
	r := NewRegistry(executor)
	r.Set("123:abc", &Config{URL: server.URL})

	result, err := r.Deliver("123:abc", map[string]interface{}{"update_id": float64(1)})
	if err != nil {
		t.Fatalf("Deliver error: %v", err)
	}

	if !result.Success {
		t.Errorf("Success = false, want true")
	}

	// No method should be executed
	if result.MethodResult != nil {
		t.Errorf("MethodResult should be nil for empty response, got %+v", result.MethodResult)
	}

	if executor.methodCalled != "" {
		t.Errorf("executor should not be called, got method %q", executor.methodCalled)
	}
}

func TestRegistry_WebhookMethodResponse_InvalidJSON(t *testing.T) {
	// Create a test server that returns invalid JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`not valid json`))
	}))
	defer server.Close()

	executor := &mockExecutor{}
	r := NewRegistry(executor)
	r.Set("123:abc", &Config{URL: server.URL})

	result, err := r.Deliver("123:abc", map[string]interface{}{"update_id": float64(1)})
	if err != nil {
		t.Fatalf("Deliver error: %v", err)
	}

	if !result.Success {
		t.Errorf("Success = false, want true")
	}

	// No method should be executed for invalid JSON
	if result.MethodResult != nil {
		t.Errorf("MethodResult should be nil for invalid JSON, got %+v", result.MethodResult)
	}
}

func TestRegistry_WebhookMethodResponse_NoMethodField(t *testing.T) {
	// Create a test server that returns valid JSON without a method field
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"acknowledged"}`))
	}))
	defer server.Close()

	executor := &mockExecutor{}
	r := NewRegistry(executor)
	r.Set("123:abc", &Config{URL: server.URL})

	result, err := r.Deliver("123:abc", map[string]interface{}{"update_id": float64(1)})
	if err != nil {
		t.Fatalf("Deliver error: %v", err)
	}

	if !result.Success {
		t.Errorf("Success = false, want true")
	}

	// No method should be executed
	if result.MethodResult != nil {
		t.Errorf("MethodResult should be nil when no method field, got %+v", result.MethodResult)
	}
}

func TestRegistry_WebhookMethodResponse_UnknownMethod(t *testing.T) {
	// Create a test server that returns an unknown method
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"method":"unknownMethod","param":"value"}`))
	}))
	defer server.Close()

	executor := &mockExecutor{}
	r := NewRegistry(executor)
	r.Set("123:abc", &Config{URL: server.URL})

	result, err := r.Deliver("123:abc", map[string]interface{}{"update_id": float64(1)})
	if err != nil {
		t.Fatalf("Deliver error: %v", err)
	}

	if !result.Success {
		t.Errorf("Success = false, want true")
	}

	if result.MethodResult == nil {
		t.Fatal("MethodResult should not be nil")
	}

	if result.MethodResult.Method != "unknownMethod" {
		t.Errorf("Method = %q, want 'unknownMethod'", result.MethodResult.Method)
	}

	// Should have an error for unknown method
	if result.MethodResult.Error == "" {
		t.Error("Expected error for unknown method")
	}

	if result.MethodResult.Response != nil {
		t.Error("Response should be nil for unknown method")
	}
}

func TestRegistry_WebhookMethodResponse_Non2xxStatus(t *testing.T) {
	// Create a test server that returns an error status
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"method":"sendMessage","chat_id":123}`))
	}))
	defer server.Close()

	executor := &mockExecutor{}
	r := NewRegistry(executor)
	r.Set("123:abc", &Config{URL: server.URL})

	result, err := r.Deliver("123:abc", map[string]interface{}{"update_id": float64(1)})
	if err != nil {
		t.Fatalf("Deliver error: %v", err)
	}

	if result.Success {
		t.Error("Success should be false for non-2xx status")
	}

	// Method should NOT be executed for failed delivery
	if result.MethodResult != nil {
		t.Errorf("MethodResult should be nil for failed delivery, got %+v", result.MethodResult)
	}
}

func TestRegistry_WebhookMethodResponse_NilExecutor(t *testing.T) {
	// Create a test server that returns a method call
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"method":"sendMessage","chat_id":123,"text":"Hello!"}`))
	}))
	defer server.Close()

	// Create registry without executor
	r := NewRegistry(nil)
	r.Set("123:abc", &Config{URL: server.URL})

	result, err := r.Deliver("123:abc", map[string]interface{}{"update_id": float64(1)})
	if err != nil {
		t.Fatalf("Deliver error: %v", err)
	}

	if !result.Success {
		t.Errorf("Success = false, want true")
	}

	if result.MethodResult == nil {
		t.Fatal("MethodResult should not be nil")
	}

	// Should have an error for nil executor
	if result.MethodResult.Error == "" {
		t.Error("Expected error for nil executor")
	}
}

func TestRegistry_WebhookMethodResponse_AnswerCallbackQuery(t *testing.T) {
	// Test with a method that returns true (no complex response)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"method":"answerCallbackQuery","callback_query_id":"123"}`))
	}))
	defer server.Close()

	executor := &mockExecutor{
		response: true,
	}
	r := NewRegistry(executor)
	r.Set("123:abc", &Config{URL: server.URL})

	result, err := r.Deliver("123:abc", map[string]interface{}{"update_id": float64(1)})
	if err != nil {
		t.Fatalf("Deliver error: %v", err)
	}

	if !result.Success {
		t.Errorf("Success = false, want true")
	}

	if result.MethodResult == nil {
		t.Fatal("MethodResult should not be nil")
	}

	if result.MethodResult.Method != "answerCallbackQuery" {
		t.Errorf("Method = %q, want 'answerCallbackQuery'", result.MethodResult.Method)
	}

	if result.MethodResult.Response != true {
		t.Errorf("Response = %v, want true", result.MethodResult.Response)
	}
}
