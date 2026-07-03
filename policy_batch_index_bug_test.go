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
	"testing"

	"github.com/casbin/casbin/v3/model"
)

// Test for issue #1733: Index corruption in batch delete
// When removing multiple policies, the PolicyMap indices become stale after the first removal
func TestBatchRemoveIndexCorruption(t *testing.T) {
	modelText := `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = r.sub == p.sub && r.obj == p.obj && r.act == p.act
`

	m, err := model.NewModelFromString(modelText)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	e, err := NewEnforcer(m)
	if err != nil {
		t.Fatalf("Failed to create enforcer: %v", err)
	}

	// Add multiple policies with commas in fields
	// The order matters for reproducing the bug
	policies := [][]string{
		{"alice", "data1", "read"},                     // index 0
		{"bob", "data,with,comma", "write"},            // index 1 - will be deleted
		{"charlie", "data3", "read"},                   // index 2
		{"dave", "another,comma,field", "write"},       // index 3 - will be deleted
		{"eve", "data5", "read"},                       // index 4
		{"frank", "yet,another,comma", "write"},        // index 5 - will be deleted
		{"grace", "data7", "read"},                     // index 6
	}

	for _, p := range policies {
		_, err := e.AddPolicy(p)
		if err != nil {
			t.Fatalf("Failed to add policy: %v", err)
		}
	}

	allPolicies, err := e.GetPolicy()
	if err != nil {
		t.Fatalf("Failed to get policies: %v", err)
	}
	t.Logf("Added %d policies", len(allPolicies))

	// Now try to remove policies at indices 1, 3, 5 in a single batch
	// This simulates the issue where multiple non-consecutive policies with commas need to be removed
	toRemove := [][]string{
		{"bob", "data,with,comma", "write"},
		{"dave", "another,comma,field", "write"},
		{"frank", "yet,another,comma", "write"},
	}

	t.Log("Attempting to remove 3 policies with commas in batch...")
	
	// Verify they all exist before removal
	for i, p := range toRemove {
		has, err := e.HasPolicy(p)
		if err != nil {
			t.Fatalf("HasPolicy failed: %v", err)
		}
		if !has {
			t.Errorf("Policy %d not found before removal: %v", i, p)
		}
	}

	removed, err := e.RemovePolicies(toRemove)
	if err != nil {
		t.Fatalf("RemovePolicies failed: %v", err)
	}

	if !removed {
		t.Error("RemovePolicies returned false")
	}

	remaining, err := e.GetPolicy()
	if err != nil {
		t.Fatalf("Failed to get remaining policies: %v", err)
	}

	t.Logf("After removal: %d policies remaining (expected 4)", len(remaining))

	if len(remaining) != 4 {
		t.Errorf("Expected 4 policies remaining, got %d", len(remaining))
		t.Log("Remaining policies:")
		for _, p := range remaining {
			t.Logf("  %v", p)
		}
	}

	// Verify each removed policy is actually gone
	for i, p := range toRemove {
		has, err := e.HasPolicy(p)
		if err != nil {
			t.Fatalf("HasPolicy check failed: %v", err)
		}
		if has {
			t.Errorf("Policy %d still exists after batch removal: %v", i, p)
		}
	}

	// Verify the policies we expected to keep are still there
	expectedRemaining := [][]string{
		{"alice", "data1", "read"},
		{"charlie", "data3", "read"},
		{"eve", "data5", "read"},
		{"grace", "data7", "read"},
	}

	for _, p := range expectedRemaining {
		has, err := e.HasPolicy(p)
		if err != nil {
			t.Fatalf("HasPolicy check failed: %v", err)
		}
		if !has {
			t.Errorf("Expected policy missing: %v", p)
		}
	}
}

// Simpler test to demonstrate the index shifting bug
func TestBatchRemoveAdjacentPolicies(t *testing.T) {
	modelText := `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = r.sub == p.sub && r.obj == p.obj && r.act == p.act
`

	m, err := model.NewModelFromString(modelText)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	e, err := NewEnforcer(m)
	if err != nil {
		t.Fatalf("Failed to create enforcer: %v", err)
	}

	// Add 5 policies
	policies := [][]string{
		{"user1", "data1", "read"},
		{"user2", "data2", "read"},
		{"user3", "data3", "read"},
		{"user4", "data4", "read"},
		{"user5", "data5", "read"},
	}

	for _, p := range policies {
		_, err := e.AddPolicy(p)
		if err != nil {
			t.Fatalf("Failed to add policy: %v", err)
		}
	}

	// Try to remove policies at indices 1, 2, 3 in one batch
	// After removing index 1, what was at index 2 is now at index 1
	// After removing (new) index 1, what was at index 3 is now at index 1
	toRemove := [][]string{
		policies[1],
		policies[2],
		policies[3],
	}

	removed, err := e.RemovePolicies(toRemove)
	if err != nil {
		t.Fatalf("RemovePolicies failed: %v", err)
	}

	if !removed {
		t.Error("RemovePolicies returned false")
	}

	remaining, err := e.GetPolicy()
	if err != nil {
		t.Fatalf("Failed to get remaining policies: %v", err)
	}

	if len(remaining) != 2 {
		t.Errorf("Expected 2 policies remaining, got %d", len(remaining))
		t.Log("Remaining:")
		for _, p := range remaining {
			t.Logf("  %v", p)
		}
	}
}
