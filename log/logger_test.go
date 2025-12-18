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

package log

import (
	"testing"
	"time"
)

// TestLogger is a simple mock logger for testing purposes.
type TestLogger struct {
	enabled   bool
	subscribe []EventType

	// Capture fields for assertions
	LastEntry  *LogEntry
	LastHandle *Handle
}

func NewTestLogger() *TestLogger {
	return &TestLogger{
		enabled: true,
	}
}

func (l *TestLogger) EnableLog(enable bool) {
	l.enabled = enable
}

func (l *TestLogger) IsEnabled() bool {
	return l.enabled
}

func (l *TestLogger) LogModel(model [][]string) {}

func (l *TestLogger) LogEnforce(matcher string, request []interface{}, result bool, explains [][]string) {
}

func (l *TestLogger) LogRole(roles []string) {}

func (l *TestLogger) LogPolicy(policy map[string][][]string) {}

func (l *TestLogger) LogError(err error, msg ...string) {}

func (l *TestLogger) Subscribe() []EventType {
	return l.subscribe
}

func (l *TestLogger) SetSubscribe(events []EventType) {
	l.subscribe = events
}

func (l *TestLogger) OnBeforeEvent(entry *LogEntry) *Handle {
	l.LastEntry = entry
	handle := NewHandle()
	l.LastHandle = handle
	return handle
}

func (l *TestLogger) OnAfterEvent(handle *Handle, entry *LogEntry) {
	l.LastEntry = entry
	l.LastHandle = handle
}

func TestEventType(t *testing.T) {
	if EventEnforce != "enforce" {
		t.Errorf("EventEnforce should be 'enforce', got '%s'", EventEnforce)
	}
	if EventPolicyAdd != "policy.add" {
		t.Errorf("EventPolicyAdd should be 'policy.add', got '%s'", EventPolicyAdd)
	}
	if EventPolicyRemove != "policy.remove" {
		t.Errorf("EventPolicyRemove should be 'policy.remove', got '%s'", EventPolicyRemove)
	}
	if EventPolicyUpdate != "policy.update" {
		t.Errorf("EventPolicyUpdate should be 'policy.update', got '%s'", EventPolicyUpdate)
	}
	if EventPolicyLoad != "policy.load" {
		t.Errorf("EventPolicyLoad should be 'policy.load', got '%s'", EventPolicyLoad)
	}
	if EventPolicySave != "policy.save" {
		t.Errorf("EventPolicySave should be 'policy.save', got '%s'", EventPolicySave)
	}
	if EventModelLoad != "model.load" {
		t.Errorf("EventModelLoad should be 'model.load', got '%s'", EventModelLoad)
	}
	if EventRoleAdd != "role.add" {
		t.Errorf("EventRoleAdd should be 'role.add', got '%s'", EventRoleAdd)
	}
	if EventRoleRemove != "role.remove" {
		t.Errorf("EventRoleRemove should be 'role.remove', got '%s'", EventRoleRemove)
	}
}

func TestHandle(t *testing.T) {
	handle := NewHandle()

	if handle.StartTime.IsZero() {
		t.Error("Handle StartTime should not be zero")
	}
	if handle.Store == nil {
		t.Error("Handle Store should not be nil")
	}

	// Test storing custom data
	handle.Store["key"] = "value"
	if handle.Store["key"] != "value" {
		t.Error("Handle Store should be able to store data")
	}

	// Test storing complex data
	handle.Store["context"] = map[string]interface{}{"trace_id": "123"}
	ctx, ok := handle.Store["context"].(map[string]interface{})
	if !ok {
		t.Error("Handle Store should be able to store complex data")
	}
	if ctx["trace_id"] != "123" {
		t.Error("Handle Store should preserve complex data")
	}
}

func TestLogEntry(t *testing.T) {
	entry := NewLogEntry(EventEnforce)

	if entry.Type != EventEnforce {
		t.Errorf("LogEntry Type should be EventEnforce, got '%s'", entry.Type)
	}
	if entry.Timestamp.IsZero() {
		t.Error("LogEntry Timestamp should not be zero")
	}
	if entry.Attributes == nil {
		t.Error("LogEntry Attributes should not be nil")
	}

	// Test setting fields
	entry.Subject = "alice"
	entry.Object = "data1"
	entry.Action = "read"
	entry.Allowed = true
	entry.Matched = [][]string{{"alice", "data1", "read"}}

	if entry.Subject != "alice" {
		t.Errorf("LogEntry Subject should be 'alice', got '%s'", entry.Subject)
	}
	if entry.Object != "data1" {
		t.Errorf("LogEntry Object should be 'data1', got '%s'", entry.Object)
	}
	if entry.Action != "read" {
		t.Errorf("LogEntry Action should be 'read', got '%s'", entry.Action)
	}
	if !entry.Allowed {
		t.Error("LogEntry Allowed should be true")
	}
}

func TestLogEntryPolicyFields(t *testing.T) {
	entry := NewLogEntry(EventPolicyAdd)

	entry.Operation = "add"
	entry.Rules = [][]string{{"alice", "data1", "read"}, {"bob", "data2", "write"}}
	entry.RuleCount = 2

	if entry.Operation != "add" {
		t.Errorf("LogEntry Operation should be 'add', got '%s'", entry.Operation)
	}
	if len(entry.Rules) != 2 {
		t.Errorf("LogEntry Rules should have 2 items, got %d", len(entry.Rules))
	}
	if entry.RuleCount != 2 {
		t.Errorf("LogEntry RuleCount should be 2, got %d", entry.RuleCount)
	}
}

func TestDefaultLoggerSubscribe(t *testing.T) {
	logger := &DefaultLogger{}

	// Default should be nil (subscribe to all)
	if logger.Subscribe() != nil {
		t.Error("Default subscribe should be nil")
	}

	// Set specific events
	logger.SetSubscribe([]EventType{EventEnforce, EventPolicyAdd})
	events := logger.Subscribe()
	if len(events) != 2 {
		t.Errorf("Subscribe should return 2 events, got %d", len(events))
	}
	if events[0] != EventEnforce {
		t.Errorf("First event should be EventEnforce, got '%s'", events[0])
	}
	if events[1] != EventPolicyAdd {
		t.Errorf("Second event should be EventPolicyAdd, got '%s'", events[1])
	}

	// Reset to nil
	logger.SetSubscribe(nil)
	if logger.Subscribe() != nil {
		t.Error("Subscribe should be nil after reset")
	}
}

func TestDefaultLoggerOnBeforeAfterEvent(t *testing.T) {
	logger := &DefaultLogger{}
	entry := NewLogEntry(EventEnforce)
	entry.Subject = "alice"
	entry.Object = "data1"
	entry.Action = "read"

	handle := logger.OnBeforeEvent(entry)
	if handle == nil {
		t.Fatal("OnBeforeEvent should return a Handle")
	}
	if handle.Store == nil {
		t.Fatal("Handle Store should not be nil")
	}

	// Simulate some operation
	time.Sleep(1 * time.Millisecond)

	entry.Duration = time.Since(entry.Timestamp)
	entry.Allowed = true

	// OnAfterEvent should not panic
	logger.OnAfterEvent(handle, entry)

	if entry.Duration <= 0 {
		t.Error("LogEntry Duration should be greater than 0")
	}
}

func TestLogUtilFunctions(t *testing.T) {
	// Reset to default logger for this test
	SetLogger(&DefaultLogger{})

	// Test that global functions work with default logger
	events := Subscribe()
	if events != nil {
		t.Error("Default logger should subscribe to all events (nil)")
	}

	entry := NewLogEntry(EventEnforce)
	handle := OnBeforeEvent(entry)
	if handle == nil {
		t.Error("OnBeforeEvent should return a Handle")
	}

	// OnAfterEvent should not panic
	OnAfterEvent(handle, entry)
}
