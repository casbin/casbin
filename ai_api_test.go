// Copyright 2026 The casbin Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package casbin

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// TestExplainWithoutConfig tests that Explain returns error when config is not set.
func TestExplainWithoutConfig(t *testing.T) {
	e, err := NewEnforcer("examples/basic_model.conf", "examples/basic_policy.csv")
	if err != nil {
		t.Fatal(err)
	}

	_, err = e.Explain("alice", "data1", "read")
	if err == nil {
		t.Error("Expected error when AI config is not set")
	}
	if !strings.Contains(err.Error(), "AI config not set") {
		t.Errorf("Expected 'AI config not set' error, got: %v", err)
	}
}

// TestExplainWithMockAPI tests Explain with a mock OpenAI-compatible API.
func TestExplainWithMockAPI(t *testing.T) {
	// Create a mock server that simulates OpenAI API
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type: application/json, got %s", r.Header.Get("Content-Type"))
		}
		if !strings.HasPrefix(r.Header.Get("Authorization"), "Bearer ") {
			t.Errorf("Expected Bearer token in Authorization header, got %s", r.Header.Get("Authorization"))
		}

		// Parse request to verify structure
		var req aiChatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("Failed to decode request: %v", err)
		}

		if req.Model != "gpt-3.5-turbo" {
			t.Errorf("Expected model gpt-3.5-turbo, got %s", req.Model)
		}

		if len(req.Messages) != 2 {
			t.Errorf("Expected 2 messages, got %d", len(req.Messages))
		}

		// Send mock response
		resp := aiChatResponse{
			Choices: []struct {
				Message aiMessage `json:"message"`
			}{
				{
					Message: aiMessage{
						Role:    "assistant",
						Content: "The request was allowed because alice has read permission on data1 according to the policy rule.",
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}))
	defer mockServer.Close()

	// Create enforcer
	e, err := NewEnforcer("examples/basic_model.conf", "examples/basic_policy.csv")
	if err != nil {
		t.Fatal(err)
	}

	// Set AI config with mock server
	e.SetAIConfig(AIConfig{
		Endpoint: mockServer.URL,
		APIKey:   "test-api-key",
		Model:    "gpt-3.5-turbo",
		Timeout:  5 * time.Second,
	})

	// Test explanation for allowed request
	explanation, err := e.Explain("alice", "data1", "read")
	if err != nil {
		t.Fatalf("Failed to get explanation: %v", err)
	}

	if explanation == "" {
		t.Error("Expected non-empty explanation")
	}

	if !strings.Contains(explanation, "allowed") {
		t.Errorf("Expected explanation to mention 'allowed', got: %s", explanation)
	}
}

// TestExplainDenied tests Explain for a denied request.
func TestExplainDenied(t *testing.T) {
	// Create a mock server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := aiChatResponse{
			Choices: []struct {
				Message aiMessage `json:"message"`
			}{
				{
					Message: aiMessage{
						Role:    "assistant",
						Content: "The request was denied because there is no policy rule that allows alice to write to data1.",
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}))
	defer mockServer.Close()

	// Create enforcer
	e, err := NewEnforcer("examples/basic_model.conf", "examples/basic_policy.csv")
	if err != nil {
		t.Fatal(err)
	}

	// Set AI config
	e.SetAIConfig(AIConfig{
		Endpoint: mockServer.URL,
		APIKey:   "test-api-key",
		Model:    "gpt-3.5-turbo",
		Timeout:  5 * time.Second,
	})

	// Test explanation for denied request
	explanation, err := e.Explain("alice", "data1", "write")
	if err != nil {
		t.Fatalf("Failed to get explanation: %v", err)
	}

	if explanation == "" {
		t.Error("Expected non-empty explanation")
	}

	if !strings.Contains(explanation, "denied") {
		t.Errorf("Expected explanation to mention 'denied', got: %s", explanation)
	}
}

// TestExplainAPIError tests handling of API errors.
func TestExplainAPIError(t *testing.T) {
	// Create a mock server that returns an error
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := aiChatResponse{
			Error: &struct {
				Message string `json:"message"`
			}{
				Message: "Invalid API key",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(resp)
	}))
	defer mockServer.Close()

	// Create enforcer
	e, err := NewEnforcer("examples/basic_model.conf", "examples/basic_policy.csv")
	if err != nil {
		t.Fatal(err)
	}

	// Set AI config
	e.SetAIConfig(AIConfig{
		Endpoint: mockServer.URL,
		APIKey:   "invalid-key",
		Model:    "gpt-3.5-turbo",
		Timeout:  5 * time.Second,
	})

	// Test that API error is properly handled
	_, err = e.Explain("alice", "data1", "read")
	if err == nil {
		t.Error("Expected error for API failure")
	}
	if !strings.Contains(err.Error(), "Invalid API key") {
		t.Errorf("Expected API error message, got: %v", err)
	}
}

// TestBuildExplainContext tests the context building function.
func TestBuildExplainContext(t *testing.T) {
	e, err := NewEnforcer("examples/basic_model.conf", "examples/basic_policy.csv")
	if err != nil {
		t.Fatal(err)
	}

	// Test with matched rules
	rvals := []interface{}{"alice", "data1", "read"}
	result := true
	matchedRules := []string{"alice, data1, read"}

	context := e.buildExplainContext(rvals, result, matchedRules)

	// Verify context contains expected elements
	if !strings.Contains(context, "alice") {
		t.Error("Context should contain subject 'alice'")
	}
	if !strings.Contains(context, "data1") {
		t.Error("Context should contain object 'data1'")
	}
	if !strings.Contains(context, "read") {
		t.Error("Context should contain action 'read'")
	}
	if !strings.Contains(context, "true") {
		t.Error("Context should contain result 'true'")
	}
	if !strings.Contains(context, "alice, data1, read") {
		t.Error("Context should contain matched rule")
	}

	// Test with no matched rules
	context2 := e.buildExplainContext(rvals, false, []string{})
	if !strings.Contains(context2, "No policy rules matched") {
		t.Error("Context should indicate no matched rules")
	}
}
