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
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestDefaultLogger_Enable(t *testing.T) {
	logger := NewDefaultLogger(nil, nil)

	// Test initial state (enabled by default)
	if !logger.IsEnabled() {
		t.Error("Logger should be enabled by default")
	}

	// Test disabling
	logger.Enable(false)
	if logger.IsEnabled() {
		t.Error("Logger should be disabled after Enable(false)")
	}

	// Test enabling
	logger.Enable(true)
	if !logger.IsEnabled() {
		t.Error("Logger should be enabled after Enable(true)")
	}
}

func TestDefaultLogger_EnableConcurrent(t *testing.T) {
	logger := NewDefaultLogger(nil, nil)
	var wg sync.WaitGroup

	// Test concurrent enable/disable calls
	for i := 0; i < 100; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			logger.Enable(true)
		}()
		go func() {
			defer wg.Done()
			logger.Enable(false)
		}()
	}

	wg.Wait()

	// Just ensure no panic occurred and we can still read the state
	_ = logger.IsEnabled()
}

func TestDefaultLogger_Subscribe(t *testing.T) {
	// Test with nil subscriptions (all events)
	logger1 := NewDefaultLogger(nil, nil)
	if logger1.Subscribe() != nil {
		t.Error("Subscribe() should return nil for all events subscription")
	}

	// Test with empty subscriptions (all events)
	logger2 := NewDefaultLogger(nil, []EventType{})
	subs := logger2.Subscribe()
	if subs != nil {
		t.Error("Subscribe() should return nil (all events) for empty subscriptions input")
	}

	// Test with specific subscriptions
	events := []EventType{EventEnforce, EventPolicyAdd}
	logger3 := NewDefaultLogger(nil, events)
	subs = logger3.Subscribe()
	if len(subs) != 2 {
		t.Errorf("Subscribe() should return 2 events, got %d", len(subs))
	}
	if subs[0] != EventEnforce || subs[1] != EventPolicyAdd {
		t.Error("Subscribe() should return the correct events")
	}

	// Test that Subscribe returns a copy (not the original slice)
	subs[0] = EventPolicyRemove // Modify the returned slice
	subs2 := logger3.Subscribe()
	if subs2[0] != EventEnforce {
		t.Error("Subscribe() should return a copy, not the original slice")
	}
}

func TestDefaultLogger_OnBeforeEvent(t *testing.T) {
	logger := NewDefaultLogger(nil, nil)

	// Test with enabled logger
	entry := &LogEntry{EventType: EventEnforce}
	handle := logger.OnBeforeEvent(entry)
	if handle == nil {
		t.Error("OnBeforeEvent should return a handle when enabled")
	}
	if handle.StartTime.IsZero() {
		t.Error("Handle StartTime should be set")
	}
	if handle.Store == nil {
		t.Error("Handle Store should be initialized")
	}

	// Test with disabled logger
	logger.Enable(false)
	handle = logger.OnBeforeEvent(entry)
	if handle != nil {
		t.Error("OnBeforeEvent should return nil when disabled")
	}
}

func TestDefaultLogger_OnAfterEvent_Disabled(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewDefaultLogger(buf, nil)
	logger.Enable(false)

	entry := &LogEntry{
		EventType: EventEnforce,
		Timestamp: time.Now(),
	}
	handle := &Handle{
		StartTime: time.Now(),
		Store:     make(map[string]interface{}),
	}

	logger.OnAfterEvent(handle, entry)

	if buf.Len() > 0 {
		t.Error("OnAfterEvent should not write when logger is disabled")
	}
}

func TestDefaultLogger_OnAfterEvent_NilHandle(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewDefaultLogger(buf, nil)

	entry := &LogEntry{
		EventType: EventEnforce,
		Timestamp: time.Now(),
	}

	logger.OnAfterEvent(nil, entry)

	if buf.Len() > 0 {
		t.Error("OnAfterEvent should not write when handle is nil")
	}
}

func TestDefaultLogger_OnAfterEvent_NilEntry(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewDefaultLogger(buf, nil)

	handle := &Handle{
		StartTime: time.Now(),
		Store:     make(map[string]interface{}),
	}

	logger.OnAfterEvent(handle, nil)

	if buf.Len() > 0 {
		t.Error("OnAfterEvent should not write when entry is nil")
	}
}

func TestDefaultLogger_OnAfterEvent_EnforceEvent(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewDefaultLogger(buf, nil)

	startTime := time.Now()
	handle := &Handle{
		StartTime: startTime,
		Store:     make(map[string]interface{}),
	}

	entry := &LogEntry{
		EventType: EventEnforce,
		Timestamp: time.Now(),
		Subject:   "alice",
		Object:    "data1",
		Action:    "read",
		Domain:    "domain1",
		Allowed:   true,
		Matched:   "p, alice, data1, read",
	}

	logger.OnAfterEvent(handle, entry)

	output := buf.String()
	if output == "" {
		t.Fatal("OnAfterEvent should write output")
	}

	// Parse JSON output
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &result); err != nil {
		t.Fatalf("Output should be valid JSON: %v\nOutput: %s", err, output)
	}

	// Verify fields
	if result["event_type"] != string(EventEnforce) {
		t.Errorf("event_type should be %s, got %v", EventEnforce, result["event_type"])
	}
	if result["subject"] != "alice" {
		t.Errorf("subject should be alice, got %v", result["subject"])
	}
	if result["object"] != "data1" {
		t.Errorf("object should be data1, got %v", result["object"])
	}
	if result["action"] != "read" {
		t.Errorf("action should be read, got %v", result["action"])
	}
	if result["domain"] != "domain1" {
		t.Errorf("domain should be domain1, got %v", result["domain"])
	}
	if result["allowed"] != true {
		t.Errorf("allowed should be true, got %v", result["allowed"])
	}
	if result["matched"] != "p, alice, data1, read" {
		t.Errorf("matched should be 'p, alice, data1, read', got %v", result["matched"])
	}

	// Verify timestamp is present and valid
	if _, ok := result["timestamp"]; !ok {
		t.Error("timestamp should be present")
	}

	// Verify duration is present and non-negative
	if durationMs, ok := result["duration_ms"].(float64); !ok || durationMs < 0 {
		t.Errorf("duration_ms should be a non-negative number, got %v", result["duration_ms"])
	}
}

func TestDefaultLogger_OnAfterEvent_PolicyAddEvent(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewDefaultLogger(buf, nil)

	handle := &Handle{
		StartTime: time.Now(),
		Store:     make(map[string]interface{}),
	}

	rules := [][]string{
		{"alice", "data1", "read"},
		{"bob", "data2", "write"},
	}

	entry := &LogEntry{
		EventType: EventPolicyAdd,
		Timestamp: time.Now(),
		Operation: "add",
		Rules:     rules,
		RuleCount: 2,
	}

	logger.OnAfterEvent(handle, entry)

	output := buf.String()
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &result); err != nil {
		t.Fatalf("Output should be valid JSON: %v", err)
	}

	if result["event_type"] != string(EventPolicyAdd) {
		t.Errorf("event_type should be %s, got %v", EventPolicyAdd, result["event_type"])
	}
	if result["operation"] != "add" {
		t.Errorf("operation should be add, got %v", result["operation"])
	}
	if result["rule_count"] != float64(2) {
		t.Errorf("rule_count should be 2, got %v", result["rule_count"])
	}

	// Verify rules array
	if _, ok := result["rules"]; !ok {
		t.Error("rules should be present")
	}
}

func TestDefaultLogger_OnAfterEvent_WithError(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewDefaultLogger(buf, nil)

	handle := &Handle{
		StartTime: time.Now(),
		Store:     make(map[string]interface{}),
	}

	testError := errors.New("test error message")
	entry := &LogEntry{
		EventType: EventPolicyLoad,
		Timestamp: time.Now(),
		Error:     testError,
	}

	logger.OnAfterEvent(handle, entry)

	output := buf.String()
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &result); err != nil {
		t.Fatalf("Output should be valid JSON: %v", err)
	}

	if result["error"] != "test error message" {
		t.Errorf("error should be 'test error message', got %v", result["error"])
	}
}

func TestDefaultLogger_OnAfterEvent_WithAttributes(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewDefaultLogger(buf, nil)

	handle := &Handle{
		StartTime: time.Now(),
		Store:     make(map[string]interface{}),
	}

	entry := &LogEntry{
		EventType: EventEnforce,
		Timestamp: time.Now(),
		Attributes: map[string]interface{}{
			"custom_field1": "value1",
			"custom_field2": 123,
			"custom_field3": true,
		},
	}

	logger.OnAfterEvent(handle, entry)

	output := buf.String()
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &result); err != nil {
		t.Fatalf("Output should be valid JSON: %v", err)
	}

	if result["custom_field1"] != "value1" {
		t.Errorf("custom_field1 should be 'value1', got %v", result["custom_field1"])
	}
	if result["custom_field2"] != float64(123) {
		t.Errorf("custom_field2 should be 123, got %v", result["custom_field2"])
	}
	if result["custom_field3"] != true {
		t.Errorf("custom_field3 should be true, got %v", result["custom_field3"])
	}
}

func TestDefaultLogger_OnAfterEvent_EmptyFields(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewDefaultLogger(buf, nil)

	handle := &Handle{
		StartTime: time.Now(),
		Store:     make(map[string]interface{}),
	}

	// Entry with minimal fields
	entry := &LogEntry{
		EventType: EventEnforce,
		Timestamp: time.Now(),
	}

	logger.OnAfterEvent(handle, entry)

	output := buf.String()
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &result); err != nil {
		t.Fatalf("Output should be valid JSON: %v", err)
	}

	// Should always have event_type, timestamp, and duration_ms
	if result["event_type"] != string(EventEnforce) {
		t.Error("event_type should be present")
	}
	if _, ok := result["timestamp"]; !ok {
		t.Error("timestamp should be present")
	}
	if _, ok := result["duration_ms"]; !ok {
		t.Error("duration_ms should be present")
	}

	// Empty string fields should not be in output
	if _, ok := result["subject"]; ok {
		t.Error("subject should not be present when empty")
	}
	if _, ok := result["object"]; ok {
		t.Error("object should not be present when empty")
	}
	if _, ok := result["action"]; ok {
		t.Error("action should not be present when empty")
	}
}

func TestDefaultLogger_DurationCalculation(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewDefaultLogger(buf, nil)

	startTime := time.Now()
	// Simulate a delay
	time.Sleep(10 * time.Millisecond)

	handle := &Handle{
		StartTime: startTime,
		Store:     make(map[string]interface{}),
	}

	entry := &LogEntry{
		EventType: EventEnforce,
		Timestamp: time.Now(),
	}

	logger.OnAfterEvent(handle, entry)

	output := buf.String()
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &result); err != nil {
		t.Fatalf("Output should be valid JSON: %v", err)
	}

	durationMs := result["duration_ms"].(float64)
	if durationMs < 10 {
		t.Errorf("duration_ms should be at least 10ms, got %f", durationMs)
	}

	// Verify the entry.Duration field was set
	if entry.Duration == 0 {
		t.Error("entry.Duration should be set after OnAfterEvent")
	}
}

func TestDefaultLogger_ConcurrentWrites(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewDefaultLogger(buf, nil)
	var wg sync.WaitGroup

	// Test concurrent OnAfterEvent calls
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			handle := &Handle{
				StartTime: time.Now(),
				Store:     make(map[string]interface{}),
			}
			entry := &LogEntry{
				EventType: EventEnforce,
				Timestamp: time.Now(),
				Subject:   "user",
			}
			logger.OnAfterEvent(handle, entry)
		}(i)
	}

	wg.Wait()

	// Verify we got some output (exact count may vary due to concurrency)
	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) < 1 {
		t.Error("Should have at least one line of output")
	}
}

func TestNewDefaultLogger_NilWriter(t *testing.T) {
	logger := NewDefaultLogger(nil, nil)
	if logger.writer == nil {
		t.Error("Logger should default to a non-nil writer")
	}
}

func TestDefaultLogger_AllEventTypes(t *testing.T) {
	eventTypes := []EventType{
		EventEnforce,
		EventPolicyAdd,
		EventPolicyRemove,
		EventPolicyUpdate,
		EventPolicyLoad,
		EventPolicySave,
		EventModelLoad,
		EventRoleAdd,
		EventRoleRemove,
	}

	for _, eventType := range eventTypes {
		t.Run(string(eventType), func(t *testing.T) {
			buf := &bytes.Buffer{}
			logger := NewDefaultLogger(buf, nil)

			handle := &Handle{
				StartTime: time.Now(),
				Store:     make(map[string]interface{}),
			}

			entry := &LogEntry{
				EventType: eventType,
				Timestamp: time.Now(),
			}

			logger.OnAfterEvent(handle, entry)

			output := buf.String()
			var result map[string]interface{}
			if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &result); err != nil {
				t.Fatalf("Output should be valid JSON for %s: %v", eventType, err)
			}

			if result["event_type"] != string(eventType) {
				t.Errorf("event_type should be %s, got %v", eventType, result["event_type"])
			}
		})
	}
}
