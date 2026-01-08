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

package detector

import (
	"fmt"
	"strings"
	"testing"

	defaultrolemanager "github.com/casbin/casbin/v3/rbac/default-role-manager"
)

func TestBFSDetector_NilRoleManager(t *testing.T) {
	detector := NewBFSDetector()
	err := detector.Check(nil)

	if err == nil {
		t.Error("Expected error for nil role manager, but got nil")
	} else {
		errMsg := err.Error()
		if !strings.Contains(errMsg, "role manager cannot be nil") {
			t.Errorf("Expected error message to contain 'role manager cannot be nil', got: %s", errMsg)
		}
	}
}

func TestBFSDetector_NoCycle(t *testing.T) {
	rm := defaultrolemanager.NewRoleManagerImpl(10)
	_ = rm.AddLink("alice", "admin")
	_ = rm.AddLink("bob", "user")
	_ = rm.AddLink("admin", "superuser")

	detector := NewBFSDetector()
	err := detector.Check(rm)

	if err != nil {
		t.Errorf("Expected no cycle, but got error: %v", err)
	}
}

func TestBFSDetector_SimpleCycle(t *testing.T) {
	rm := defaultrolemanager.NewRoleManagerImpl(10)
	_ = rm.AddLink("A", "B")
	_ = rm.AddLink("B", "A")

	detector := NewBFSDetector()
	err := detector.Check(rm)

	if err == nil {
		t.Error("Expected cycle detection error, but got nil")
	} else {
		errMsg := err.Error()
		if !strings.Contains(errMsg, "cycle detected") {
			t.Errorf("Expected error message to contain 'cycle detected', got: %s", errMsg)
		}
		// Should contain both A and B in the cycle
		if !strings.Contains(errMsg, "A") || !strings.Contains(errMsg, "B") {
			t.Errorf("Expected error message to contain both A and B, got: %s", errMsg)
		}
	}
}

func TestBFSDetector_ComplexCycle(t *testing.T) {
	rm := defaultrolemanager.NewRoleManagerImpl(10)
	_ = rm.AddLink("A", "B")
	_ = rm.AddLink("B", "C")
	_ = rm.AddLink("C", "A")

	detector := NewBFSDetector()
	err := detector.Check(rm)

	if err == nil {
		t.Error("Expected cycle detection error, but got nil")
	} else {
		errMsg := err.Error()
		if !strings.Contains(errMsg, "cycle detected") {
			t.Errorf("Expected error message to contain 'cycle detected', got: %s", errMsg)
		}
		// Should contain A, B, and C in the cycle
		if !strings.Contains(errMsg, "A") || !strings.Contains(errMsg, "B") || !strings.Contains(errMsg, "C") {
			t.Errorf("Expected error message to contain A, B, and C, got: %s", errMsg)
		}
	}
}

func TestBFSDetector_SelfLoop(t *testing.T) {
	rm := defaultrolemanager.NewRoleManagerImpl(10)
	_ = rm.AddLink("A", "A")

	detector := NewBFSDetector()
	err := detector.Check(rm)

	if err == nil {
		t.Error("Expected cycle detection error for self-loop, but got nil")
	} else {
		errMsg := err.Error()
		if !strings.Contains(errMsg, "cycle detected") {
			t.Errorf("Expected error message to contain 'cycle detected', got: %s", errMsg)
		}
	}
}

func TestBFSDetector_MultipleCycles(t *testing.T) {
	rm := defaultrolemanager.NewRoleManagerImpl(10)
	// First cycle: A -> B -> A
	_ = rm.AddLink("A", "B")
	_ = rm.AddLink("B", "A")
	// Second cycle: C -> D -> C
	_ = rm.AddLink("C", "D")
	_ = rm.AddLink("D", "C")

	detector := NewBFSDetector()
	err := detector.Check(rm)

	if err == nil {
		t.Error("Expected cycle detection error, but got nil")
	} else {
		errMsg := err.Error()
		if !strings.Contains(errMsg, "cycle detected") {
			t.Errorf("Expected error message to contain 'cycle detected', got: %s", errMsg)
		}
	}
}

func TestBFSDetector_DisconnectedComponents(t *testing.T) {
	rm := defaultrolemanager.NewRoleManagerImpl(10)
	// Component 1: alice -> admin -> superuser
	_ = rm.AddLink("alice", "admin")
	_ = rm.AddLink("admin", "superuser")
	// Component 2: bob -> user
	_ = rm.AddLink("bob", "user")
	// Component 3: carol -> moderator
	_ = rm.AddLink("carol", "moderator")

	detector := NewBFSDetector()
	err := detector.Check(rm)

	if err != nil {
		t.Errorf("Expected no cycle in disconnected components, but got error: %v", err)
	}
}

func TestBFSDetector_ComplexGraphWithCycle(t *testing.T) {
	rm := defaultrolemanager.NewRoleManagerImpl(10)
	// Build a complex graph with one cycle
	_ = rm.AddLink("u1", "g1")
	_ = rm.AddLink("u2", "g1")
	_ = rm.AddLink("g1", "g2")
	_ = rm.AddLink("g2", "g3")
	_ = rm.AddLink("g3", "g1") // Creates cycle: g1 -> g2 -> g3 -> g1
	_ = rm.AddLink("u3", "g4")

	detector := NewBFSDetector()
	err := detector.Check(rm)

	if err == nil {
		t.Error("Expected cycle detection error, but got nil")
	} else {
		errMsg := err.Error()
		if !strings.Contains(errMsg, "cycle detected") {
			t.Errorf("Expected error message to contain 'cycle detected', got: %s", errMsg)
		}
	}
}

func TestBFSDetector_LongCycle(t *testing.T) {
	rm := defaultrolemanager.NewRoleManagerImpl(20)
	// Create a long cycle: A -> B -> C -> D -> E -> A
	_ = rm.AddLink("A", "B")
	_ = rm.AddLink("B", "C")
	_ = rm.AddLink("C", "D")
	_ = rm.AddLink("D", "E")
	_ = rm.AddLink("E", "A")

	detector := NewBFSDetector()
	err := detector.Check(rm)

	if err == nil {
		t.Error("Expected cycle detection error, but got nil")
	} else {
		errMsg := err.Error()
		if !strings.Contains(errMsg, "cycle detected") {
			t.Errorf("Expected error message to contain 'cycle detected', got: %s", errMsg)
		}
	}
}

func TestBFSDetector_EmptyRoleManager(t *testing.T) {
	rm := defaultrolemanager.NewRoleManagerImpl(10)

	detector := NewBFSDetector()
	err := detector.Check(rm)

	if err != nil {
		t.Errorf("Expected no error for empty role manager, but got: %v", err)
	}
}

func TestBFSDetector_LargeGraphNoCycle(t *testing.T) {
	rm := defaultrolemanager.NewRoleManagerImpl(100)

	// Build a large graph with no cycles: a tree structure
	// Create 100 levels: u0 -> u1 -> u2 -> ... -> u99
	for i := 0; i < 99; i++ {
		user := fmt.Sprintf("u%d", i)
		role := fmt.Sprintf("u%d", i+1)
		_ = rm.AddLink(user, role)
	}

	detector := NewBFSDetector()
	err := detector.Check(rm)

	if err != nil {
		t.Errorf("Expected no cycle in large graph, but got error: %v", err)
	}
}

func TestBFSDetector_LargeGraphWithCycle(t *testing.T) {
	rm := defaultrolemanager.NewRoleManagerImpl(100)

	// Build a large graph with a cycle at the end
	// Create a chain: u0 -> u1 -> u2 -> ... -> u99 -> u0
	for i := 0; i < 99; i++ {
		user := fmt.Sprintf("u%d", i)
		role := fmt.Sprintf("u%d", i+1)
		_ = rm.AddLink(user, role)
	}
	// Add the cycle
	_ = rm.AddLink("u99", "u0")

	detector := NewBFSDetector()
	err := detector.Check(rm)

	if err == nil {
		t.Error("Expected cycle detection error in large graph, but got nil")
	} else {
		errMsg := err.Error()
		if !strings.Contains(errMsg, "cycle detected") {
			t.Errorf("Expected error message to contain 'cycle detected', got: %s", errMsg)
		}
	}
}

// Performance test with 10,000 roles.
func TestBFSDetector_PerformanceLargeGraph(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	// Use a higher maxHierarchyLevel to support deep hierarchies
	rm := defaultrolemanager.NewRoleManagerImpl(10000)

	// Build a large tree structure with 10,000 roles
	// Each role has up to 3 children
	numRoles := 10000
	for i := 0; i < numRoles; i++ {
		role := fmt.Sprintf("r%d", i)
		// Add links to create a tree structure
		child1 := (i * 3) + 1
		child2 := (i * 3) + 2
		child3 := (i * 3) + 3
		if child1 < numRoles {
			_ = rm.AddLink(fmt.Sprintf("r%d", child1), role)
		}
		if child2 < numRoles {
			_ = rm.AddLink(fmt.Sprintf("r%d", child2), role)
		}
		if child3 < numRoles {
			_ = rm.AddLink(fmt.Sprintf("r%d", child3), role)
		}
	}

	detector := NewBFSDetector()
	err := detector.Check(rm)

	if err != nil {
		t.Errorf("Expected no cycle in large performance test, but got error: %v", err)
	}
}

func TestBFSDetector_MultipleInheritance(t *testing.T) {
	rm := defaultrolemanager.NewRoleManagerImpl(10)
	// User inherits from multiple roles
	_ = rm.AddLink("alice", "admin")
	_ = rm.AddLink("alice", "moderator")
	_ = rm.AddLink("admin", "superuser")
	_ = rm.AddLink("moderator", "user")

	detector := NewBFSDetector()
	err := detector.Check(rm)

	if err != nil {
		t.Errorf("Expected no cycle with multiple inheritance, but got error: %v", err)
	}
}

func TestBFSDetector_DiamondPattern(t *testing.T) {
	rm := defaultrolemanager.NewRoleManagerImpl(10)
	// Diamond pattern: alice -> admin, alice -> moderator, admin -> superuser, moderator -> superuser
	_ = rm.AddLink("alice", "admin")
	_ = rm.AddLink("alice", "moderator")
	_ = rm.AddLink("admin", "superuser")
	_ = rm.AddLink("moderator", "superuser")

	detector := NewBFSDetector()
	err := detector.Check(rm)

	if err != nil {
		t.Errorf("Expected no cycle in diamond pattern, but got error: %v", err)
	}
}

func TestBFSDetector_DiamondPatternWithCycle(t *testing.T) {
	rm := defaultrolemanager.NewRoleManagerImpl(10)
	// Diamond pattern with cycle: alice -> admin, alice -> moderator, admin -> superuser, moderator -> superuser, superuser -> alice
	_ = rm.AddLink("alice", "admin")
	_ = rm.AddLink("alice", "moderator")
	_ = rm.AddLink("admin", "superuser")
	_ = rm.AddLink("moderator", "superuser")
	_ = rm.AddLink("superuser", "alice") // Creates cycle

	detector := NewBFSDetector()
	err := detector.Check(rm)

	if err == nil {
		t.Error("Expected cycle detection error in diamond pattern with cycle, but got nil")
	} else {
		errMsg := err.Error()
		if !strings.Contains(errMsg, "cycle detected") {
			t.Errorf("Expected error message to contain 'cycle detected', got: %s", errMsg)
		}
	}
}
