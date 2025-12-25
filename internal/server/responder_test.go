// internal/server/responder_test.go
package server

import (
	"testing"

	"github.com/watzon/tg-mock/gen"
)

func TestGenerateResponse(t *testing.T) {
	r := NewResponder()

	t.Run("getMe returns User", func(t *testing.T) {
		spec := gen.Methods["getMe"]
		params := map[string]interface{}{}

		result, err := r.Generate(spec, params)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Result should be a User
		user, ok := result.(map[string]interface{})
		if !ok {
			t.Fatalf("expected map, got %T", result)
		}

		if _, ok := user["id"]; !ok {
			t.Error("User should have id field")
		}
		if _, ok := user["is_bot"]; !ok {
			t.Error("User should have is_bot field")
		}
	})

	t.Run("sendMessage reflects chat_id", func(t *testing.T) {
		spec := gen.Methods["sendMessage"]
		params := map[string]interface{}{
			"chat_id": int64(12345),
			"text":    "Hello world",
		}

		result, err := r.Generate(spec, params)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		msg, ok := result.(map[string]interface{})
		if !ok {
			t.Fatalf("expected map, got %T", result)
		}

		chat, ok := msg["chat"].(map[string]interface{})
		if !ok {
			t.Fatal("expected chat field")
		}

		if chat["id"] != int64(12345) {
			t.Errorf("chat.id = %v, want 12345", chat["id"])
		}

		if msg["text"] != "Hello world" {
			t.Errorf("text = %v, want 'Hello world'", msg["text"])
		}
	})

	t.Run("Boolean return type returns true", func(t *testing.T) {
		spec := gen.Methods["close"]
		params := map[string]interface{}{}

		result, err := r.Generate(spec, params)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		boolResult, ok := result.(bool)
		if !ok {
			t.Fatalf("expected bool, got %T", result)
		}

		if !boolResult {
			t.Error("expected true, got false")
		}
	})

	t.Run("getUpdates returns empty array", func(t *testing.T) {
		spec := gen.Methods["getUpdates"]
		params := map[string]interface{}{}

		result, err := r.Generate(spec, params)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		arr, ok := result.([]interface{})
		if !ok {
			t.Fatalf("expected []interface{}, got %T", result)
		}

		if len(arr) != 0 {
			t.Errorf("expected empty array, got %d elements", len(arr))
		}
	})

	t.Run("message IDs are unique and incrementing", func(t *testing.T) {
		r2 := NewResponder() // Fresh responder for this test
		spec := gen.Methods["sendMessage"]
		params := map[string]interface{}{
			"chat_id": int64(1),
			"text":    "test",
		}

		result1, _ := r2.Generate(spec, params)
		result2, _ := r2.Generate(spec, params)

		msg1 := result1.(map[string]interface{})
		msg2 := result2.(map[string]interface{})

		id1 := msg1["message_id"].(int64)
		id2 := msg2["message_id"].(int64)

		if id1 >= id2 {
			t.Errorf("message IDs should be incrementing: got %d, %d", id1, id2)
		}
	})
}
