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

// TestAIPolicyManagementAPI tests the management APIs for AI policies.
func TestAIPolicyManagementAPI(t *testing.T) {
	e, err := NewEnforcer("examples/ai_policy_model.conf", "examples/ai_policy.csv")
	if err != nil {
		t.Fatal(err)
	}

	// Test GetAIPolicy
	policies, err := e.GetAIPolicy()
	if err != nil {
		t.Fatalf("GetAIPolicy failed: %v", err)
	}
	if len(policies) != 2 {
		t.Errorf("Expected 2 AI policies, got %d", len(policies))
	}

	// Test HasAIPolicy
	has, err := e.HasAIPolicy("allow US residential IPs to read data1")
	if err != nil {
		t.Fatalf("HasAIPolicy failed: %v", err)
	}
	if !has {
		t.Error("Expected AI policy to exist")
	}

	// Test AddAIPolicy
	added, err := e.AddAIPolicy("deny requests with suspicious patterns")
	if err != nil {
		t.Fatalf("AddAIPolicy failed: %v", err)
	}
	if !added {
		t.Error("Expected AI policy to be added")
	}

	// Verify the policy was added
	policies, err = e.GetAIPolicy()
	if err != nil {
		t.Fatalf("GetAIPolicy failed: %v", err)
	}
	if len(policies) != 3 {
		t.Errorf("Expected 3 AI policies after adding, got %d", len(policies))
	}

	// Test AddAIPolicy with duplicate (should not add)
	added, err = e.AddAIPolicy("deny requests with suspicious patterns")
	if err != nil {
		t.Fatalf("AddAIPolicy failed: %v", err)
	}
	if added {
		t.Error("Expected duplicate AI policy not to be added")
	}

	// Test RemoveAIPolicy
	removed, err := e.RemoveAIPolicy("deny requests with suspicious patterns")
	if err != nil {
		t.Fatalf("RemoveAIPolicy failed: %v", err)
	}
	if !removed {
		t.Error("Expected AI policy to be removed")
	}

	// Verify the policy was removed
	policies, err = e.GetAIPolicy()
	if err != nil {
		t.Fatalf("GetAIPolicy failed: %v", err)
	}
	if len(policies) != 2 {
		t.Errorf("Expected 2 AI policies after removing, got %d", len(policies))
	}
}

// TestAIPolicyBulkOperations tests bulk operations for AI policies.
func TestAIPolicyBulkOperations(t *testing.T) {
	e, err := NewEnforcer("examples/ai_policy_model.conf")
	if err != nil {
		t.Fatal(err)
	}

	// Test AddAIPolicies
	rules := [][]string{
		{"allow authenticated users to read public data"},
		{"deny anonymous users from writing sensitive data"},
		{"allow admin users all access"},
	}

	added, err := e.AddAIPolicies(rules)
	if err != nil {
		t.Fatalf("AddAIPolicies failed: %v", err)
	}
	if !added {
		t.Error("Expected AI policies to be added")
	}

	// Verify the policies were added
	policies, err := e.GetAIPolicy()
	if err != nil {
		t.Fatalf("GetAIPolicy failed: %v", err)
	}
	if len(policies) != 3 {
		t.Errorf("Expected 3 AI policies, got %d", len(policies))
	}

	// Test RemoveAIPolicies
	removeRules := [][]string{
		{"allow authenticated users to read public data"},
		{"deny anonymous users from writing sensitive data"},
	}

	removed, err := e.RemoveAIPolicies(removeRules)
	if err != nil {
		t.Fatalf("RemoveAIPolicies failed: %v", err)
	}
	if !removed {
		t.Error("Expected AI policies to be removed")
	}

	// Verify the policies were removed
	policies, err = e.GetAIPolicy()
	if err != nil {
		t.Fatalf("GetAIPolicy failed: %v", err)
	}
	if len(policies) != 1 {
		t.Errorf("Expected 1 AI policy after removing, got %d", len(policies))
	}
}

// TestAIPolicyUpdate tests updating AI policies.
func TestAIPolicyUpdate(t *testing.T) {
	e, err := NewEnforcer("examples/ai_policy_model.conf")
	if err != nil {
		t.Fatal(err)
	}

	// Add a policy first
	_, err = e.AddAIPolicy("allow read access to public data")
	if err != nil {
		t.Fatalf("AddAIPolicy failed: %v", err)
	}

	// Update the policy
	updated, err := e.UpdateAIPolicy(
		[]string{"allow read access to public data"},
		[]string{"allow read and write access to public data"},
	)
	if err != nil {
		t.Fatalf("UpdateAIPolicy failed: %v", err)
	}
	if !updated {
		t.Error("Expected AI policy to be updated")
	}

	// Verify the policy was updated
	has, err := e.HasAIPolicy("allow read and write access to public data")
	if err != nil {
		t.Fatalf("HasAIPolicy failed: %v", err)
	}
	if !has {
		t.Error("Expected updated AI policy to exist")
	}

	// Verify old policy doesn't exist
	has, err = e.HasAIPolicy("allow read access to public data")
	if err != nil {
		t.Fatalf("HasAIPolicy failed: %v", err)
	}
	if has {
		t.Error("Expected old AI policy not to exist")
	}
}

// TestAIPolicyEnforcement tests AI policy enforcement with a mock LLM API.
func TestAIPolicyEnforcement(t *testing.T) {
	// Create a mock server that simulates LLM API
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Parse request
		var req aiChatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("Failed to decode request: %v", err)
		}

		// Determine response based on request content
		userMessage := req.Messages[1].Content
		var responseContent string

		if strings.Contains(userMessage, "192.168.2.1") && strings.Contains(userMessage, "data1") && strings.Contains(userMessage, "read") {
			if strings.Contains(userMessage, "allow US residential IPs to read data1") {
				responseContent = "ALLOW"
			} else {
				responseContent = "DENY"
			}
		} else if strings.Contains(userMessage, "credential") || strings.Contains(userMessage, "secret") {
			responseContent = "DENY"
		} else {
			responseContent = "DENY"
		}

		resp := aiChatResponse{
			Choices: []struct {
				Message aiMessage `json:"message"`
			}{
				{
					Message: aiMessage{
						Role:    "assistant",
						Content: responseContent,
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}))
	defer mockServer.Close()

	// Create enforcer with AI policy
	e, err := NewEnforcer("examples/ai_policy_model.conf")
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

	// Add an AI policy
	_, err = e.AddAIPolicy("allow US residential IPs to read data1")
	if err != nil {
		t.Fatalf("AddAIPolicy failed: %v", err)
	}

	// Test enforcement - should be allowed by AI policy
	allowed, err := e.Enforce("192.168.2.1", "data1", "read")
	if err != nil {
		t.Fatalf("Enforce failed: %v", err)
	}
	if !allowed {
		t.Error("Expected request to be allowed by AI policy")
	}
}

// TestAIPolicyWithoutAIConfig tests that enforcement works when AI config is not set.
func TestAIPolicyWithoutAIConfig(t *testing.T) {
	e, err := NewEnforcer("examples/ai_policy_model.conf")
	if err != nil {
		t.Fatal(err)
	}

	// Add an AI policy without setting AI config
	_, err = e.AddAIPolicy("allow all requests")
	if err != nil {
		t.Fatalf("AddAIPolicy failed: %v", err)
	}

	// Test enforcement - should fall back to deny since AI evaluation fails
	allowed, err := e.Enforce("alice", "data1", "read")
	if err != nil {
		t.Fatalf("Enforce failed: %v", err)
	}
	// Without AI config, AI policy evaluation will fail and be denied
	if allowed {
		t.Error("Expected request to be denied when AI config is not set")
	}
}

// TestAIPolicyWithTraditionalPolicies tests AI policies working alongside traditional policies.
func TestAIPolicyWithTraditionalPolicies(t *testing.T) {
	// Create a mock server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := aiChatResponse{
			Choices: []struct {
				Message aiMessage `json:"message"`
			}{
				{
					Message: aiMessage{
						Role:    "assistant",
						Content: "ALLOW",
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}))
	defer mockServer.Close()

	e, err := NewEnforcer("examples/ai_policy_model.conf")
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

	// Add both traditional and AI policies
	_, err = e.AddPolicy("alice", "data1", "read")
	if err != nil {
		t.Fatalf("AddPolicy failed: %v", err)
	}

	_, err = e.AddAIPolicy("allow all authenticated users")
	if err != nil {
		t.Fatalf("AddAIPolicy failed: %v", err)
	}

	// Test enforcement - AI policy is checked first
	allowed, err := e.Enforce("bob", "data2", "write")
	if err != nil {
		t.Fatalf("Enforce failed: %v", err)
	}
	if !allowed {
		t.Error("Expected request to be allowed by AI policy")
	}
}

// TestGetFilteredAIPolicy tests filtering AI policies.
func TestGetFilteredAIPolicy(t *testing.T) {
	e, err := NewEnforcer("examples/ai_policy_model.conf")
	if err != nil {
		t.Fatal(err)
	}

	// Add multiple AI policies
	rules := [][]string{
		{"allow read access"},
		{"allow write access"},
		{"deny delete access"},
	}

	_, err = e.AddAIPolicies(rules)
	if err != nil {
		t.Fatalf("AddAIPolicies failed: %v", err)
	}

	// Test filtering
	filtered, err := e.GetFilteredAIPolicy(0, "allow read access")
	if err != nil {
		t.Fatalf("GetFilteredAIPolicy failed: %v", err)
	}
	if len(filtered) != 1 {
		t.Errorf("Expected 1 filtered AI policy, got %d", len(filtered))
	}
	if filtered[0][0] != "allow read access" {
		t.Errorf("Expected 'allow read access', got %s", filtered[0][0])
	}
}

// TestRemoveFilteredAIPolicy tests removing filtered AI policies.
func TestRemoveFilteredAIPolicy(t *testing.T) {
	e, err := NewEnforcer("examples/ai_policy_model.conf")
	if err != nil {
		t.Fatal(err)
	}

	// Add multiple AI policies
	rules := [][]string{
		{"allow read access to public data"},
		{"allow read access to private data"},
		{"deny write access"},
	}

	_, err = e.AddAIPolicies(rules)
	if err != nil {
		t.Fatalf("AddAIPolicies failed: %v", err)
	}

	// Remove policies that start with "allow read"
	// Note: This removes based on exact match at the specified field index
	removed, err := e.RemoveFilteredAIPolicy(0, "allow read access to public data")
	if err != nil {
		t.Fatalf("RemoveFilteredAIPolicy failed: %v", err)
	}
	if !removed {
		t.Error("Expected AI policies to be removed")
	}

	// Verify
	policies, err := e.GetAIPolicy()
	if err != nil {
		t.Fatalf("GetAIPolicy failed: %v", err)
	}
	if len(policies) != 2 {
		t.Errorf("Expected 2 AI policies after removal, got %d", len(policies))
	}
}
