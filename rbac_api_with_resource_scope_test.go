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

func testEnforceWithTenant(t *testing.T, e *Enforcer, sub string, tenant string, obj interface{}, act string, res bool) {
	t.Helper()
	if myRes, _ := e.Enforce(sub, tenant, obj, act); myRes != res {
		t.Errorf("%s, %s, %v, %s: %t, supposed to be %t", sub, tenant, obj, act, myRes, res)
	}
}

func TestRBACWithResourceScope(t *testing.T) {
	e, _ := NewEnforcer("examples/rbac_with_resource_scope_model.conf", "examples/rbac_with_resource_scope_policy.csv")

	// Test user1 has reader role for resource1
	testGetRoles(t, e, []string{"reader"}, "user1", "resource1")
	testGetRoles(t, e, []string{}, "user1", "resource2")

	// Test user2 has reader role for resource2
	testGetRoles(t, e, []string{"reader"}, "user2", "resource2")
	testGetRoles(t, e, []string{}, "user2", "resource1")

	// Test user3 has writer role for resource1
	testGetRoles(t, e, []string{"writer"}, "user3", "resource1")
	testGetRoles(t, e, []string{}, "user3", "resource2")

	// Test enforcement - user1 can read resource1 but not resource2
	testEnforce(t, e, "user1", "resource1", "read", true)
	testEnforce(t, e, "user1", "resource2", "read", false)
	testEnforce(t, e, "user1", "resource1", "write", false)

	// Test enforcement - user2 can read resource2 but not resource1
	testEnforce(t, e, "user2", "resource1", "read", false)
	testEnforce(t, e, "user2", "resource2", "read", true)
	testEnforce(t, e, "user2", "resource2", "write", false)

	// Test enforcement - user3 can write to resource1
	testEnforce(t, e, "user3", "resource1", "write", true)
	testEnforce(t, e, "user3", "resource2", "write", false)
	testEnforce(t, e, "user3", "resource1", "read", false)

	// Test GetUsersForRole with resource scope
	testGetUsers(t, e, []string{"user1"}, "reader", "resource1")
	testGetUsers(t, e, []string{"user2"}, "reader", "resource2")
	testGetUsers(t, e, []string{"user3"}, "writer", "resource1")

	// Test AddRoleForUser with resource scope
	_, _ = e.AddRoleForUser("user4", "reader", "resource1")
	testGetRoles(t, e, []string{"reader"}, "user4", "resource1")
	testEnforce(t, e, "user4", "resource1", "read", true)
	testEnforce(t, e, "user4", "resource2", "read", false)

	// Test DeleteRoleForUser with resource scope
	_, _ = e.DeleteRoleForUser("user4", "reader", "resource1")
	testGetRoles(t, e, []string{}, "user4", "resource1")
	testEnforce(t, e, "user4", "resource1", "read", false)
}

func TestRBACWithResourceScopeAndTenant(t *testing.T) {
	e, _ := NewEnforcer("examples/rbac_with_resource_scope_tenant_model.conf", "examples/rbac_with_resource_scope_tenant_policy.csv")

	// Test user1 has reader role for tenant1::resource1
	testGetRoles(t, e, []string{"reader"}, "user1", "tenant1::resource1")
	testGetRoles(t, e, []string{}, "user1", "tenant1::resource2")

	// Test user2 has reader role for tenant1::resource2
	testGetRoles(t, e, []string{"reader"}, "user2", "tenant1::resource2")
	testGetRoles(t, e, []string{}, "user2", "tenant1::resource1")

	// Test user3 has reader role for tenant2::resource1
	testGetRoles(t, e, []string{"reader"}, "user3", "tenant2::resource1")
	testGetRoles(t, e, []string{}, "user3", "tenant1::resource1")

	// Test user4 has writer role for tenant1::resource1
	testGetRoles(t, e, []string{"writer"}, "user4", "tenant1::resource1")

	// Test enforcement - user1 can read resource1 in tenant1 only
	testEnforceWithTenant(t, e, "user1", "tenant1", "resource1", "read", true)
	testEnforceWithTenant(t, e, "user1", "tenant1", "resource2", "read", false)
	testEnforceWithTenant(t, e, "user1", "tenant2", "resource1", "read", false)
	testEnforceWithTenant(t, e, "user1", "tenant1", "resource1", "write", false)

	// Test enforcement - user2 can read resource2 in tenant1 only
	testEnforceWithTenant(t, e, "user2", "tenant1", "resource1", "read", false)
	testEnforceWithTenant(t, e, "user2", "tenant1", "resource2", "read", true)
	testEnforceWithTenant(t, e, "user2", "tenant2", "resource2", "read", false)

	// Test enforcement - user3 can read resource1 in tenant2 only
	testEnforceWithTenant(t, e, "user3", "tenant1", "resource1", "read", false)
	testEnforceWithTenant(t, e, "user3", "tenant2", "resource1", "read", true)

	// Test enforcement - user4 can write to resource1 in tenant1
	testEnforceWithTenant(t, e, "user4", "tenant1", "resource1", "write", true)
	testEnforceWithTenant(t, e, "user4", "tenant1", "resource2", "write", false)
	testEnforceWithTenant(t, e, "user4", "tenant2", "resource1", "write", false)

	// Test GetUsersForRole with tenant::resource scope
	testGetUsers(t, e, []string{"user1"}, "reader", "tenant1::resource1")
	testGetUsers(t, e, []string{"user2"}, "reader", "tenant1::resource2")
	testGetUsers(t, e, []string{"user3"}, "reader", "tenant2::resource1")

	// Test AddRoleForUser with tenant::resource scope
	_, _ = e.AddRoleForUser("user5", "reader", "tenant1::resource1")
	testGetRoles(t, e, []string{"reader"}, "user5", "tenant1::resource1")
	testEnforceWithTenant(t, e, "user5", "tenant1", "resource1", "read", true)
	testEnforceWithTenant(t, e, "user5", "tenant1", "resource2", "read", false)

	// Test DeleteRoleForUser with tenant::resource scope
	_, _ = e.DeleteRoleForUser("user5", "reader", "tenant1::resource1")
	testGetRoles(t, e, []string{}, "user5", "tenant1::resource1")
	testEnforceWithTenant(t, e, "user5", "tenant1", "resource1", "read", false)
}

func TestRBACWithResourceScopeMultitenancy(t *testing.T) {
	e, _ := NewEnforcer("examples/rbac_with_resource_scope_tenant_model.conf", "examples/rbac_with_resource_scope_tenant_policy.csv")

	// Verify isolation: user1 and user3 both have reader role, but for different tenant::resource combinations
	// user1 -> reader -> tenant1::resource1
	// user3 -> reader -> tenant2::resource1

	// user1 should only access tenant1::resource1
	testEnforceWithTenant(t, e, "user1", "tenant1", "resource1", "read", true)
	testEnforceWithTenant(t, e, "user1", "tenant2", "resource1", "read", false)

	// user3 should only access tenant2::resource1
	testEnforceWithTenant(t, e, "user3", "tenant2", "resource1", "read", true)
	testEnforceWithTenant(t, e, "user3", "tenant1", "resource1", "read", false)

	// Verify that adding a role to one user doesn't affect another user with the same role in a different scope
	_, _ = e.AddRoleForUser("user1", "writer", "tenant1::resource1")
	testEnforceWithTenant(t, e, "user1", "tenant1", "resource1", "write", true)
	testEnforceWithTenant(t, e, "user3", "tenant2", "resource1", "write", false) // user3 should not be affected

	// Clean up
	_, _ = e.DeleteRoleForUser("user1", "writer", "tenant1::resource1")
	testEnforceWithTenant(t, e, "user1", "tenant1", "resource1", "write", false)
}
