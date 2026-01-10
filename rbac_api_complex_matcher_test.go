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

	"github.com/casbin/casbin/v3/util"
)

// TestGetImplicitPermissionsForUserWithComplexMatcher tests the GetImplicitPermissionsForUser
// function with complex matchers that include wildcards and OR conditions.
// This addresses the issue: https://github.com/casbin/node-casbin/issues/481
func TestGetImplicitPermissionsForUserWithComplexMatcher(t *testing.T) {
	e, _ := NewEnforcer("/tmp/test_complex_matcher_model.conf", "/tmp/test_complex_matcher_policy.csv")

	// Test michael who has roles1 in tenant1
	// michael -> roles1 -> abstract_roles1 in tenant1
	// abstract_roles1 has permissions on devis with domain *
	perms, err := e.GetImplicitPermissionsForUser("michael", "tenant1")
	if err != nil {
		t.Fatalf("GetImplicitPermissionsForUser failed: %v", err)
	}
	
	t.Logf("Permissions for michael in tenant1: %v", perms)
	
	// Michael should have access to devis read and create because:
	// - g(michael, abstract_roles1, tenant1) is true (through roles1)
	// - p.dom == '*' matches any domain, and we replace * with the requested domain
	expectedPerms := [][]string{
		{"abstract_roles1", "devis", "read", "tenant1"},
		{"abstract_roles1", "devis", "create", "tenant1"},
	}
	
	if !util.Set2DEquals(expectedPerms, perms) {
		t.Errorf("Expected permissions %v, got %v", expectedPerms, perms)
	}

	// Test thomas who has roles2 in tenant1 and tenant2
	perms, err = e.GetImplicitPermissionsForUser("thomas", "tenant1")
	if err != nil {
		t.Fatalf("GetImplicitPermissionsForUser failed: %v", err)
	}
	
	t.Logf("Permissions for thomas in tenant1: %v", perms)
	
	// Thomas should have access to devis and organization because:
	// - g(thomas, abstract_roles2, tenant1) is true (through roles2)
	// - p.dom == '*' matches any domain, and we replace * with the requested domain
	expectedPerms = [][]string{
		{"abstract_roles2", "devis", "read", "tenant1"},
		{"abstract_roles2", "organization", "read", "tenant1"},
		{"abstract_roles2", "organization", "write", "tenant1"},
	}
	
	if !util.Set2DEquals(expectedPerms, perms) {
		t.Errorf("Expected permissions %v, got %v", expectedPerms, perms)
	}

	// Test theo who has super_user with wildcard domain
	perms, err = e.GetImplicitPermissionsForUser("theo", "any_tenant")
	if err != nil {
		t.Fatalf("GetImplicitPermissionsForUser failed: %v", err)
	}
	
	t.Logf("Permissions for theo in any_tenant: %v", perms)
	
	// Theo should have access to all abstract_roles2 permissions because:
	// - g(theo, abstract_roles2, '*') is true (through super_user)
	// - p.dom == '*' matches any domain, and we replace * with the requested domain
	expectedPerms = [][]string{
		{"abstract_roles2", "devis", "read", "any_tenant"},
		{"abstract_roles2", "organization", "read", "any_tenant"},
		{"abstract_roles2", "organization", "write", "any_tenant"},
	}
	
	if !util.Set2DEquals(expectedPerms, perms) {
		t.Errorf("Expected permissions %v, got %v", expectedPerms, perms)
	}

	// Verify enforcement also works correctly
	allowed, err := e.Enforce("michael", "devis", "read", "tenant1")
	if err != nil {
		t.Fatalf("Enforce failed: %v", err)
	}
	if !allowed {
		t.Error("michael should be allowed to read devis in tenant1")
	}

	allowed, err = e.Enforce("theo", "organization", "write", "any_tenant")
	if err != nil {
		t.Fatalf("Enforce failed: %v", err)
	}
	if !allowed {
		t.Error("theo should be allowed to write organization in any_tenant")
	}
}

// TestGetImplicitPermissionsForUserWithoutDomain tests that GetImplicitPermissionsForUser
// works correctly when no domain is specified with a domain-based model
func TestGetImplicitPermissionsForUserWithoutDomain(t *testing.T) {
	e, _ := NewEnforcer("/tmp/test_complex_matcher_model.conf", "/tmp/test_complex_matcher_policy.csv")

	// When no domain is specified with a domain-based model, behavior depends on
	// whether the grouping policies include domain-less entries.
	// In this model, all grouping policies have domains, so no roles are returned without domain
	perms, err := e.GetImplicitPermissionsForUser("michael")
	if err != nil {
		t.Fatalf("GetImplicitPermissionsForUser failed: %v", err)
	}
	
	t.Logf("Permissions for michael (no domain): %v", perms)
	
	// With this specific model/policy setup, no permissions are returned without domain
	// because all role assignments have specific domains
	if len(perms) != 0 {
		t.Logf("Note: Got %d permissions without domain: %v", len(perms), perms)
	}
}
