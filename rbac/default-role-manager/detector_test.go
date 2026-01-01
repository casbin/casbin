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

package defaultrolemanager_test

import (
	"strings"
	"testing"

	"github.com/casbin/casbin/v3"
)

// TestEnforcer_AutoCycleDetection tests that cycle detection is automatically enabled
// for the default enforcer and prevents cyclic role hierarchies.
func TestEnforcer_AutoCycleDetection(t *testing.T) {
	// Initialize a standard Enforcer
	e, err := casbin.NewEnforcer("../../examples/rbac_model.conf", "")
	if err != nil {
		t.Fatalf("Failed to create enforcer: %v", err)
	}

	// Add a valid policy: alice -> bob
	ok, err := e.AddGroupingPolicy("alice", "bob")
	if err != nil {
		t.Fatalf("Failed to add first grouping policy: %v", err)
	}
	if !ok {
		t.Fatalf("Expected first policy to be added successfully")
	}

	// Verify the policy was added
	hasPolicy, err := e.HasGroupingPolicy("alice", "bob")
	if err != nil {
		t.Fatalf("Failed to check policy: %v", err)
	}
	if !hasPolicy {
		t.Fatalf("Expected policy 'alice -> bob' to exist")
	}

	// Add a cyclic policy: bob -> alice (this should fail)
	ok, err = e.AddGroupingPolicy("bob", "alice")
	if err == nil {
		t.Fatalf("Expected error when adding cyclic policy, but got none")
	}
	if !strings.Contains(err.Error(), "cycle detected") {
		t.Fatalf("Expected 'cycle detected' error, got: %v", err)
	}
	if ok {
		t.Fatalf("Expected AddGroupingPolicy to return false for cyclic policy")
	}

	// Verify the cyclic policy was rolled back (not added)
	hasPolicy, err = e.HasGroupingPolicy("bob", "alice")
	if err != nil {
		t.Fatalf("Failed to check cyclic policy: %v", err)
	}
	if hasPolicy {
		t.Fatalf("Expected cyclic policy 'bob -> alice' to NOT exist (should be rolled back)")
	}

	// Verify the original policy still exists
	hasPolicy, err = e.HasGroupingPolicy("alice", "bob")
	if err != nil {
		t.Fatalf("Failed to check original policy: %v", err)
	}
	if !hasPolicy {
		t.Fatalf("Expected original policy 'alice -> bob' to still exist")
	}
}

// TestEnforcer_DetectAPI tests the public Detect() API method.
func TestEnforcer_DetectAPI(t *testing.T) {
	// Initialize a new Enforcer with a valid hierarchy (A -> B -> C)
	e, err := casbin.NewEnforcer("../../examples/rbac_model.conf", "")
	if err != nil {
		t.Fatalf("Failed to create enforcer: %v", err)
	}

	// Add valid policies: A -> B -> C
	_, err = e.AddGroupingPolicy("A", "B")
	if err != nil {
		t.Fatalf("Failed to add policy A -> B: %v", err)
	}
	_, err = e.AddGroupingPolicy("B", "C")
	if err != nil {
		t.Fatalf("Failed to add policy B -> C: %v", err)
	}

	// Call Detect() and assert it returns nil (no cycle)
	err = e.Detect()
	if err != nil {
		t.Fatalf("Expected no cycle in valid hierarchy, but got error: %v", err)
	}
}

// TestEnforcer_ComplexCycleDetection tests more complex cycle scenarios.
func TestEnforcer_ComplexCycleDetection(t *testing.T) {
	e, err := casbin.NewEnforcer("../../examples/rbac_model.conf", "")
	if err != nil {
		t.Fatalf("Failed to create enforcer: %v", err)
	}

	// Build a valid hierarchy: alice -> bob -> charlie
	_, err = e.AddGroupingPolicy("alice", "bob")
	if err != nil {
		t.Fatalf("Failed to add alice -> bob: %v", err)
	}
	_, err = e.AddGroupingPolicy("bob", "charlie")
	if err != nil {
		t.Fatalf("Failed to add bob -> charlie: %v", err)
	}

	// Try to create a cycle: charlie -> alice
	ok, err := e.AddGroupingPolicy("charlie", "alice")
	if err == nil {
		t.Fatalf("Expected error when creating cycle, but got none")
	}
	if !strings.Contains(err.Error(), "cycle detected") {
		t.Fatalf("Expected 'cycle detected' error, got: %v", err)
	}
	if ok {
		t.Fatalf("Expected AddGroupingPolicy to return false for cyclic policy")
	}

	// Verify all valid policies still exist
	hasPolicy, _ := e.HasGroupingPolicy("alice", "bob")
	if !hasPolicy {
		t.Fatalf("Expected policy 'alice -> bob' to still exist")
	}
	hasPolicy, _ = e.HasGroupingPolicy("bob", "charlie")
	if !hasPolicy {
		t.Fatalf("Expected policy 'bob -> charlie' to still exist")
	}

	// Verify the cyclic policy was not added
	hasPolicy, _ = e.HasGroupingPolicy("charlie", "alice")
	if hasPolicy {
		t.Fatalf("Expected cyclic policy 'charlie -> alice' to NOT exist")
	}
}

// TestEnforcer_SelfLoopDetection tests detection of self-loops.
func TestEnforcer_SelfLoopDetection(t *testing.T) {
	e, err := casbin.NewEnforcer("../../examples/rbac_model.conf", "")
	if err != nil {
		t.Fatalf("Failed to create enforcer: %v", err)
	}

	// Try to create a self-loop: alice -> alice
	ok, err := e.AddGroupingPolicy("alice", "alice")
	if err == nil {
		t.Fatalf("Expected error when creating self-loop, but got none")
	}
	if !strings.Contains(err.Error(), "cycle detected") {
		t.Fatalf("Expected 'cycle detected' error, got: %v", err)
	}
	if ok {
		t.Fatalf("Expected AddGroupingPolicy to return false for self-loop")
	}

	// Verify the self-loop was not added
	hasPolicy, _ := e.HasGroupingPolicy("alice", "alice")
	if hasPolicy {
		t.Fatalf("Expected self-loop 'alice -> alice' to NOT exist")
	}
}
