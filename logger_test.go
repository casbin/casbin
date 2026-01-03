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
	"bytes"
	"strings"
	"testing"

	"github.com/casbin/casbin/v3/log"
)

func TestEnforcerWithDefaultLogger(t *testing.T) {
	// Create enforcer with RBAC model and policy
	e, err := NewEnforcer("examples/rbac_model.conf", "examples/rbac_policy.csv")
	if err != nil {
		t.Fatalf("Failed to create enforcer: %v", err)
	}

	// Create a buffer to capture log output
	var buf bytes.Buffer
	logger := log.NewDefaultLogger()
	logger.SetOutput(&buf)

	// Set up a callback to track log entries
	var callbackEntries []*log.LogEntry
	err = logger.SetLogCallback(func(entry *log.LogEntry) error {
		// Create a copy of the entry to store
		entryCopy := *entry
		callbackEntries = append(callbackEntries, &entryCopy)
		return nil
	})
	if err != nil {
		t.Fatalf("Failed to set log callback: %v", err)
	}

	// Set the logger on the enforcer
	e.SetLogger(logger)

	// Test Enforce events
	result, err := e.Enforce("alice", "data1", "read")
	if err != nil {
		t.Fatalf("Enforce failed: %v", err)
	}
	if !result {
		t.Errorf("Expected alice to have read access to data1")
	}

	result, err = e.Enforce("bob", "data2", "write")
	if err != nil {
		t.Fatalf("Enforce failed: %v", err)
	}
	if !result {
		t.Errorf("Expected bob to have write access to data2")
	}

	// Test AddPolicy event
	added, err := e.AddPolicy("charlie", "data3", "read")
	if err != nil {
		t.Fatalf("AddPolicy failed: %v", err)
	}
	if !added {
		t.Errorf("Expected policy to be added")
	}

	// Test RemovePolicy event
	removed, err := e.RemovePolicy("charlie", "data3", "read")
	if err != nil {
		t.Fatalf("RemovePolicy failed: %v", err)
	}
	if !removed {
		t.Errorf("Expected policy to be removed")
	}

	// Test SavePolicy event
	err = e.SavePolicy()
	if err != nil {
		t.Fatalf("SavePolicy failed: %v", err)
	}

	// Test LoadPolicy event
	err = e.LoadPolicy()
	if err != nil {
		t.Fatalf("LoadPolicy failed: %v", err)
	}

	// Verify buffer output
	logOutput := buf.String()
	if !strings.Contains(logOutput, "[enforce]") {
		t.Errorf("Expected log output to contain enforce events")
	}
	if !strings.Contains(logOutput, "[addPolicy]") {
		t.Errorf("Expected log output to contain addPolicy event")
	}
	if !strings.Contains(logOutput, "[removePolicy]") {
		t.Errorf("Expected log output to contain removePolicy event")
	}
	if !strings.Contains(logOutput, "[savePolicy]") {
		t.Errorf("Expected log output to contain savePolicy event")
	}
	if !strings.Contains(logOutput, "[loadPolicy]") {
		t.Errorf("Expected log output to contain loadPolicy event")
	}

	// Verify callback was called
	if len(callbackEntries) == 0 {
		t.Fatalf("Expected callback to be called, but got no entries")
	}

	// Verify callback entries contain the expected event types
	foundEnforce := false
	foundAddPolicy := false
	foundRemovePolicy := false
	foundSavePolicy := false
	foundLoadPolicy := false

	for _, entry := range callbackEntries {
		switch entry.EventType {
		case log.EventEnforce:
			foundEnforce = true
			if entry.Subject == "" && entry.Object == "" && entry.Action == "" {
				t.Errorf("Expected enforce entry to have subject, object, and action")
			}
		case log.EventAddPolicy:
			foundAddPolicy = true
			if entry.RuleCount != 1 {
				t.Errorf("Expected addPolicy entry to have RuleCount=1, got %d", entry.RuleCount)
			}
		case log.EventRemovePolicy:
			foundRemovePolicy = true
			if entry.RuleCount != 1 {
				t.Errorf("Expected removePolicy entry to have RuleCount=1, got %d", entry.RuleCount)
			}
		case log.EventSavePolicy:
			foundSavePolicy = true
			if entry.RuleCount == 0 {
				t.Errorf("Expected savePolicy entry to have RuleCount>0")
			}
		case log.EventLoadPolicy:
			foundLoadPolicy = true
			if entry.RuleCount == 0 {
				t.Errorf("Expected loadPolicy entry to have RuleCount>0")
			}
		}
	}

	if !foundEnforce {
		t.Errorf("Expected to find EventEnforce in callback entries")
	}
	if !foundAddPolicy {
		t.Errorf("Expected to find EventAddPolicy in callback entries")
	}
	if !foundRemovePolicy {
		t.Errorf("Expected to find EventRemovePolicy in callback entries")
	}
	if !foundSavePolicy {
		t.Errorf("Expected to find EventSavePolicy in callback entries")
	}
	if !foundLoadPolicy {
		t.Errorf("Expected to find EventLoadPolicy in callback entries")
	}
}

func TestSetEventTypes(t *testing.T) {
	// Create enforcer with RBAC model and policy
	e, err := NewEnforcer("examples/rbac_model.conf", "examples/rbac_policy.csv")
	if err != nil {
		t.Fatalf("Failed to create enforcer: %v", err)
	}

	// Create a buffer to capture log output
	var buf bytes.Buffer
	logger := log.NewDefaultLogger()
	logger.SetOutput(&buf)

	// Set up a callback to track log entries
	var callbackEntries []*log.LogEntry
	err = logger.SetLogCallback(func(entry *log.LogEntry) error {
		// Create a copy of the entry to store
		entryCopy := *entry
		callbackEntries = append(callbackEntries, &entryCopy)
		return nil
	})
	if err != nil {
		t.Fatalf("Failed to set log callback: %v", err)
	}

	// Configure logger to only log EventEnforce and EventAddPolicy
	err = logger.SetEventTypes([]log.EventType{log.EventEnforce, log.EventAddPolicy})
	if err != nil {
		t.Fatalf("Failed to set event types: %v", err)
	}

	// Set the logger on the enforcer
	e.SetLogger(logger)

	// Perform various operations
	_, err = e.Enforce("alice", "data1", "read")
	if err != nil {
		t.Fatalf("Enforce failed: %v", err)
	}

	_, err = e.AddPolicy("charlie", "data3", "read")
	if err != nil {
		t.Fatalf("AddPolicy failed: %v", err)
	}

	_, err = e.RemovePolicy("charlie", "data3", "read")
	if err != nil {
		t.Fatalf("RemovePolicy failed: %v", err)
	}

	err = e.LoadPolicy()
	if err != nil {
		t.Fatalf("LoadPolicy failed: %v", err)
	}

	// Verify buffer output only contains EventEnforce and EventAddPolicy
	logOutput := buf.String()
	if !strings.Contains(logOutput, "[enforce]") {
		t.Errorf("Expected log output to contain enforce events")
	}
	if !strings.Contains(logOutput, "[addPolicy]") {
		t.Errorf("Expected log output to contain addPolicy event")
	}
	if strings.Contains(logOutput, "[removePolicy]") {
		t.Errorf("Did not expect log output to contain removePolicy event")
	}
	if strings.Contains(logOutput, "[loadPolicy]") {
		t.Errorf("Did not expect log output to contain loadPolicy event")
	}

	// Verify callback entries
	foundEnforce := false
	foundAddPolicy := false
	foundRemovePolicy := false
	foundLoadPolicy := false

	for _, entry := range callbackEntries {
		// All entries should be called back regardless of IsActive
		switch entry.EventType {
		case log.EventEnforce:
			foundEnforce = true
			if !entry.IsActive {
				t.Errorf("Expected enforce entry to be active")
			}
		case log.EventAddPolicy:
			foundAddPolicy = true
			if !entry.IsActive {
				t.Errorf("Expected addPolicy entry to be active")
			}
		case log.EventRemovePolicy:
			foundRemovePolicy = true
			if entry.IsActive {
				t.Errorf("Expected removePolicy entry to be inactive")
			}
		case log.EventLoadPolicy:
			foundLoadPolicy = true
			if entry.IsActive {
				t.Errorf("Expected loadPolicy entry to be inactive")
			}
		}
	}

	if !foundEnforce {
		t.Errorf("Expected to find EventEnforce in callback entries")
	}
	if !foundAddPolicy {
		t.Errorf("Expected to find EventAddPolicy in callback entries")
	}
	if !foundRemovePolicy {
		t.Errorf("Expected to find EventRemovePolicy in callback entries")
	}
	if !foundLoadPolicy {
		t.Errorf("Expected to find EventLoadPolicy in callback entries")
	}

	// Verify that only active events were logged to buffer
	activeCount := 0
	for _, entry := range callbackEntries {
		if entry.IsActive {
			activeCount++
		}
	}

	// Count lines in buffer output (rough approximation)
	lines := strings.Split(strings.TrimSpace(logOutput), "\n")
	nonEmptyLines := 0
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			nonEmptyLines++
		}
	}

	// We should have log output for active events
	if nonEmptyLines == 0 {
		t.Errorf("Expected some log output for active events")
	}
}
