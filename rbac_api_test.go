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
	"github.com/casbin/casbin/util"
	"log"
	"testing"
)

func testGetRoles(t *testing.T, e *Enforcer, name string, res []string) {
	myRes := e.GetRolesForUser(name)
	log.Print("Roles for ", name, ": ", myRes)

	if !util.SetEquals(res, myRes) {
		t.Error("Roles for ", name, ": ", myRes, ", supposed to be ", res)
	}
}

func TestRoleAPI(t *testing.T) {
	e := NewEnforcer("examples/rbac_model.conf", "examples/rbac_policy.csv")

	testGetRoles(t, e, "alice", []string{"data2_admin"})
	testGetRoles(t, e, "bob", []string{})
	testGetRoles(t, e, "data2_admin", []string{})
	testGetRoles(t, e, "non_exist", []string{})

	e.AddRoleForUser("alice", "data1_admin")

	testGetRoles(t, e, "alice", []string{"data1_admin", "data2_admin"})
	testGetRoles(t, e, "bob", []string{})
	testGetRoles(t, e, "data2_admin", []string{})

	e.DeleteRolesForUser("alice")

	testGetRoles(t, e, "alice", []string{})
	testGetRoles(t, e, "bob", []string{})
	testGetRoles(t, e, "data2_admin", []string{})

	e.AddRoleForUser("alice", "data1_admin")
	e.DeleteUser("alice")

	testGetRoles(t, e, "alice", []string{})
	testGetRoles(t, e, "bob", []string{})
	testGetRoles(t, e, "data2_admin", []string{})

	e.AddRoleForUser("alice", "data2_admin")

	testEnforce(t, e, "alice", "data1", "read", true)
	testEnforce(t, e, "alice", "data1", "write", false)
	testEnforce(t, e, "alice", "data2", "read", true)
	testEnforce(t, e, "alice", "data2", "write", true)
	testEnforce(t, e, "bob", "data1", "read", false)
	testEnforce(t, e, "bob", "data1", "write", false)
	testEnforce(t, e, "bob", "data2", "read", false)
	testEnforce(t, e, "bob", "data2", "write", true)

	e.DeleteRole("data2_admin")

	testEnforce(t, e, "alice", "data1", "read", true)
	testEnforce(t, e, "alice", "data1", "write", false)
	testEnforce(t, e, "alice", "data2", "read", false)
	testEnforce(t, e, "alice", "data2", "write", false)
	testEnforce(t, e, "bob", "data1", "read", false)
	testEnforce(t, e, "bob", "data1", "write", false)
	testEnforce(t, e, "bob", "data2", "read", false)
	testEnforce(t, e, "bob", "data2", "write", true)
}

func testGetPermissions(t *testing.T, e *Enforcer, name string, res []string) {
	myRes := e.GetPermissionsForUser(name)
	log.Print("Permissions for ", name, ": ", myRes)

	if !util.SetEquals(res, myRes) {
		t.Error("Permissions for ", name, ": ", myRes, ", supposed to be ", res)
	}
}

func TestPermissionAPI(t *testing.T) {
	e := NewEnforcer("examples/basic_model_without_resources.conf", "examples/basic_policy_without_resources.csv")

	testEnforceWithoutUsers(t, e, "alice", "read", true)
	testEnforceWithoutUsers(t, e, "alice", "write", false)
	testEnforceWithoutUsers(t, e, "bob", "read", false)
	testEnforceWithoutUsers(t, e, "bob", "write", true)

	testGetPermissions(t, e, "alice", []string{"read"})
	testGetPermissions(t, e, "bob", []string{"write"})

	e.DeletePermission("read")

	testEnforceWithoutUsers(t, e, "alice", "read", false)
	testEnforceWithoutUsers(t, e, "alice", "write", false)
	testEnforceWithoutUsers(t, e, "bob", "read", false)
	testEnforceWithoutUsers(t, e, "bob", "write", true)

	e.AddPermissionForUser("bob", "read")

	testEnforceWithoutUsers(t, e, "alice", "read", false)
	testEnforceWithoutUsers(t, e, "alice", "write", false)
	testEnforceWithoutUsers(t, e, "bob", "read", true)
	testEnforceWithoutUsers(t, e, "bob", "write", true)

	e.DeletePermissionsForUser("bob")

	testEnforceWithoutUsers(t, e, "alice", "read", false)
	testEnforceWithoutUsers(t, e, "alice", "write", false)
	testEnforceWithoutUsers(t, e, "bob", "read", false)
	testEnforceWithoutUsers(t, e, "bob", "write", false)
}
