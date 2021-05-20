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

	"github.com/casbin/casbin/v2/errors"
	defaultrolemanager "github.com/casbin/casbin/v2/rbac/default-role-manager"
	"github.com/casbin/casbin/v2/util"
)

func testGetRoles(t *testing.T, e *Enforcer, res []string, name string, domain ...string) {
	t.Helper()
	myRes, err := e.GetRolesForUser(name, domain...)
	if err != nil {
		t.Error("Roles for ", name, " could not be fetched: ", err.Error())
	}
	t.Log("Roles for ", name, ": ", myRes)

	if !util.SetEquals(res, myRes) {
		t.Error("Roles for ", name, ": ", myRes, ", supposed to be ", res)
	}
}

func testGetUsers(t *testing.T, e *Enforcer, res []string, name string, domain ...string) {
	t.Helper()
	myRes, err := e.GetUsersForRole(name, domain...)
	switch err {
	case nil:
		break
	case errors.ERR_NAME_NOT_FOUND:
		t.Log("No name found")
	default:
		t.Error("Users for ", name, " could not be fetched: ", err.Error())
	}
	t.Log("Users for ", name, ": ", myRes)

	if !util.SetEquals(res, myRes) {
		t.Error("Users for ", name, ": ", myRes, ", supposed to be ", res)
	}
}

func testHasRole(t *testing.T, e *Enforcer, name string, role string, res bool, domain ...string) {
	t.Helper()
	myRes, err := e.HasRoleForUser(name, role, domain...)
	if err != nil {
		t.Error("HasRoleForUser returned an error: ", err.Error())
	}
	t.Log(name, " has role ", role, ": ", myRes)

	if res != myRes {
		t.Error(name, " has role ", role, ": ", myRes, ", supposed to be ", res)
	}
}

func TestRoleAPI(t *testing.T) {
	e, _ := NewEnforcer("examples/rbac_model.conf", "examples/rbac_policy.csv")

	testGetRoles(t, e, []string{"data2_admin"}, "alice")
	testGetRoles(t, e, []string{}, "bob")
	testGetRoles(t, e, []string{}, "data2_admin")
	testGetRoles(t, e, []string{}, "non_exist")

	testHasRole(t, e, "alice", "data1_admin", false)
	testHasRole(t, e, "alice", "data2_admin", true)

	_, _ = e.AddRoleForUser("alice", "data1_admin")

	testGetRoles(t, e, []string{"data1_admin", "data2_admin"}, "alice")
	testGetRoles(t, e, []string{}, "bob")
	testGetRoles(t, e, []string{}, "data2_admin")

	_, _ = e.DeleteRoleForUser("alice", "data1_admin")

	testGetRoles(t, e, []string{"data2_admin"}, "alice")
	testGetRoles(t, e, []string{}, "bob")
	testGetRoles(t, e, []string{}, "data2_admin")

	_, _ = e.DeleteRolesForUser("alice")

	testGetRoles(t, e, []string{}, "alice")
	testGetRoles(t, e, []string{}, "bob")
	testGetRoles(t, e, []string{}, "data2_admin")

	_, _ = e.AddRoleForUser("alice", "data1_admin")
	_, _ = e.DeleteUser("alice")

	testGetRoles(t, e, []string{}, "alice")
	testGetRoles(t, e, []string{}, "bob")
	testGetRoles(t, e, []string{}, "data2_admin")

	_, _ = e.AddRoleForUser("alice", "data2_admin")

	testEnforce(t, e, "alice", "data1", "read", false)
	testEnforce(t, e, "alice", "data1", "write", false)
	testEnforce(t, e, "alice", "data2", "read", true)
	testEnforce(t, e, "alice", "data2", "write", true)
	testEnforce(t, e, "bob", "data1", "read", false)
	testEnforce(t, e, "bob", "data1", "write", false)
	testEnforce(t, e, "bob", "data2", "read", false)
	testEnforce(t, e, "bob", "data2", "write", true)

	_, _ = e.DeleteRole("data2_admin")

	testEnforce(t, e, "alice", "data1", "read", false)
	testEnforce(t, e, "alice", "data1", "write", false)
	testEnforce(t, e, "alice", "data2", "read", false)
	testEnforce(t, e, "alice", "data2", "write", false)
	testEnforce(t, e, "bob", "data1", "read", false)
	testEnforce(t, e, "bob", "data1", "write", false)
	testEnforce(t, e, "bob", "data2", "read", false)
	testEnforce(t, e, "bob", "data2", "write", true)
}

func TestRoleAPI_Domains(t *testing.T) {
	e, _ := NewEnforcer("examples/rbac_with_domains_model.conf", "examples/rbac_with_domains_policy.csv")

	testHasRole(t, e, "alice", "admin", true, "domain1")
	testHasRole(t, e, "alice", "admin", false, "domain2")
	testGetRoles(t, e, []string{"admin"}, "alice", "domain1")
	testGetRoles(t, e, []string{}, "bob", "domain1")
	testGetRoles(t, e, []string{}, "admin", "domain1")
	testGetRoles(t, e, []string{}, "non_exist", "domain1")
	testGetRoles(t, e, []string{}, "alice", "domain2")
	testGetRoles(t, e, []string{"admin"}, "bob", "domain2")
	testGetRoles(t, e, []string{}, "admin", "domain2")
	testGetRoles(t, e, []string{}, "non_exist", "domain2")

	_, _ = e.DeleteRoleForUser("alice", "admin", "domain1")
	_, _ = e.AddRoleForUser("bob", "admin", "domain1")

	testGetRoles(t, e, []string{}, "alice", "domain1")
	testGetRoles(t, e, []string{"admin"}, "bob", "domain1")
	testGetRoles(t, e, []string{}, "admin", "domain1")
	testGetRoles(t, e, []string{}, "non_exist", "domain1")
	testGetRoles(t, e, []string{}, "alice", "domain2")
	testGetRoles(t, e, []string{"admin"}, "bob", "domain2")
	testGetRoles(t, e, []string{}, "admin", "domain2")
	testGetRoles(t, e, []string{}, "non_exist", "domain2")

	_, _ = e.AddRoleForUser("alice", "admin", "domain1")
	_, _ = e.DeleteRolesForUser("bob", "domain1")

	testGetRoles(t, e, []string{"admin"}, "alice", "domain1")
	testGetRoles(t, e, []string{}, "bob", "domain1")
	testGetRoles(t, e, []string{}, "admin", "domain1")
	testGetRoles(t, e, []string{}, "non_exist", "domain1")
	testGetRoles(t, e, []string{}, "alice", "domain2")
	testGetRoles(t, e, []string{"admin"}, "bob", "domain2")
	testGetRoles(t, e, []string{}, "admin", "domain2")
	testGetRoles(t, e, []string{}, "non_exist", "domain2")

	_, _ = e.AddRolesForUser("bob", []string{"admin", "admin1", "admin2"}, "domain1")

	testGetRoles(t, e, []string{"admin", "admin1", "admin2"}, "bob", "domain1")

	testGetPermissions(t, e, "admin", [][]string{{"admin", "domain1", "data1", "read"}, {"admin", "domain1", "data1", "write"}}, "domain1")
	testGetPermissions(t, e, "admin", [][]string{{"admin", "domain2", "data2", "read"}, {"admin", "domain2", "data2", "write"}}, "domain2")

}

func TestEnforcer_AddRolesForUser(t *testing.T) {
	e, _ := NewEnforcer("examples/rbac_model.conf", "examples/rbac_policy.csv")

	_, _ = e.AddRolesForUser("alice", []string{"data1_admin", "data2_admin", "data3_admin"})
	// The "alice" already has "data2_admin" , it will be return false. So "alice" just has "data2_admin".
	testGetRoles(t, e, []string{"data2_admin"}, "alice")
	// delete role
	_, _ = e.DeleteRoleForUser("alice", "data2_admin")

	_, _ = e.AddRolesForUser("alice", []string{"data1_admin", "data2_admin", "data3_admin"})
	testGetRoles(t, e, []string{"data1_admin", "data2_admin", "data3_admin"}, "alice")
	testEnforce(t, e, "alice", "data1", "read", true)
	testEnforce(t, e, "alice", "data2", "read", true)
	testEnforce(t, e, "alice", "data2", "write", true)
}

func testGetPermissions(t *testing.T, e *Enforcer, name string, res [][]string, domain ...string) {
	t.Helper()
	myRes := e.GetPermissionsForUser(name, domain...)
	t.Log("Permissions for ", name, ": ", myRes)

	if !util.Array2DEquals(res, myRes) {
		t.Error("Permissions for ", name, ": ", myRes, ", supposed to be ", res)
	}
}

func testHasPermission(t *testing.T, e *Enforcer, name string, permission []string, res bool) {
	t.Helper()
	myRes := e.HasPermissionForUser(name, permission...)
	t.Log(name, " has permission ", util.ArrayToString(permission), ": ", myRes)

	if res != myRes {
		t.Error(name, " has permission ", util.ArrayToString(permission), ": ", myRes, ", supposed to be ", res)
	}
}

func TestPermissionAPI(t *testing.T) {
	e, _ := NewEnforcer("examples/basic_without_resources_model.conf", "examples/basic_without_resources_policy.csv")

	testEnforceWithoutUsers(t, e, "alice", "read", true)
	testEnforceWithoutUsers(t, e, "alice", "write", false)
	testEnforceWithoutUsers(t, e, "bob", "read", false)
	testEnforceWithoutUsers(t, e, "bob", "write", true)

	testGetPermissions(t, e, "alice", [][]string{{"alice", "read"}})
	testGetPermissions(t, e, "bob", [][]string{{"bob", "write"}})

	testHasPermission(t, e, "alice", []string{"read"}, true)
	testHasPermission(t, e, "alice", []string{"write"}, false)
	testHasPermission(t, e, "bob", []string{"read"}, false)
	testHasPermission(t, e, "bob", []string{"write"}, true)

	_, _ = e.DeletePermission("read")

	testEnforceWithoutUsers(t, e, "alice", "read", false)
	testEnforceWithoutUsers(t, e, "alice", "write", false)
	testEnforceWithoutUsers(t, e, "bob", "read", false)
	testEnforceWithoutUsers(t, e, "bob", "write", true)

	_, _ = e.AddPermissionForUser("bob", "read")

	testEnforceWithoutUsers(t, e, "alice", "read", false)
	testEnforceWithoutUsers(t, e, "alice", "write", false)
	testEnforceWithoutUsers(t, e, "bob", "read", true)
	testEnforceWithoutUsers(t, e, "bob", "write", true)

	_, _ = e.DeletePermissionForUser("bob", "read")

	testEnforceWithoutUsers(t, e, "alice", "read", false)
	testEnforceWithoutUsers(t, e, "alice", "write", false)
	testEnforceWithoutUsers(t, e, "bob", "read", false)
	testEnforceWithoutUsers(t, e, "bob", "write", true)

	_, _ = e.DeletePermissionsForUser("bob")

	testEnforceWithoutUsers(t, e, "alice", "read", false)
	testEnforceWithoutUsers(t, e, "alice", "write", false)
	testEnforceWithoutUsers(t, e, "bob", "read", false)
	testEnforceWithoutUsers(t, e, "bob", "write", false)
}

func testGetImplicitRoles(t *testing.T, e *Enforcer, name string, res []string) {
	t.Helper()
	myRes, _ := e.GetImplicitRolesForUser(name)
	t.Log("Implicit roles for ", name, ": ", myRes)

	if !util.ArrayEquals(res, myRes) {
		t.Error("Implicit roles for ", name, ": ", myRes, ", supposed to be ", res)
	}
}

func testGetImplicitRolesInDomain(t *testing.T, e *Enforcer, name string, domain string, res []string) {
	t.Helper()
	myRes, _ := e.GetImplicitRolesForUser(name, domain)
	t.Log("Implicit roles in domain ", domain, " for ", name, ": ", myRes)

	if !util.ArrayEquals(res, myRes) {
		t.Error("Implicit roles in domain ", domain, " for ", name, ": ", myRes, ", supposed to be ", res)
	}
}

func TestImplicitRoleAPI(t *testing.T) {
	e, _ := NewEnforcer("examples/rbac_model.conf", "examples/rbac_with_hierarchy_policy.csv")

	testGetPermissions(t, e, "alice", [][]string{{"alice", "data1", "read"}})
	testGetPermissions(t, e, "bob", [][]string{{"bob", "data2", "write"}})

	testGetImplicitRoles(t, e, "alice", []string{"admin", "data1_admin", "data2_admin"})
	testGetImplicitRoles(t, e, "bob", []string{})

	e, _ = NewEnforcer("examples/rbac_with_pattern_model.conf", "examples/rbac_with_pattern_policy.csv")

	e.GetRoleManager().(*defaultrolemanager.RoleManager).AddMatchingFunc("matcher", util.KeyMatch)
	e.AddNamedMatchingFunc("g2", "matcher", util.KeyMatch)

	//testGetImplicitRoles(t, e, "cathy", []string{"/book/1/2/3/4/5", "pen_admin", "/book/*", "book_group"})
	testGetImplicitRoles(t, e, "cathy", []string{"/book/1/2/3/4/5", "pen_admin"})
	testGetRoles(t, e, []string{"/book/1/2/3/4/5", "pen_admin"}, "cathy")
}

func testGetImplicitPermissions(t *testing.T, e *Enforcer, name string, res [][]string) {
	t.Helper()
	myRes, _ := e.GetImplicitPermissionsForUser(name)
	t.Log("Implicit permissions for ", name, ": ", myRes)

	if !util.Array2DEquals(res, myRes) {
		t.Error("Implicit permissions for ", name, ": ", myRes, ", supposed to be ", res)
	}
}

func testGetImplicitPermissionsWithDomain(t *testing.T, e *Enforcer, name string, domain string, res [][]string) {
	t.Helper()
	myRes, _ := e.GetImplicitPermissionsForUser(name, domain)
	t.Log("Implicit permissions for", name, "under", domain, ":", myRes)

	if !util.Array2DEquals(res, myRes) {
		t.Error("Implicit permissions for", name, "under", domain, ":", myRes, ", supposed to be ", res)
	}
}

func TestImplicitPermissionAPI(t *testing.T) {
	e, _ := NewEnforcer("examples/rbac_model.conf", "examples/rbac_with_hierarchy_policy.csv")

	testGetPermissions(t, e, "alice", [][]string{{"alice", "data1", "read"}})
	testGetPermissions(t, e, "bob", [][]string{{"bob", "data2", "write"}})

	testGetImplicitPermissions(t, e, "alice", [][]string{{"alice", "data1", "read"}, {"data1_admin", "data1", "read"}, {"data1_admin", "data1", "write"}, {"data2_admin", "data2", "read"}, {"data2_admin", "data2", "write"}})
	testGetImplicitPermissions(t, e, "bob", [][]string{{"bob", "data2", "write"}})
}

func TestImplicitPermissionAPIWithDomain(t *testing.T) {
	e, _ := NewEnforcer("examples/rbac_with_domains_model.conf", "examples/rbac_with_hierarchy_with_domains_policy.csv")
	testGetImplicitPermissionsWithDomain(t, e, "alice", "domain1", [][]string{{"alice", "domain1", "data2", "read"}, {"role:reader", "domain1", "data1", "read"}, {"role:writer", "domain1", "data1", "write"}})
}

func testGetImplicitUsers(t *testing.T, e *Enforcer, res []string, permission ...string) {
	t.Helper()
	myRes, _ := e.GetImplicitUsersForPermission(permission...)
	t.Log("Implicit users for permission: ", permission, ": ", myRes)

	sort.Strings(res)
	sort.Strings(myRes)

	if !util.ArrayEquals(res, myRes) {
		t.Error("Implicit users for permission: ", permission, ": ", myRes, ", supposed to be ", res)
	}
}

func TestImplicitUserAPI(t *testing.T) {
	e, _ := NewEnforcer("examples/rbac_model.conf", "examples/rbac_with_hierarchy_policy.csv")

	testGetImplicitUsers(t, e, []string{"alice"}, "data1", "read")
	testGetImplicitUsers(t, e, []string{"alice"}, "data1", "write")
	testGetImplicitUsers(t, e, []string{"alice"}, "data2", "read")
	testGetImplicitUsers(t, e, []string{"alice", "bob"}, "data2", "write")

	e.ClearPolicy()
	_, _ = e.AddPolicy("admin", "data1", "read")
	_, _ = e.AddPolicy("bob", "data1", "read")
	_, _ = e.AddGroupingPolicy("alice", "admin")
	testGetImplicitUsers(t, e, []string{"alice", "bob"}, "data1", "read")
}

func testGetImplicitResourcesForUser(t *testing.T, e *Enforcer, res [][]string, user string, domain ...string) {
	t.Helper()
	myRes, _ := e.GetImplicitResourcesForUser(user, domain...)
	t.Log("Implicit resources for user: ", user, ": ", myRes)

	lessFunc := func(arr [][]string) func(int, int) bool {
		return func(i, j int) bool {
			policy1, policy2 := arr[i], arr[j]
			for k := range policy1 {
				if policy1[k] == policy2[k] {
					continue
				}
				return policy1[k] < policy2[k]
			}
			return true
		}
	}

	sort.Slice(res, lessFunc(res))
	sort.Slice(myRes, lessFunc(myRes))

	if !util.Array2DEquals(res, myRes) {
		t.Error("Implicit resources for user: ", user, ": ", myRes, ", supposed to be ", res)
	}
}

func TestGetImplicitResourcesForUser(t *testing.T) {
	e, _ := NewEnforcer("examples/rbac_with_pattern_model.conf", "examples/rbac_with_pattern_policy.csv")
	testGetImplicitResourcesForUser(t, e, [][]string{
		{"alice", "/pen/1", "GET"},
		{"alice", "/pen2/1", "GET"},
		{"alice", "/book/:id", "GET"},
		{"alice", "/book2/{id}", "GET"},
		{"alice", "/book/*", "GET"},
		{"alice", "book_group", "GET"},
	}, "alice")
	testGetImplicitResourcesForUser(t, e, [][]string{
		{"bob", "pen_group", "GET"},
		{"bob", "/pen/:id", "GET"},
		{"bob", "/pen2/{id}", "GET"},
	}, "bob")
	testGetImplicitResourcesForUser(t, e, [][]string{
		{"cathy", "pen_group", "GET"},
		{"cathy", "/pen/:id", "GET"},
		{"cathy", "/pen2/{id}", "GET"},
	}, "cathy")
}

func TestImplicitUsersForRole(t *testing.T) {
	e, _ := NewEnforcer("examples/rbac_with_pattern_model.conf", "examples/rbac_with_pattern_policy.csv")

	testGetImplicitUsersForRole(t, e, "book_admin", []string{"alice"})
	testGetImplicitUsersForRole(t, e, "pen_admin", []string{"cathy", "bob"})

	testGetImplicitUsersForRole(t, e, "book_group", []string{"/book/*", "/book/:id", "/book2/{id}"})
	testGetImplicitUsersForRole(t, e, "pen_group", []string{"/pen/:id", "/pen2/{id}"})
}

func testGetImplicitUsersForRole(t *testing.T, e *Enforcer, name string, res []string) {
	t.Helper()
	myRes, _ := e.GetImplicitUsersForRole(name)
	t.Log("Implicit users for ", name, ": ", myRes)
	sort.Strings(res)
	sort.Strings(myRes)

	if !util.ArrayEquals(res, myRes) {
		t.Error("Implicit users for ", name, ": ", myRes, ", supposed to be ", res)
	}
}
