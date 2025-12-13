// Copyright 2024 The casbin Authors. All Rights Reserved.
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

// TestEventLogger is a test logger implementation that captures events
type TestEventLogger struct {
	enabled   bool
	subscribe []EventType
	events    []*LogEntry
}

func NewTestEventLogger(subscribe ...EventType) *TestEventLogger {
	return &TestEventLogger{
		enabled:   true,
		subscribe: subscribe,
		events:    make([]*LogEntry, 0),
	}
}

func (l *TestEventLogger) Enable(enabled bool) {
	l.enabled = enabled
}

func (l *TestEventLogger) IsEnabled() bool {
	return l.enabled
}

func (l *TestEventLogger) Subscribe() []EventType {
	return l.subscribe
}

func (l *TestEventLogger) OnBeforeEvent(entry *LogEntry) *Handle {
	return NewHandle()
}

func (l *TestEventLogger) OnAfterEvent(handle *Handle, entry *LogEntry) {
	// Store a copy of the entry
	l.events = append(l.events, entry)
}

func (l *TestEventLogger) GetEvents() []*LogEntry {
	return l.events
}

func (l *TestEventLogger) Clear() {
	l.events = make([]*LogEntry, 0)
}

func TestNewHandle(t *testing.T) {
	handle := NewHandle()
	if handle == nil {
		t.Error("NewHandle should not return nil")
	}
	if handle.Store == nil {
		t.Error("Handle.Store should be initialized")
	}
	if handle.StartTime.IsZero() {
		t.Error("Handle.StartTime should be set")
	}
}

func TestDefaultEventLogger(t *testing.T) {
	logger := NewDefaultEventLogger()
	
	if logger.IsEnabled() {
		t.Error("DefaultEventLogger should be disabled by default")
	}
	
	logger.Enable(true)
	if !logger.IsEnabled() {
		t.Error("DefaultEventLogger should be enabled after Enable(true)")
	}
	
	if logger.Subscribe() != nil {
		t.Error("DefaultEventLogger should subscribe to all events by default (nil)")
	}
	
	// Test that it doesn't panic when called
	entry := &LogEntry{
		Type:      EventEnforce,
		Timestamp: time.Now(),
	}
	handle := logger.OnBeforeEvent(entry)
	if handle == nil {
		t.Error("OnBeforeEvent should return a handle")
	}
	
	logger.OnAfterEvent(handle, entry)
}

func TestTestEventLogger(t *testing.T) {
	logger := NewTestEventLogger(EventEnforce, EventPolicyAdd)
	
	if !logger.IsEnabled() {
		t.Error("TestEventLogger should be enabled by default")
	}
	
	subscribe := logger.Subscribe()
	if len(subscribe) != 2 {
		t.Errorf("Expected 2 subscriptions, got %d", len(subscribe))
	}
	
	// Test event capture
	entry := &LogEntry{
		Type:      EventEnforce,
		Timestamp: time.Now(),
		Subject:   "alice",
		Object:    "data1",
		Action:    "read",
		Allowed:   true,
	}
	
	handle := logger.OnBeforeEvent(entry)
	entry.Duration = time.Since(handle.StartTime)
	logger.OnAfterEvent(handle, entry)
	
	events := logger.GetEvents()
	if len(events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(events))
	}
	
	if events[0].Type != EventEnforce {
		t.Errorf("Expected EventEnforce, got %v", events[0].Type)
	}
	
	if events[0].Subject != "alice" {
		t.Errorf("Expected subject 'alice', got %s", events[0].Subject)
	}
}

func TestEventTypes(t *testing.T) {
	types := []EventType{
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
	
	for _, eventType := range types {
		if string(eventType) == "" {
			t.Errorf("Event type should not be empty: %v", eventType)
		}
	}
}

func TestLogEntry(t *testing.T) {
	entry := &LogEntry{
		Type:       EventEnforce,
		Timestamp:  time.Now(),
		Duration:   time.Millisecond * 10,
		Request:    []interface{}{"alice", "data1", "read"},
		Subject:    "alice",
		Object:     "data1",
		Action:     "read",
		Allowed:    true,
		Matched:    [][]string{{"alice", "data1", "read"}},
		Operation:  "",
		Rules:      nil,
		RuleCount:  0,
		Error:      nil,
		Attributes: make(map[string]interface{}),
	}
	
	if entry.Type != EventEnforce {
		t.Error("Entry type mismatch")
	}
	
	if entry.Subject != "alice" {
		t.Error("Entry subject mismatch")
	}
	
	if !entry.Allowed {
		t.Error("Entry should be allowed")
	}
	
	if entry.Duration < time.Millisecond {
		t.Error("Entry duration too short")
	}
}
