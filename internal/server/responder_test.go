// internal/server/responder_test.go
package server

import (
	"testing"

	"github.com/watzon/tg-mock/gen"
	"github.com/watzon/tg-mock/internal/faker"
)

func TestGenerateResponse(t *testing.T) {
	// Use a fixed seed for deterministic tests
	f := faker.New(faker.Config{Seed: 12345})
	r := NewResponder(f)

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
		f2 := faker.New(faker.Config{Seed: 12345})
		r2 := NewResponder(f2) // Fresh responder for this test
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

	t.Run("getFile returns File with provided file_id", func(t *testing.T) {
		spec := gen.Methods["getFile"]
		params := map[string]interface{}{
			"file_id": "AgACAgIAAxkBAAIBZ2ABC123DEF456",
		}

		result, err := r.Generate(spec, params)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		file, ok := result.(map[string]interface{})
		if !ok {
			t.Fatalf("expected map, got %T", result)
		}

		if file["file_id"] != "AgACAgIAAxkBAAIBZ2ABC123DEF456" {
			t.Errorf("file_id = %v, want 'AgACAgIAAxkBAAIBZ2ABC123DEF456'", file["file_id"])
		}

		// Check that file_unique_id and file_path exist (values are now generated)
		if _, ok := file["file_unique_id"]; !ok {
			t.Error("file_unique_id should exist")
		}
		if _, ok := file["file_path"]; !ok {
			t.Error("file_path should exist")
		}
		if _, ok := file["file_size"]; !ok {
			t.Error("file_size should exist")
		}
	})

	t.Run("getFile generates file_id when not provided", func(t *testing.T) {
		spec := gen.Methods["getFile"]
		params := map[string]interface{}{}

		result, err := r.Generate(spec, params)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		file, ok := result.(map[string]interface{})
		if !ok {
			t.Fatalf("expected map, got %T", result)
		}

		// File should still have all required fields
		if _, ok := file["file_id"]; !ok {
			t.Error("file_id should exist")
		}
		if _, ok := file["file_unique_id"]; !ok {
			t.Error("file_unique_id should exist")
		}
		if _, ok := file["file_path"]; !ok {
			t.Error("file_path should exist")
		}
	})
}

func TestGenerateWithOverrides(t *testing.T) {
	f := faker.New(faker.Config{Seed: 12345})
	r := NewResponder(f)

	t.Run("overrides are applied", func(t *testing.T) {
		spec := gen.Methods["sendMessage"]
		params := map[string]interface{}{
			"chat_id": int64(12345),
			"text":    "Original text",
		}
		overrides := map[string]interface{}{
			"message_id": int64(99999),
			"text":       "Overridden text",
		}

		result, err := r.GenerateWithOverrides(spec, params, overrides)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		msg := result.(map[string]interface{})

		if msg["message_id"] != int64(99999) {
			t.Errorf("message_id = %v, want 99999", msg["message_id"])
		}
		if msg["text"] != "Overridden text" {
			t.Errorf("text = %v, want 'Overridden text'", msg["text"])
		}
	})

	t.Run("nested overrides are applied", func(t *testing.T) {
		spec := gen.Methods["sendMessage"]
		params := map[string]interface{}{
			"chat_id": int64(12345),
			"text":    "Hello",
		}
		overrides := map[string]interface{}{
			"chat": map[string]interface{}{
				"title": "Custom Chat Title",
			},
		}

		result, err := r.GenerateWithOverrides(spec, params, overrides)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		msg := result.(map[string]interface{})
		chat := msg["chat"].(map[string]interface{})

		if chat["title"] != "Custom Chat Title" {
			t.Errorf("chat.title = %v, want 'Custom Chat Title'", chat["title"])
		}
		// Original id should still be there
		if chat["id"] != int64(12345) {
			t.Errorf("chat.id = %v, want 12345", chat["id"])
		}
	})
}
