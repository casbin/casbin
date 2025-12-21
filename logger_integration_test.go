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

package casbin

import (
	"testing"

	"github.com/casbin/casbin/v3/log"
)

// TestLogger is a test logger implementation that captures events
type TestLogger struct {
	enabled   bool
	subscribe []log.EventType
	events    []*log.LogEntry
}

func NewTestLogger(subscribe ...log.EventType) *TestLogger {
	return &TestLogger{
		enabled:   true,
		subscribe: subscribe,
		events:    make([]*log.LogEntry, 0),
	}
}

func (l *TestLogger) Enable(enabled bool) {
	l.enabled = enabled
}

func (l *TestLogger) IsEnabled() bool {
	return l.enabled
}

func (l *TestLogger) Subscribe() []log.EventType {
	return l.subscribe
}

func (l *TestLogger) OnBeforeEvent(entry *log.LogEntry) *log.Handle {
	return log.NewHandle()
}

func (l *TestLogger) OnAfterEvent(handle *log.Handle, entry *log.LogEntry) {
	// Store a copy of the entry
	l.events = append(l.events, entry)
}

func (l *TestLogger) GetEvents() []*log.LogEntry {
	return l.events
}

func (l *TestLogger) Clear() {
	l.events = make([]*log.LogEntry, 0)
}

func TestEnforcerWithLogger(t *testing.T) {
	e, _ := NewEnforcer("examples/rbac_model.conf", "examples/rbac_policy.csv")
	
	logger := NewTestLogger(log.EventEnforce)
	e.SetLogger(logger)
	
	// Test enforce
	result, _ := e.Enforce("alice", "data1", "read")
	if !result {
		t.Error("alice should be able to read data1")
	}
	
	events := logger.GetEvents()
	if len(events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(events))
	}
	
	event := events[0]
	if event.Type != log.EventEnforce {
		t.Errorf("Expected EventEnforce, got %v", event.Type)
	}
	
	if event.Subject != "alice" {
		t.Errorf("Expected subject 'alice', got %s", event.Subject)
	}
	
	if event.Object != "data1" {
		t.Errorf("Expected object 'data1', got %s", event.Object)
	}
	
	if event.Action != "read" {
		t.Errorf("Expected action 'read', got %s", event.Action)
	}
	
	if !event.Allowed {
		t.Error("Event should show allowed=true")
	}
	
	if event.Error != nil {
		t.Errorf("Event should not have error, got: %v", event.Error)
	}
	
	if event.Duration <= 0 {
		t.Error("Event duration should be > 0")
	}
}

func TestEnforcerPolicyAddEvent(t *testing.T) {
	e, _ := NewEnforcer("examples/rbac_model.conf", "examples/rbac_policy.csv")
	
	logger := NewTestLogger(log.EventPolicyAdd)
	e.SetLogger(logger)
	
	// Test add policy
	ok, _ := e.AddPolicy("eve", "data3", "read")
	if !ok {
		t.Error("AddPolicy should succeed")
	}
	
	events := logger.GetEvents()
	if len(events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(events))
	}
	
	event := events[0]
	if event.Type != log.EventPolicyAdd {
		t.Errorf("Expected EventPolicyAdd, got %v", event.Type)
	}
	
	if event.Operation != "add" {
		t.Errorf("Expected operation 'add', got %s", event.Operation)
	}
	
	if event.RuleCount != 1 {
		t.Errorf("Expected rule count 1, got %d", event.RuleCount)
	}
	
	if len(event.Rules) != 1 {
		t.Errorf("Expected 1 rule, got %d", len(event.Rules))
	}
}

func TestEnforcerPolicyRemoveEvent(t *testing.T) {
	e, _ := NewEnforcer("examples/rbac_model.conf", "examples/rbac_policy.csv")
	
	logger := NewTestLogger(log.EventPolicyRemove)
	e.SetLogger(logger)
	
	// Test remove policy
	ok, _ := e.RemovePolicy("alice", "data1", "read")
	if !ok {
		t.Error("RemovePolicy should succeed")
	}
	
	events := logger.GetEvents()
	if len(events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(events))
	}
	
	event := events[0]
	if event.Type != log.EventPolicyRemove {
		t.Errorf("Expected EventPolicyRemove, got %v", event.Type)
	}
	
	if event.Operation != "remove" {
		t.Errorf("Expected operation 'remove', got %s", event.Operation)
	}
}

func TestEnforcerLoadPolicyEvent(t *testing.T) {
	e, _ := NewEnforcer("examples/rbac_model.conf", "examples/rbac_policy.csv")
	
	logger := NewTestLogger(log.EventPolicyLoad)
	e.SetLogger(logger)
	
	// Test load policy
	err := e.LoadPolicy()
	if err != nil {
		t.Errorf("LoadPolicy should succeed, got error: %v", err)
	}
	
	events := logger.GetEvents()
	if len(events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(events))
	}
	
	event := events[0]
	if event.Type != log.EventPolicyLoad {
		t.Errorf("Expected EventPolicyLoad, got %v", event.Type)
	}
	
	if event.Operation != "load" {
		t.Errorf("Expected operation 'load', got %s", event.Operation)
	}
	
	if event.RuleCount == 0 {
		t.Error("Expected rule count > 0")
	}
}

func TestLoggerSubscription(t *testing.T) {
	e, _ := NewEnforcer("examples/rbac_model.conf", "examples/rbac_policy.csv")
	
	// Subscribe only to enforce events
	logger := NewTestLogger(log.EventEnforce)
	e.SetLogger(logger)
	
	// Enforce should be logged
	e.Enforce("alice", "data1", "read")
	
	// AddPolicy should NOT be logged
	e.AddPolicy("eve", "data3", "read")
	
	events := logger.GetEvents()
	if len(events) != 1 {
		t.Errorf("Expected 1 event (only enforce), got %d", len(events))
	}
	
	if events[0].Type != log.EventEnforce {
		t.Errorf("Expected EventEnforce, got %v", events[0].Type)
	}
}

func TestLoggerSubscribeToAll(t *testing.T) {
	e, _ := NewEnforcer("examples/rbac_model.conf", "examples/rbac_policy.csv")
	
	// Subscribe to all events (pass no arguments)
	logger := NewTestLogger()
	e.SetLogger(logger)
	
	// Multiple operations
	e.Enforce("alice", "data1", "read")
	e.AddPolicy("eve", "data3", "read")
	e.RemovePolicy("eve", "data3", "read")
	
	events := logger.GetEvents()
	if len(events) != 3 {
		t.Errorf("Expected 3 events, got %d", len(events))
	}
	
	// Verify event types
	expectedTypes := []log.EventType{
		log.EventEnforce,
		log.EventPolicyAdd,
		log.EventPolicyRemove,
	}
	
	for i, expected := range expectedTypes {
		if events[i].Type != expected {
			t.Errorf("Event %d: expected %v, got %v", i, expected, events[i].Type)
		}
	}
}

func TestLoggerDisabled(t *testing.T) {
	e, _ := NewEnforcer("examples/rbac_model.conf", "examples/rbac_policy.csv")
	
	logger := NewTestLogger(log.EventEnforce)
	logger.Enable(false)
	e.SetLogger(logger)
	
	// This should not be logged
	e.Enforce("alice", "data1", "read")
	
	events := logger.GetEvents()
	if len(events) != 0 {
		t.Errorf("Expected 0 events when logger is disabled, got %d", len(events))
	}
}
