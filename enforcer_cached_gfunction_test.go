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

import "testing"

// TestCachedGFunctionAfterAddingGroupingPolicy tests that adding new grouping policies
// at runtime correctly updates permission evaluation without requiring a restart or manual cache clearing.
// This is a regression test for the issue where GenerateGFunction's memoization cache
// caused stale permission results after adding new grouping policies.
func TestCachedGFunctionAfterAddingGroupingPolicy(t *testing.T) {
	e, err := NewEnforcer("examples/rbac_model.conf", "examples/rbac_policy.csv")
	if err != nil {
		t.Fatalf("Failed to create enforcer: %v", err)
	}

	// Initial state: bob should not have access to data2 (read)
	// bob has no roles initially
	ok, err := e.Enforce("bob", "data2", "read")
	if err != nil {
		t.Fatalf("Enforce failed: %v", err)
	}
	if ok {
		t.Error("bob should not have read access to data2 initially")
	}

	// Add a new grouping policy: bob becomes data2_admin
	// data2_admin has read and write access to data2
	_, err = e.AddGroupingPolicy("bob", "data2_admin")
	if err != nil {
		t.Fatalf("Failed to add grouping policy: %v", err)
	}

	// Now bob should have read access to data2 through the data2_admin role
	// This should work immediately without needing to clear cache or restart
	ok, err = e.Enforce("bob", "data2", "read")
	if err != nil {
		t.Fatalf("Enforce failed after adding grouping policy: %v", err)
	}
	if !ok {
		t.Error("bob should have read access to data2 after being added to data2_admin role")
	}

	// Also verify write access works
	ok, err = e.Enforce("bob", "data2", "write")
	if err != nil {
		t.Fatalf("Enforce failed: %v", err)
	}
	if !ok {
		t.Error("bob should have write access to data2 after being added to data2_admin role")
	}
}

// TestCachedGFunctionWithMultipleEnforceCalls tests that the g() function cache
// properly invalidates when grouping policies change, even after multiple enforce calls.
func TestCachedGFunctionWithMultipleEnforceCalls(t *testing.T) {
	e, err := NewEnforcer("examples/rbac_model.conf", "examples/rbac_policy.csv")
	if err != nil {
		t.Fatalf("Failed to create enforcer: %v", err)
	}

	// Make multiple enforce calls to ensure the g() function closure is cached
	for i := 0; i < 5; i++ {
		ok, enforceErr := e.Enforce("charlie", "data2", "read")
		if enforceErr != nil {
			t.Fatalf("Enforce failed on iteration %d: %v", i, enforceErr)
		}
		if ok {
			t.Errorf("charlie should not have read access to data2 on iteration %d", i)
		}
	}

	// Add grouping policy
	_, err = e.AddGroupingPolicy("charlie", "data2_admin")
	if err != nil {
		t.Fatalf("Failed to add grouping policy: %v", err)
	}

	// Immediately verify the change took effect
	ok, err := e.Enforce("charlie", "data2", "read")
	if err != nil {
		t.Fatalf("Enforce failed after adding grouping policy: %v", err)
	}
	if !ok {
		t.Error("charlie should have read access to data2 immediately after being added to data2_admin role")
	}

	// Make multiple calls to ensure it stays consistent
	for i := 0; i < 5; i++ {
		ok, enforceErr := e.Enforce("charlie", "data2", "read")
		if enforceErr != nil {
			t.Fatalf("Enforce failed on iteration %d after policy change: %v", i, enforceErr)
		}
		if !ok {
			t.Errorf("charlie should have read access to data2 on iteration %d after policy change", i)
		}
	}
}

// TestCachedGFunctionAfterRemovingGroupingPolicy tests that removing grouping policies
// also properly invalidates the g() function cache.
func TestCachedGFunctionAfterRemovingGroupingPolicy(t *testing.T) {
	e, err := NewEnforcer("examples/rbac_model.conf", "examples/rbac_policy.csv")
	if err != nil {
		t.Fatalf("Failed to create enforcer: %v", err)
	}

	// alice initially has the data2_admin role
	ok, err := e.Enforce("alice", "data2", "read")
	if err != nil {
		t.Fatalf("Enforce failed: %v", err)
	}
	if !ok {
		t.Error("alice should have read access to data2 initially")
	}

	// Remove alice from data2_admin role
	_, err = e.RemoveGroupingPolicy("alice", "data2_admin")
	if err != nil {
		t.Fatalf("Failed to remove grouping policy: %v", err)
	}

	// Now alice should not have access to data2 (she only has access to data1)
	ok, err = e.Enforce("alice", "data2", "read")
	if err != nil {
		t.Fatalf("Enforce failed after removing grouping policy: %v", err)
	}
	if ok {
		t.Error("alice should not have read access to data2 after being removed from data2_admin role")
	}

	// Verify alice still has access to data1 (direct policy)
	ok, err = e.Enforce("alice", "data1", "read")
	if err != nil {
		t.Fatalf("Enforce failed: %v", err)
	}
	if !ok {
		t.Error("alice should still have read access to data1 (direct policy)")
	}
}

// TestCachedGFunctionAfterBuildRoleLinks tests the specific scenario mentioned in the bug report:
// adding grouping policies and calling BuildRoleLinks() manually should properly invalidate the cache.
func TestCachedGFunctionAfterBuildRoleLinks(t *testing.T) {
	e, err := NewEnforcer("examples/rbac_model.conf", "examples/rbac_policy.csv")
	if err != nil {
		t.Fatalf("Failed to create enforcer: %v", err)
	}

	// First, make some enforce calls to ensure the g() function closure is created and cached
	// This will cache "bob" NOT having data2_admin role in the g() function's sync.Map
	for i := 0; i < 3; i++ {
		ok, enforceErr := e.Enforce("bob", "data2", "read")
		if enforceErr != nil {
			t.Fatalf("Enforce failed on iteration %d: %v", i, enforceErr)
		}
		if ok {
			t.Errorf("bob should not have read access to data2 on iteration %d (before adding role)", i)
		}
	}

	// Disable autoBuildRoleLinks to manually control when role links are rebuilt
	e.EnableAutoBuildRoleLinks(false)

	// Manually add the grouping policy to the model (bypassing BuildIncrementalRoleLinks)
	// This simulates the scenario where policies are loaded from database
	err = e.model.AddPolicy("g", "g", []string{"bob", "data2_admin"})
	if err != nil {
		t.Fatalf("Failed to add grouping policy to model: %v", err)
	}

	// Manually build role links as mentioned in the issue
	// This is the key part - BuildRoleLinks() should invalidate the matcher map cache
	err = e.BuildRoleLinks()
	if err != nil {
		t.Fatalf("Failed to build role links: %v", err)
	}

	// Now bob should have read access to data2 through the data2_admin role
	// This is where the bug would manifest - if BuildRoleLinks() doesn't invalidate the cache,
	// the old g() function closure with "bob->data2_admin = false" cached will still be used
	ok, err := e.Enforce("bob", "data2", "read")
	if err != nil {
		t.Fatalf("Enforce failed after BuildRoleLinks: %v", err)
	}
	if !ok {
		t.Error("bob should have read access to data2 after BuildRoleLinks() - this indicates the g() function cache was not properly invalidated")
	}
}
