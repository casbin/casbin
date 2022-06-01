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
	"sort"
	"testing"

	"github.com/casbin/casbin/v2/util"
)

// testGetUsersInDomain: Add by Gordon
func testGetUsersInDomain(t *testing.T, e *Enforcer, name string, domain string, res []string) {
	t.Helper()
	myRes := e.GetUsersForRoleInDomain(name, domain)
	t.Log("Users for ", name, " under ", domain, ": ", myRes)

	if !util.SetEquals(res, myRes) {
		t.Error("Users for ", name, " under ", domain, ": ", myRes, ", supposed to be ", res)
	}
}

func testGetRolesInDomain(t *testing.T, e *Enforcer, name string, domain string, res []string) {
	t.Helper()
	myRes := e.GetRolesForUserInDomain(name, domain)
	t.Log("Roles for ", name, " under ", domain, ": ", myRes)

	if !util.SetEquals(res, myRes) {
		t.Error("Roles for ", name, " under ", domain, ": ", myRes, ", supposed to be ", res)
	}
}

func TestGetImplicitRolesForDomainUser(t *testing.T) {
	e, _ := NewEnforcer("examples/rbac_with_domains_model.conf", "examples/rbac_with_hierarchy_with_domains_policy.csv")

	// This is only able to retrieve the first level of roles.
	testGetRolesInDomain(t, e, "alice", "domain1", []string{"role:global_admin"})

	// Retrieve all inherit roles. It supports domains as well.
	testGetImplicitRolesInDomain(t, e, "alice", "domain1", []string{"role:global_admin", "role:reader", "role:writer"})
}

// TestUserAPIWithDomains: Add by Gordon
func TestUserAPIWithDomains(t *testing.T) {
	e, _ := NewEnforcer("examples/rbac_with_domains_model.conf", "examples/rbac_with_domains_policy.csv")

	testGetUsers(t, e, []string{"alice"}, "admin", "domain1")
	testGetUsersInDomain(t, e, "admin", "domain1", []string{"alice"})

	testGetUsers(t, e, []string{}, "non_exist", "domain1")
	testGetUsersInDomain(t, e, "non_exist", "domain1", []string{})

	testGetUsers(t, e, []string{"bob"}, "admin", "domain2")
	testGetUsersInDomain(t, e, "admin", "domain2", []string{"bob"})

	testGetUsers(t, e, []string{}, "non_exist", "domain2")
	testGetUsersInDomain(t, e, "non_exist", "domain2", []string{})

	_, _ = e.DeleteRoleForUserInDomain("alice", "admin", "domain1")
	_, _ = e.AddRoleForUserInDomain("bob", "admin", "domain1")

	testGetUsers(t, e, []string{"bob"}, "admin", "domain1")
	testGetUsersInDomain(t, e, "admin", "domain1", []string{"bob"})

	testGetUsers(t, e, []string{}, "non_exist", "domain1")
	testGetUsersInDomain(t, e, "non_exist", "domain1", []string{})

	testGetUsers(t, e, []string{"bob"}, "admin", "domain2")
	testGetUsersInDomain(t, e, "admin", "domain2", []string{"bob"})

	testGetUsers(t, e, []string{}, "non_exist", "domain2")
	testGetUsersInDomain(t, e, "non_exist", "domain2", []string{})
}

func TestRoleAPIWithDomains(t *testing.T) {
	e, _ := NewEnforcer("examples/rbac_with_domains_model.conf", "examples/rbac_with_domains_policy.csv")

	testGetRoles(t, e, []string{"admin"}, "alice", "domain1")
	testGetRolesInDomain(t, e, "alice", "domain1", []string{"admin"})

	testGetRoles(t, e, []string{}, "bob", "domain1")
	testGetRolesInDomain(t, e, "bob", "domain1", []string{})

	testGetRoles(t, e, []string{}, "admin", "domain1")
	testGetRolesInDomain(t, e, "admin", "domain1", []string{})

	testGetRoles(t, e, []string{}, "non_exist", "domain1")
	testGetRolesInDomain(t, e, "non_exist", "domain1", []string{})

	testGetRoles(t, e, []string{}, "alice", "domain2")
	testGetRolesInDomain(t, e, "alice", "domain2", []string{})

	testGetRoles(t, e, []string{"admin"}, "bob", "domain2")
	testGetRolesInDomain(t, e, "bob", "domain2", []string{"admin"})

	testGetRoles(t, e, []string{}, "admin", "domain2")
	testGetRolesInDomain(t, e, "admin", "domain2", []string{})

	testGetRoles(t, e, []string{}, "non_exist", "domain2")
	testGetRolesInDomain(t, e, "non_exist", "domain2", []string{})

	_, _ = e.DeleteRoleForUserInDomain("alice", "admin", "domain1")
	_, _ = e.AddRoleForUserInDomain("bob", "admin", "domain1")

	testGetRoles(t, e, []string{}, "alice", "domain1")
	testGetRolesInDomain(t, e, "alice", "domain1", []string{})

	testGetRoles(t, e, []string{"admin"}, "bob", "domain1")
	testGetRolesInDomain(t, e, "bob", "domain1", []string{"admin"})

	testGetRoles(t, e, []string{}, "admin", "domain1")
	testGetRolesInDomain(t, e, "admin", "domain1", []string{})

	testGetRoles(t, e, []string{}, "non_exist", "domain1")
	testGetRolesInDomain(t, e, "non_exist", "domain1", []string{})

	testGetRoles(t, e, []string{}, "alice", "domain2")
	testGetRolesInDomain(t, e, "alice", "domain2", []string{})

	testGetRoles(t, e, []string{"admin"}, "bob", "domain2")
	testGetRolesInDomain(t, e, "bob", "domain2", []string{"admin"})

	testGetRoles(t, e, []string{}, "admin", "domain2")
	testGetRolesInDomain(t, e, "admin", "domain2", []string{})

	testGetRoles(t, e, []string{}, "non_exist", "domain2")
	testGetRolesInDomain(t, e, "non_exist", "domain2", []string{})

	_, _ = e.AddRoleForUserInDomain("alice", "admin", "domain1")
	_, _ = e.DeleteRolesForUserInDomain("bob", "domain1")

	testGetRoles(t, e, []string{"admin"}, "alice", "domain1")
	testGetRolesInDomain(t, e, "alice", "domain1", []string{"admin"})

	testGetRoles(t, e, []string{}, "bob", "domain1")
	testGetRolesInDomain(t, e, "bob", "domain1", []string{})

	testGetRoles(t, e, []string{}, "admin", "domain1")
	testGetRolesInDomain(t, e, "admin", "domain1", []string{})

	testGetRoles(t, e, []string{}, "non_exist", "domain1")
	testGetRolesInDomain(t, e, "non_exist", "domain1", []string{})

	testGetRoles(t, e, []string{}, "alice", "domain2")
	testGetRolesInDomain(t, e, "alice", "domain2", []string{})

	testGetRoles(t, e, []string{"admin"}, "bob", "domain2")
	testGetRolesInDomain(t, e, "bob", "domain2", []string{"admin"})

	testGetRoles(t, e, []string{}, "admin", "domain2")
	testGetRolesInDomain(t, e, "admin", "domain2", []string{})

	testGetRoles(t, e, []string{}, "non_exist", "domain2")
	testGetRolesInDomain(t, e, "non_exist", "domain2", []string{})

}

func testGetPermissionsInDomain(t *testing.T, e *Enforcer, name string, domain string, res [][]string) {
	t.Helper()
	myRes := e.GetPermissionsForUserInDomain(name, domain)
	t.Log("Permissions for ", name, " under ", domain, ": ", myRes)

	if !util.Array2DEquals(res, myRes) {
		t.Error("Permissions for ", name, " under ", domain, ": ", myRes, ", supposed to be ", res)
	}
}

func TestPermissionAPIInDomain(t *testing.T) {
	e, _ := NewEnforcer("examples/rbac_with_domains_model.conf", "examples/rbac_with_domains_policy.csv")

	testGetPermissionsInDomain(t, e, "alice", "domain1", [][]string{{"admin", "domain1", "data1", "read"}, {"admin", "domain1", "data1", "write"}})
	testGetPermissionsInDomain(t, e, "bob", "domain1", [][]string{})
	testGetPermissionsInDomain(t, e, "admin", "domain1", [][]string{{"admin", "domain1", "data1", "read"}, {"admin", "domain1", "data1", "write"}})
	testGetPermissionsInDomain(t, e, "non_exist", "domain1", [][]string{})

	testGetPermissionsInDomain(t, e, "alice", "domain2", [][]string{})
	testGetPermissionsInDomain(t, e, "bob", "domain2", [][]string{{"admin", "domain2", "data2", "read"}, {"admin", "domain2", "data2", "write"}})
	testGetPermissionsInDomain(t, e, "admin", "domain2", [][]string{{"admin", "domain2", "data2", "read"}, {"admin", "domain2", "data2", "write"}})
	testGetPermissionsInDomain(t, e, "non_exist", "domain2", [][]string{})
}

func testGetDomainsForUser(t *testing.T, e *Enforcer, res []string, user string) {
	t.Helper()
	myRes, _ := e.GetDomainsForUser(user)

	sort.Strings(myRes)
	sort.Strings(res)

	if !util.SetEquals(res, myRes) {
		t.Error("domains for user: ", user, ": ", myRes, ",  supposed to be ", res)
	}
}

func TestGetDomainsForUser(t *testing.T) {
	e, _ := NewEnforcer("examples/rbac_with_domains_model.conf", "examples/rbac_with_domains_policy2.csv")

	testGetDomainsForUser(t, e, []string{"domain1", "domain2"}, "alice")
	testGetDomainsForUser(t, e, []string{"domain2", "domain3"}, "bob")
	testGetDomainsForUser(t, e, []string{"domain3"}, "user")
}

func testGetAllUsersByDomain(t *testing.T, e *Enforcer, domain string, expected []string) {
	if !util.SetEquals(e.GetAllUsersByDomain(domain), expected) {
		t.Errorf("users in %s: %v, supposed to be %v\n", domain, e.GetAllUsersByDomain(domain), expected)
	}
}

func TestGetAllUsersByDomain(t *testing.T) {
	e, _ := NewEnforcer("examples/rbac_with_domains_model.conf", "examples/rbac_with_domains_policy.csv")

	testGetAllUsersByDomain(t, e, "domain1", []string{"alice", "admin"})
	testGetAllUsersByDomain(t, e, "domain2", []string{"bob", "admin"})
}

func testDeleteAllUsersByDomain(t *testing.T, domain string, expectedPolicy, expectedGroupingPolicy [][]string) {
	e, _ := NewEnforcer("examples/rbac_with_domains_model.conf", "examples/rbac_with_domains_policy.csv")

	_, _ = e.DeleteAllUsersByDomain(domain)
	if !util.Array2DEquals(e.GetPolicy(), expectedPolicy) {
		t.Errorf("policy in %s: %v, supposed to be %v\n", domain, e.GetPolicy(), expectedPolicy)
	}
	if !util.Array2DEquals(e.GetGroupingPolicy(), expectedGroupingPolicy) {
		t.Errorf("grouping policy in %s: %v, supposed to be %v\n", domain, e.GetGroupingPolicy(), expectedGroupingPolicy)
	}
}

func TestDeleteAllUsersByDomain(t *testing.T) {
	testDeleteAllUsersByDomain(t, "domain1", [][]string{
		{"admin", "domain2", "data2", "read"},
		{"admin", "domain2", "data2", "write"},
	}, [][]string{
		{"bob", "admin", "domain2"},
	})
	testDeleteAllUsersByDomain(t, "domain2", [][]string{
		{"admin", "domain1", "data1", "read"},
		{"admin", "domain1", "data1", "write"},
	}, [][]string{
		{"alice", "admin", "domain1"},
	})
}

// testGetAllDomains tests GetAllDomains()
func testGetAllDomains(t *testing.T, e *Enforcer, res []string) {
	t.Helper()
	myRes, _ := e.GetAllDomains()
	sort.Strings(myRes)
	sort.Strings(res)
	if !util.ArrayEquals(res, myRes) {
		t.Error("domains: ", myRes, ", supposed to be ", res)
	}
}

func TestGetAllDomains(t *testing.T) {
	e, _ := NewEnforcer("examples/rbac_with_domains_model.conf", "examples/rbac_with_domains_policy.csv")

	testGetAllDomains(t, e, []string{"domain1", "domain2"})
}
