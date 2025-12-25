// internal/scenario/scenario_test.go
package scenario

import "testing"

func TestScenarioMatch(t *testing.T) {
	s := &Scenario{
		ID:     "test-1",
		Method: "sendMessage",
		Match: map[string]interface{}{
			"chat_id": float64(123),
		},
		Times: 1,
		Response: &ErrorResponse{
			ErrorCode:   400,
			Description: "Bad Request: chat not found",
		},
	}

	// Should match
	params := map[string]interface{}{
		"chat_id": float64(123),
		"text":    "hello",
	}
	if !s.Matches("sendMessage", params) {
		t.Error("expected scenario to match")
	}

	// Wrong method
	if s.Matches("getMe", params) {
		t.Error("expected scenario not to match different method")
	}

	// Wrong chat_id
	params2 := map[string]interface{}{
		"chat_id": float64(456),
	}
	if s.Matches("sendMessage", params2) {
		t.Error("expected scenario not to match different chat_id")
	}
}

func TestScenarioWildcardMethod(t *testing.T) {
	s := &Scenario{
		ID:     "wildcard-1",
		Method: "*",
		Match:  map[string]interface{}{},
		Times:  0, // unlimited
	}

	// Should match any method
	if !s.Matches("sendMessage", nil) {
		t.Error("expected wildcard to match sendMessage")
	}
	if !s.Matches("getMe", nil) {
		t.Error("expected wildcard to match getMe")
	}
}

func TestScenarioUse(t *testing.T) {
	s := &Scenario{
		ID:    "use-test",
		Times: 2,
	}

	// First use should be valid
	if !s.Use() {
		t.Error("expected first use to be valid")
	}
	if s.Exhausted() {
		t.Error("expected scenario not to be exhausted after first use")
	}

	// Second use should be valid
	if !s.Use() {
		t.Error("expected second use to be valid")
	}
	if !s.Exhausted() {
		t.Error("expected scenario to be exhausted after second use")
	}

	// Third use should be invalid
	if s.Use() {
		t.Error("expected third use to be invalid")
	}
}

func TestScenarioUnlimited(t *testing.T) {
	s := &Scenario{
		ID:    "unlimited-test",
		Times: 0, // unlimited
	}

	for i := 0; i < 100; i++ {
		if !s.Use() {
			t.Errorf("expected unlimited use to always be valid, failed at iteration %d", i)
		}
		if s.Exhausted() {
			t.Errorf("expected unlimited scenario to never be exhausted, failed at iteration %d", i)
		}
	}
}

func TestEngineAdd(t *testing.T) {
	e := NewEngine()

	// Add with custom ID
	s1 := &Scenario{ID: "custom-1", Method: "sendMessage"}
	id := e.Add(s1)
	if id != "custom-1" {
		t.Errorf("expected custom ID, got %s", id)
	}

	// Add without ID (should generate one)
	s2 := &Scenario{Method: "getMe"}
	id2 := e.Add(s2)
	if id2 == "" {
		t.Error("expected generated ID")
	}
	if s2.ID != id2 {
		t.Error("expected scenario ID to be updated")
	}
}

func TestEngineFind(t *testing.T) {
	e := NewEngine()

	s := &Scenario{
		ID:     "find-test",
		Method: "sendMessage",
		Match: map[string]interface{}{
			"chat_id": float64(123),
		},
		Times: 1,
	}
	e.Add(s)

	// Should find matching scenario
	found := e.Find("sendMessage", map[string]interface{}{"chat_id": float64(123)})
	if found == nil {
		t.Error("expected to find scenario")
	}
	if found.ID != "find-test" {
		t.Errorf("expected find-test, got %s", found.ID)
	}

	// Should not find for wrong method
	notFound := e.Find("getMe", map[string]interface{}{"chat_id": float64(123)})
	if notFound != nil {
		t.Error("expected not to find scenario for wrong method")
	}

	// Use the scenario to exhaust it
	s.Use()

	// Should not find exhausted scenario
	exhausted := e.Find("sendMessage", map[string]interface{}{"chat_id": float64(123)})
	if exhausted != nil {
		t.Error("expected not to find exhausted scenario")
	}
}

func TestEngineList(t *testing.T) {
	e := NewEngine()

	e.Add(&Scenario{ID: "s1", Method: "sendMessage"})
	e.Add(&Scenario{ID: "s2", Method: "getMe"})

	list := e.List()
	if len(list) != 2 {
		t.Errorf("expected 2 scenarios, got %d", len(list))
	}
}

func TestEngineRemove(t *testing.T) {
	e := NewEngine()

	e.Add(&Scenario{ID: "to-remove", Method: "sendMessage"})
	e.Add(&Scenario{ID: "to-keep", Method: "getMe"})

	// Remove existing
	if !e.Remove("to-remove") {
		t.Error("expected remove to return true for existing scenario")
	}

	// Verify removed
	list := e.List()
	if len(list) != 1 {
		t.Errorf("expected 1 scenario after remove, got %d", len(list))
	}
	if list[0].ID != "to-keep" {
		t.Error("expected remaining scenario to be 'to-keep'")
	}

	// Remove non-existing
	if e.Remove("non-existing") {
		t.Error("expected remove to return false for non-existing scenario")
	}
}

func TestEngineClear(t *testing.T) {
	e := NewEngine()

	e.Add(&Scenario{ID: "s1", Method: "sendMessage"})
	e.Add(&Scenario{ID: "s2", Method: "getMe"})

	e.Clear()

	list := e.List()
	if len(list) != 0 {
		t.Errorf("expected 0 scenarios after clear, got %d", len(list))
	}
}
