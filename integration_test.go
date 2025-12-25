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
}
