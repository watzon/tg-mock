// integration_test.go
package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/watzon/tg-mock/internal/server"
)

func TestIntegration(t *testing.T) {
	srv := server.New(server.Config{
		Port:    0,
		Verbose: false,
	})

	ts := httptest.NewServer(srv.Router())
	defer ts.Close()

	t.Run("getMe", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/bot123:abc/getMe")
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		if result["ok"] != true {
			t.Error("expected ok=true")
		}

		// Verify we get a User object in result
		resultData, ok := result["result"].(map[string]interface{})
		if !ok {
			t.Error("expected result to be a User object")
		} else {
			// Verify it has is_bot field
			if _, ok := resultData["is_bot"]; !ok {
				t.Error("expected result to have is_bot field")
			}
		}
	})

	t.Run("sendMessage", func(t *testing.T) {
		body := bytes.NewBufferString(`{"chat_id":123,"text":"Hello"}`)
		resp, err := http.Post(ts.URL+"/bot123:abc/sendMessage", "application/json", body)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		if result["ok"] != true {
			t.Errorf("expected ok=true, got %v", result)
		}

		// Verify we get a Message object in result
		resultData, ok := result["result"].(map[string]interface{})
		if !ok {
			t.Error("expected result to be a Message object")
		} else {
			// Verify message has expected fields
			if _, ok := resultData["message_id"]; !ok {
				t.Error("expected result to have message_id field")
			}
			// Verify text is reflected in the response
			if text, ok := resultData["text"].(string); !ok || text != "Hello" {
				t.Errorf("expected text to be 'Hello', got %v", resultData["text"])
			}
		}
	})

	t.Run("scenario error", func(t *testing.T) {
		// Add scenario
		scenario := `{"method":"sendMessage","match":{"chat_id":999},"times":1,"response":{"error_code":400,"description":"Bad Request: chat not found"}}`
		_, err := http.Post(ts.URL+"/__control/scenarios", "application/json", bytes.NewBufferString(scenario))
		if err != nil {
			t.Fatal(err)
		}

		// Trigger scenario
		body := bytes.NewBufferString(`{"chat_id":999,"text":"test"}`)
		resp, err := http.Post(ts.URL+"/bot123:abc/sendMessage", "application/json", body)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 400 {
			t.Errorf("expected 400, got %d", resp.StatusCode)
		}

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		if result["ok"] != false {
			t.Error("expected ok=false")
		}
		if result["description"] != "Bad Request: chat not found" {
			t.Errorf("expected description 'Bad Request: chat not found', got %v", result["description"])
		}
	})

	t.Run("header scenario", func(t *testing.T) {
		req, _ := http.NewRequest("POST", ts.URL+"/bot123:abc/sendMessage", bytes.NewBufferString(`{"chat_id":1,"text":"test"}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-TG-Mock-Scenario", "rate_limit")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 429 {
			t.Errorf("expected 429, got %d", resp.StatusCode)
		}

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		if result["ok"] != false {
			t.Error("expected ok=false")
		}

		// Verify retry_after is present in parameters
		if params, ok := result["parameters"].(map[string]interface{}); ok {
			if retryAfter, ok := params["retry_after"].(float64); !ok || retryAfter != 30 {
				t.Errorf("expected retry_after=30, got %v", params["retry_after"])
			}
		} else {
			t.Error("expected parameters with retry_after")
		}
	})

	t.Run("request inspector - records requests", func(t *testing.T) {
		// Clear existing requests
		req, _ := http.NewRequest("DELETE", ts.URL+"/__control/requests", nil)
		http.DefaultClient.Do(req)

		// Make a Bot API request
		body := bytes.NewBufferString(`{"chat_id":123,"text":"inspector test"}`)
		_, err := http.Post(ts.URL+"/bot123:abc/sendMessage", "application/json", body)
		if err != nil {
			t.Fatal(err)
		}

		// Get recorded requests
		resp, err := http.Get(ts.URL + "/__control/requests")
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		requests, ok := result["requests"].([]interface{})
		if !ok || len(requests) == 0 {
			t.Fatal("expected at least one recorded request")
		}

		lastReq := requests[len(requests)-1].(map[string]interface{})
		if lastReq["method"] != "sendMessage" {
			t.Errorf("expected method=sendMessage, got %v", lastReq["method"])
		}
		if lastReq["token"] != "123:abc" {
			t.Errorf("expected token=123:abc, got %v", lastReq["token"])
		}
		if lastReq["is_error"] != false {
			t.Errorf("expected is_error=false, got %v", lastReq["is_error"])
		}
	})

	t.Run("request inspector - filters by method", func(t *testing.T) {
		// Clear existing requests
		req, _ := http.NewRequest("DELETE", ts.URL+"/__control/requests", nil)
		http.DefaultClient.Do(req)

		// Make different method calls
		http.Get(ts.URL + "/bot123:abc/getMe")
		http.Post(ts.URL+"/bot123:abc/sendMessage", "application/json", bytes.NewBufferString(`{"chat_id":1,"text":"test"}`))
		http.Get(ts.URL + "/bot123:abc/getMe")

		// Filter by method
		resp, err := http.Get(ts.URL + "/__control/requests?method=getMe")
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		requests := result["requests"].([]interface{})
		if len(requests) != 2 {
			t.Errorf("expected 2 getMe requests, got %d", len(requests))
		}
	})

	t.Run("request inspector - records auth failures", func(t *testing.T) {
		// Clear existing requests
		req, _ := http.NewRequest("DELETE", ts.URL+"/__control/requests", nil)
		http.DefaultClient.Do(req)

		// Make a request with invalid token format
		http.Get(ts.URL + "/botinvalid/getMe")

		// Get recorded requests
		resp, err := http.Get(ts.URL + "/__control/requests")
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		requests := result["requests"].([]interface{})
		if len(requests) != 1 {
			t.Fatalf("expected 1 request, got %d", len(requests))
		}

		req1 := requests[0].(map[string]interface{})
		if req1["is_error"] != true {
			t.Error("expected is_error=true for auth failure")
		}
		if req1["status_code"].(float64) != 401 {
			t.Errorf("expected status_code=401, got %v", req1["status_code"])
		}
	})

	t.Run("request inspector - state includes request count", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/__control/state")
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		if _, ok := result["requests_recorded"]; !ok {
			t.Error("expected requests_recorded in state")
		}
	})

	t.Run("request inspector - reset clears requests", func(t *testing.T) {
		// Make a request first
		http.Get(ts.URL + "/bot123:abc/getMe")

		// Reset
		http.Post(ts.URL+"/__control/reset", "", nil)

		// Check requests are cleared
		resp, err := http.Get(ts.URL + "/__control/requests")
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		if result["count"].(float64) != 0 {
			t.Errorf("expected count=0 after reset, got %v", result["count"])
		}
	})

	t.Run("request inspector - records scenario ID", func(t *testing.T) {
		// Clear existing requests
		req, _ := http.NewRequest("DELETE", ts.URL+"/__control/requests", nil)
		http.DefaultClient.Do(req)

		// Make a request with header scenario
		req, _ = http.NewRequest("POST", ts.URL+"/bot123:abc/sendMessage", bytes.NewBufferString(`{"chat_id":1,"text":"test"}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-TG-Mock-Scenario", "rate_limit")
		http.DefaultClient.Do(req)

		// Get recorded requests
		resp, err := http.Get(ts.URL + "/__control/requests")
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		requests := result["requests"].([]interface{})
		if len(requests) != 1 {
			t.Fatalf("expected 1 request, got %d", len(requests))
		}

		req1 := requests[0].(map[string]interface{})
		if req1["scenario_id"] != "header:rate_limit" {
			t.Errorf("expected scenario_id=header:rate_limit, got %v", req1["scenario_id"])
		}
	})
}
