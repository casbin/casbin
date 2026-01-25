// Copyright 2017 The casbin Authors. All Rights Reserved.
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
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// AIConfig contains configuration for AI API calls.
type AIConfig struct {
	// Endpoint is the API endpoint (e.g., "https://api.openai.com/v1/chat/completions")
	Endpoint string
	// APIKey is the authentication key for the API
	APIKey string
	// Model is the model to use (e.g., "gpt-3.5-turbo", "gpt-4")
	Model string
	// Timeout for API requests (default: 30s)
	Timeout time.Duration
}

// aiMessage represents a message in the OpenAI chat format.
type aiMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// aiChatRequest represents the request to OpenAI chat completions API.
type aiChatRequest struct {
	Model    string      `json:"model"`
	Messages []aiMessage `json:"messages"`
}

// aiChatResponse represents the response from OpenAI chat completions API.
type aiChatResponse struct {
	Choices []struct {
		Message aiMessage `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// SetAIConfig sets the configuration for AI API calls.
func (e *Enforcer) SetAIConfig(config AIConfig) {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	e.aiConfig = config
}

// Explain returns an AI-generated explanation of why Enforce returned a particular result.
// It calls the configured OpenAI-compatible API to generate a natural language explanation.
func (e *Enforcer) Explain(rvals ...interface{}) (string, error) {
	if e.aiConfig.Endpoint == "" {
		return "", errors.New("AI config not set, use SetAIConfig first")
	}

	// Get enforcement result and matched rules
	result, matchedRules, err := e.EnforceEx(rvals...)
	if err != nil {
		return "", fmt.Errorf("failed to enforce: %w", err)
	}

	// Build context for AI
	explainContext := e.buildExplainContext(rvals, result, matchedRules)

	// Call AI API
	explanation, err := e.callAIAPI(explainContext)
	if err != nil {
		return "", fmt.Errorf("failed to get AI explanation: %w", err)
	}

	return explanation, nil
}

// buildExplainContext builds the context string for AI explanation.
func (e *Enforcer) buildExplainContext(rvals []interface{}, result bool, matchedRules []string) string {
	var sb strings.Builder

	// Add request information
	sb.WriteString("Authorization Request:\n")
	sb.WriteString(fmt.Sprintf("Subject: %v\n", rvals[0]))
	if len(rvals) > 1 {
		sb.WriteString(fmt.Sprintf("Object: %v\n", rvals[1]))
	}
	if len(rvals) > 2 {
		sb.WriteString(fmt.Sprintf("Action: %v\n", rvals[2]))
	}
	sb.WriteString(fmt.Sprintf("\nEnforcement Result: %v\n", result))

	// Add matched rules
	if len(matchedRules) > 0 {
		sb.WriteString("\nMatched Policy Rules:\n")
		for _, rule := range matchedRules {
			sb.WriteString(fmt.Sprintf("- %s\n", rule))
		}
	} else {
		sb.WriteString("\nNo policy rules matched.\n")
	}

	// Add model information
	sb.WriteString("\nAccess Control Model:\n")
	if m, ok := e.model["m"]; ok {
		for key, ast := range m {
			sb.WriteString(fmt.Sprintf("Matcher (%s): %s\n", key, ast.Value))
		}
	}
	if eff, ok := e.model["e"]; ok {
		for key, ast := range eff {
			sb.WriteString(fmt.Sprintf("Effect (%s): %s\n", key, ast.Value))
		}
	}

	// Add all policies
	policies, _ := e.GetPolicy()
	if len(policies) > 0 {
		sb.WriteString("\nAll Policy Rules:\n")
		for _, policy := range policies {
			sb.WriteString(fmt.Sprintf("- %s\n", strings.Join(policy, ", ")))
		}
	}

	return sb.String()
}

// callAIAPI calls the configured AI API to get an explanation.
func (e *Enforcer) callAIAPI(explainContext string) (string, error) {
	// Prepare the request
	messages := []aiMessage{
		{
			Role: "system",
			Content: "You are an expert in access control and authorization systems. " +
				"Explain why an authorization request was allowed or denied based on the " +
				"provided access control model, policies, and enforcement result. " +
				"Be clear, concise, and educational.",
		},
		{
			Role:    "user",
			Content: fmt.Sprintf("Please explain the following authorization decision:\n\n%s", explainContext),
		},
	}

	reqBody := aiChatRequest{
		Model:    e.aiConfig.Model,
		Messages: messages,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request with context
	reqCtx, cancel := context.WithTimeout(context.Background(), e.aiConfig.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, e.aiConfig.Endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+e.aiConfig.APIKey)

	// Execute request
	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Parse response
	var chatResp aiChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for API errors
	if chatResp.Error != nil {
		return "", fmt.Errorf("API error: %s", chatResp.Error.Message)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Extract explanation
	if len(chatResp.Choices) == 0 {
		return "", errors.New("no response from AI")
	}

	return chatResp.Choices[0].Message.Content, nil
}
