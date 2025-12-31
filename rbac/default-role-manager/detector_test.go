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

package defaultrolemanager

import (
	"strings"
	"testing"

	"github.com/casbin/casbin/v3/detector"
)

// TestDetectorIntegration_ValidChain tests that valid inheritance chains do not trigger errors.
func TestDetectorIntegration_ValidChain(t *testing.T) {
	rm := NewRoleManagerImpl(10)
	d := detector.NewDefaultDetector()
	rm.SetDetector(d)

	// Add a valid chain: alice -> admin -> superuser
	err := rm.AddLink("alice", "admin")
	if err != nil {
		t.Errorf("Expected no error for valid link, got: %v", err)
	}

	err = rm.AddLink("admin", "superuser")
	if err != nil {
		t.Errorf("Expected no error for valid link, got: %v", err)
	}

	// Add another valid chain
	err = rm.AddLink("bob", "user")
	if err != nil {
		t.Errorf("Expected no error for valid link, got: %v", err)
	}

	// Verify the links work correctly
	testRole(t, rm, "alice", "admin", true)
	testRole(t, rm, "alice", "superuser", true)
	testRole(t, rm, "bob", "user", true)
}

// TestDetectorIntegration_SimpleCycle tests detection of a simple cycle: A -> B -> A.
func TestDetectorIntegration_SimpleCycle(t *testing.T) {
	rm := NewRoleManagerImpl(10)
	d := detector.NewDefaultDetector()
	rm.SetDetector(d)

	// Add first link: A -> B
	err := rm.AddLink("A", "B")
	if err != nil {
		t.Errorf("Expected no error for first link, got: %v", err)
	}

	// Try to create cycle: B -> A
	err = rm.AddLink("B", "A")
	if err == nil {
		t.Error("Expected cycle detection error, but got nil")
	} else {
		errMsg := err.Error()
		if !strings.Contains(errMsg, "cycle detected") {
			t.Errorf("Expected error message to contain 'cycle detected', got: %s", errMsg)
		}
	}

	// Verify that the second link was rolled back
	testRole(t, rm, "A", "B", true)
	testRole(t, rm, "B", "A", false) // Should be false because rollback removed this link
}

// TestDetectorIntegration_ComplexCycle tests detection of a complex cycle: A -> B -> C -> A.
func TestDetectorIntegration_ComplexCycle(t *testing.T) {
	rm := NewRoleManagerImpl(10)
	d := detector.NewDefaultDetector()
	rm.SetDetector(d)

	// Add first two links
	err := rm.AddLink("A", "B")
	if err != nil {
		t.Errorf("Expected no error for first link, got: %v", err)
	}

	err = rm.AddLink("B", "C")
	if err != nil {
		t.Errorf("Expected no error for second link, got: %v", err)
	}

	// Try to create cycle: C -> A
	err = rm.AddLink("C", "A")
	if err == nil {
		t.Error("Expected cycle detection error, but got nil")
	} else {
		errMsg := err.Error()
		if !strings.Contains(errMsg, "cycle detected") {
			t.Errorf("Expected error message to contain 'cycle detected', got: %s", errMsg)
		}
	}

	// Verify state after rollback
	testRole(t, rm, "A", "B", true)
	testRole(t, rm, "B", "C", true)
	testRole(t, rm, "C", "A", false) // Should be false because rollback removed this link
	testRole(t, rm, "A", "C", true)  // A should still reach C through B
}

// TestDetectorIntegration_SelfLoop tests detection of a self-loop: A -> A.
func TestDetectorIntegration_SelfLoop(t *testing.T) {
	rm := NewRoleManagerImpl(10)
	d := detector.NewDefaultDetector()
	rm.SetDetector(d)

	// Try to create a self-loop
	err := rm.AddLink("A", "A")
	if err == nil {
		t.Error("Expected cycle detection error for self-loop, but got nil")
	} else {
		errMsg := err.Error()
		if !strings.Contains(errMsg, "cycle detected") {
			t.Errorf("Expected error message to contain 'cycle detected', got: %s", errMsg)
		}
	}

	// Verify that the self-loop was not added
	roles, _ := rm.GetRoles("A")
	if len(roles) != 0 {
		t.Errorf("Expected no roles for A after rollback, got: %v", roles)
	}
}

// TestDetectorIntegration_RollbackVerification tests that rollback properly removes the illegal link.
func TestDetectorIntegration_RollbackVerification(t *testing.T) {
	rm := NewRoleManagerImpl(10)
	d := detector.NewDefaultDetector()
	rm.SetDetector(d)

	// Create a chain
	_ = rm.AddLink("u1", "g1")
	_ = rm.AddLink("g1", "g2")
	_ = rm.AddLink("g2", "g3")

	// Try to add a cycle
	err := rm.AddLink("g3", "u1")
	if err == nil {
		t.Error("Expected cycle detection error")
	}

	// Verify the state is consistent
	testRole(t, rm, "u1", "g1", true)
	testRole(t, rm, "g1", "g2", true)
	testRole(t, rm, "g2", "g3", true)
	testRole(t, rm, "g3", "u1", false) // Should be false after rollback

	// Verify that g3 has no outgoing links
	roles, _ := rm.GetRoles("g3")
	if len(roles) != 0 {
		t.Errorf("Expected g3 to have no roles after rollback, got: %v", roles)
	}
}

// TestDetectorIntegration_ComplexGraph tests a more complex graph scenario.
func TestDetectorIntegration_ComplexGraph(t *testing.T) {
	rm := NewRoleManagerImpl(10)
	d := detector.NewDefaultDetector()
	rm.SetDetector(d)

	// Build a valid complex graph
	_ = rm.AddLink("u1", "g1")
	_ = rm.AddLink("u2", "g1")
	_ = rm.AddLink("u3", "g2")
	_ = rm.AddLink("g1", "g3")
	_ = rm.AddLink("g2", "g3")

	// Try to add a cycle
	err := rm.AddLink("g3", "g1")
	if err == nil {
		t.Error("Expected cycle detection error")
	}

	// Verify the graph is still valid
	testRole(t, rm, "u1", "g1", true)
	testRole(t, rm, "u1", "g3", true)
	testRole(t, rm, "u2", "g3", true)
	testRole(t, rm, "u3", "g2", true)
	testRole(t, rm, "u3", "g3", true)
	testRole(t, rm, "g3", "g1", false) // Cycle was prevented
}

// TestDetectorIntegration_MultipleInheritance tests multiple inheritance without cycles.
func TestDetectorIntegration_MultipleInheritance(t *testing.T) {
	rm := NewRoleManagerImpl(10)
	d := detector.NewDefaultDetector()
	rm.SetDetector(d)

	// User inherits from multiple roles (diamond pattern)
	err := rm.AddLink("alice", "admin")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	err = rm.AddLink("alice", "moderator")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	err = rm.AddLink("admin", "superuser")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	err = rm.AddLink("moderator", "superuser")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify the relationships
	testRole(t, rm, "alice", "admin", true)
	testRole(t, rm, "alice", "moderator", true)
	testRole(t, rm, "alice", "superuser", true)
}

// TestDetectorIntegration_NoDetector tests that without a detector, cycles are allowed.
func TestDetectorIntegration_NoDetector(t *testing.T) {
	rm := NewRoleManagerImpl(10)
	// No detector is set

	// These operations should succeed even though they create a cycle
	err := rm.AddLink("A", "B")
	if err != nil {
		t.Errorf("Expected no error without detector, got: %v", err)
	}

	err = rm.AddLink("B", "A")
	if err != nil {
		t.Errorf("Expected no error without detector, got: %v", err)
	}

	// Both links should exist
	roles, _ := rm.GetRoles("A")
	if len(roles) != 1 || roles[0] != "B" {
		t.Errorf("Expected A to have role B, got: %v", roles)
	}
}

// TestDetectorIntegration_DisconnectedComponents tests multiple disconnected components.
func TestDetectorIntegration_DisconnectedComponents(t *testing.T) {
	rm := NewRoleManagerImpl(10)
	d := detector.NewDefaultDetector()
	rm.SetDetector(d)

	// Component 1: alice -> admin -> superuser
	_ = rm.AddLink("alice", "admin")
	_ = rm.AddLink("admin", "superuser")

	// Component 2: bob -> user
	_ = rm.AddLink("bob", "user")

	// Component 3: carol -> moderator
	_ = rm.AddLink("carol", "moderator")

	// All should succeed
	testRole(t, rm, "alice", "admin", true)
	testRole(t, rm, "alice", "superuser", true)
	testRole(t, rm, "bob", "user", true)
	testRole(t, rm, "carol", "moderator", true)

	// Try to create a cycle in one component
	err := rm.AddLink("superuser", "alice")
	if err == nil {
		t.Error("Expected cycle detection error")
	}

	// Verify other components are unaffected
	testRole(t, rm, "bob", "user", true)
	testRole(t, rm, "carol", "moderator", true)
}

// TestDetectorIntegration_LongCycle tests detection of a longer cycle.
func TestDetectorIntegration_LongCycle(t *testing.T) {
	rm := NewRoleManagerImpl(20)
	d := detector.NewDefaultDetector()
	rm.SetDetector(d)

	// Create a long chain: A -> B -> C -> D -> E
	_ = rm.AddLink("A", "B")
	_ = rm.AddLink("B", "C")
	_ = rm.AddLink("C", "D")
	_ = rm.AddLink("D", "E")

	// Try to close the cycle: E -> A
	err := rm.AddLink("E", "A")
	if err == nil {
		t.Error("Expected cycle detection error")
	}

	// Verify the chain is intact
	testRole(t, rm, "A", "B", true)
	testRole(t, rm, "A", "E", true)
	testRole(t, rm, "E", "A", false) // Cycle was prevented
}

// TestDetectorIntegration_AfterDeleteLink tests that detector works correctly after deleting links.
func TestDetectorIntegration_AfterDeleteLink(t *testing.T) {
	rm := NewRoleManagerImpl(10)
	d := detector.NewDefaultDetector()
	rm.SetDetector(d)

	// Create a valid chain
	_ = rm.AddLink("A", "B")
	_ = rm.AddLink("B", "C")

	// Delete a link
	_ = rm.DeleteLink("B", "C")

	// Now we should be able to add C -> B without creating a cycle
	err := rm.AddLink("C", "B")
	if err != nil {
		t.Errorf("Expected no error after deleting link, got: %v", err)
	}

	// But now A -> B -> ? and C -> B, so adding B -> A would create a cycle A -> B -> A
	// Let's test that
	_ = rm.DeleteLink("C", "B")
	err = rm.AddLink("B", "A")
	if err == nil {
		t.Error("Expected cycle detection error")
	}
}
