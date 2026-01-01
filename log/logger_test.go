// Copyright 2025 The casbin Authors. All Rights Reserved.
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

package log

import (
	"testing"
	"time"
)

// TestEventTypeConstants verifies all event type constants are defined correctly.
func TestEventTypeConstants(t *testing.T) {
	tests := []struct {
		name     string
		event    EventType
		expected string
	}{
		{"EventEnforce", EventEnforce, "enforce"},
		{"EventPolicyAdd", EventPolicyAdd, "policy_add"},
		{"EventPolicyRemove", EventPolicyRemove, "policy_remove"},
		{"EventPolicyUpdate", EventPolicyUpdate, "policy_update"},
		{"EventPolicyLoad", EventPolicyLoad, "policy_load"},
		{"EventPolicySave", EventPolicySave, "policy_save"},
		{"EventModelLoad", EventModelLoad, "model_load"},
		{"EventRoleAdd", EventRoleAdd, "role_add"},
		{"EventRoleRemove", EventRoleRemove, "role_remove"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.event) != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, string(tt.event))
			}
		})
	}
}

// TestHandleStructure verifies the Handle struct has the required fields.
func TestHandleStructure(t *testing.T) {
	now := time.Now()
	store := map[string]interface{}{
		"key1": "value1",
		"key2": 123,
	}

	handle := Handle{
		StartTime: now,
		Store:     store,
	}

	if !handle.StartTime.Equal(now) {
		t.Errorf("Expected StartTime to be %v, got %v", now, handle.StartTime)
	}

	if handle.Store["key1"] != "value1" {
		t.Errorf("Expected Store[key1] to be 'value1', got %v", handle.Store["key1"])
	}
}

// TestLogEntryStructure verifies the LogEntry struct has all required fields.
func TestLogEntryStructure(t *testing.T) {
	now := time.Now()
	duration := 100 * time.Millisecond

	entry := LogEntry{
		EventType: EventEnforce,
		Timestamp: now,
		Duration:  duration,
		Subject:   "alice",
		Object:    "data1",
		Action:    "read",
		Domain:    "domain1",
		Allowed:   true,
		Matched:   "p, alice, data1, read",
		Operation: "add",
		Rules:     [][]string{{"alice", "data1", "read"}},
		RuleCount: 1,
		Error:     nil,
		Attributes: map[string]interface{}{
			"custom": "value",
		},
	}

	if entry.EventType != EventEnforce {
		t.Errorf("Expected EventType to be EventEnforce, got %v", entry.EventType)
	}

	if !entry.Timestamp.Equal(now) {
		t.Errorf("Expected Timestamp to be %v, got %v", now, entry.Timestamp)
	}

	if entry.Duration != duration {
		t.Errorf("Expected Duration to be %v, got %v", duration, entry.Duration)
	}

	if entry.Subject != "alice" {
		t.Errorf("Expected Subject to be 'alice', got %v", entry.Subject)
	}

	if entry.Object != "data1" {
		t.Errorf("Expected Object to be 'data1', got %v", entry.Object)
	}

	if entry.Action != "read" {
		t.Errorf("Expected Action to be 'read', got %v", entry.Action)
	}

	if entry.Domain != "domain1" {
		t.Errorf("Expected Domain to be 'domain1', got %v", entry.Domain)
	}

	if entry.Allowed != true {
		t.Errorf("Expected Allowed to be true, got %v", entry.Allowed)
	}

	if entry.Matched != "p, alice, data1, read" {
		t.Errorf("Expected Matched to be 'p, alice, data1, read', got %v", entry.Matched)
	}

	if entry.Operation != "add" {
		t.Errorf("Expected Operation to be 'add', got %v", entry.Operation)
	}

	if len(entry.Rules) != 1 || len(entry.Rules[0]) != 3 {
		t.Errorf("Expected Rules to have 1 rule with 3 elements, got %v", entry.Rules)
	}

	if entry.RuleCount != 1 {
		t.Errorf("Expected RuleCount to be 1, got %v", entry.RuleCount)
	}

	if entry.Error != nil {
		t.Errorf("Expected Error to be nil, got %v", entry.Error)
	}

	if entry.Attributes["custom"] != "value" {
		t.Errorf("Expected Attributes['custom'] to be 'value', got %v", entry.Attributes["custom"])
	}
}

// mockLogger is a simple implementation of the Logger interface for testing.
type mockLogger struct {
	enabled      bool
	events       []EventType
	beforeCalled bool
	afterCalled  bool
}

func (m *mockLogger) Enable(enabled bool) {
	m.enabled = enabled
}

func (m *mockLogger) IsEnabled() bool {
	return m.enabled
}

func (m *mockLogger) Subscribe() []EventType {
	return m.events
}

func (m *mockLogger) OnBeforeEvent(entry *LogEntry) *Handle {
	m.beforeCalled = true
	return &Handle{
		StartTime: time.Now(),
		Store:     make(map[string]interface{}),
	}
}

func (m *mockLogger) OnAfterEvent(handle *Handle, entry *LogEntry) {
	m.afterCalled = true
}

// TestLoggerInterface verifies the Logger interface can be implemented.
func TestLoggerInterface(t *testing.T) {
	logger := &mockLogger{
		enabled: true,
		events:  []EventType{EventEnforce, EventPolicyAdd},
	}

	// Test Enable/IsEnabled.
	if !logger.IsEnabled() {
		t.Error("Expected logger to be enabled")
	}

	logger.Enable(false)
	if logger.IsEnabled() {
		t.Error("Expected logger to be disabled")
	}

	// Test Subscribe.
	events := logger.Subscribe()
	if len(events) != 2 {
		t.Errorf("Expected 2 events, got %d", len(events))
	}

	// Test OnBeforeEvent.
	entry := &LogEntry{EventType: EventEnforce}
	handle := logger.OnBeforeEvent(entry)

	if !logger.beforeCalled {
		t.Error("Expected OnBeforeEvent to be called")
	}

	if handle == nil {
		t.Error("Expected handle to be non-nil")
	}

	// Test OnAfterEvent.
	logger.OnAfterEvent(handle, entry)

	if !logger.afterCalled {
		t.Error("Expected OnAfterEvent to be called")
	}
}
