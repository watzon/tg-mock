// internal/tokens/registry_test.go
package tokens

import "testing"

func TestValidateFormat(t *testing.T) {
	tests := []struct {
		token string
		valid bool
	}{
		{"123456789:ABC-DEF1234ghIkl-zyx57W2v1u123ew11", true},
		{"123456:abcdef", true},
		{"invalid", false},
		{"", false},
		{"123456:", false},
		{":ABC", false},
	}

	for _, tt := range tests {
		t.Run(tt.token, func(t *testing.T) {
			if got := ValidateFormat(tt.token); got != tt.valid {
				t.Errorf("ValidateFormat(%q) = %v, want %v", tt.token, got, tt.valid)
			}
		})
	}
}

func TestRegistry(t *testing.T) {
	r := NewRegistry()

	// Register a token
	r.Register("123:abc", TokenInfo{Status: StatusActive, BotName: "TestBot"})

	// Check status
	info, ok := r.Get("123:abc")
	if !ok {
		t.Fatal("expected token to be registered")
	}
	if info.Status != StatusActive {
		t.Errorf("got status %v, want %v", info.Status, StatusActive)
	}

	// Unknown token
	_, ok = r.Get("unknown:token")
	if ok {
		t.Error("expected unknown token to not be found")
	}
}

func TestRegistryDelete(t *testing.T) {
	r := NewRegistry()

	// Register a token
	r.Register("123:abc", TokenInfo{Status: StatusActive, BotName: "TestBot"})

	// Verify it exists
	_, ok := r.Get("123:abc")
	if !ok {
		t.Fatal("expected token to be registered")
	}

	// Delete it
	r.Delete("123:abc")

	// Verify it's gone
	_, ok = r.Get("123:abc")
	if ok {
		t.Error("expected token to be deleted")
	}
}

func TestRegistryUpdateStatus(t *testing.T) {
	r := NewRegistry()

	// Register a token
	r.Register("123:abc", TokenInfo{Status: StatusActive, BotName: "TestBot"})

	// Update status
	updated := r.UpdateStatus("123:abc", StatusBanned)
	if !updated {
		t.Fatal("expected UpdateStatus to return true")
	}

	// Verify status was updated
	info, ok := r.Get("123:abc")
	if !ok {
		t.Fatal("expected token to still exist")
	}
	if info.Status != StatusBanned {
		t.Errorf("got status %v, want %v", info.Status, StatusBanned)
	}
	// Verify BotName was preserved
	if info.BotName != "TestBot" {
		t.Errorf("got BotName %v, want TestBot", info.BotName)
	}

	// Try to update unknown token
	updated = r.UpdateStatus("unknown:token", StatusDeactivated)
	if updated {
		t.Error("expected UpdateStatus to return false for unknown token")
	}
}
