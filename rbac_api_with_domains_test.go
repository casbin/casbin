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
	"log"
	"testing"

	"github.com/casbin/casbin/util"
)

func testGetRolesInDomain(t *testing.T, e *Enforcer, name string, domain string, res []string) {
	t.Helper()
	myRes := e.GetRolesForUserInDomain(name, domain)
	log.Print("Roles for ", name, " under ", domain, " : ", myRes)

	if !util.SetEquals(res, myRes) {
		t.Error("Roles for", name, "under", domain, ":", myRes, ", supposed to be ", res)
	}
}

func TestRoleAPIWithDomains(t *testing.T) {
	e := NewEnforcer("examples/rbac_with_domains_model.conf", "examples/rbac_with_domains_policy.csv")

	testGetRolesInDomain(t, e, "alice", "domain1", []string{"admin"})
	testGetRolesInDomain(t, e, "bob", "domain1", []string{})
	testGetRolesInDomain(t, e, "admin", "domain1", []string{})
	testGetRolesInDomain(t, e, "non_exist", "domain1", []string{})

	testGetRolesInDomain(t, e, "alice", "domain2", []string{})
	testGetRolesInDomain(t, e, "bob", "domain2", []string{"admin"})
	testGetRolesInDomain(t, e, "admin", "domain2", []string{})
	testGetRolesInDomain(t, e, "non_exist", "domain2", []string{})

	e.DeleteRoleForUserInDomain("alice", "admin", "domain1")
	e.AddRoleForUserInDomain("bob", "admin", "domain1")

	testGetRolesInDomain(t, e, "alice", "domain1", []string{})
	testGetRolesInDomain(t, e, "bob", "domain1", []string{"admin"})
	testGetRolesInDomain(t, e, "admin", "domain1", []string{})
	testGetRolesInDomain(t, e, "non_exist", "domain1", []string{})

	testGetRolesInDomain(t, e, "alice", "domain2", []string{})
	testGetRolesInDomain(t, e, "bob", "domain2", []string{"admin"})
	testGetRolesInDomain(t, e, "admin", "domain2", []string{})
	testGetRolesInDomain(t, e, "non_exist", "domain2", []string{})
}

func testGetPermissionsInDomain(t *testing.T, e *Enforcer, name string, domain string, res [][]string) {
	t.Helper()
	myRes := e.GetPermissionsForUserInDomain(name, domain)
	log.Print("Permissions for ", name, " under ", domain, " : ", myRes)

	if !util.Array2DEquals(res, myRes) {
		t.Error("Permissions for", name, "under", domain, ":", myRes, ", supposed to be ", res)
	}
}

func TestPermissionAPIInDomain(t *testing.T) {
	e := NewEnforcer("examples/rbac_with_domains_model.conf", "examples/rbac_with_domains_policy.csv")

	testGetPermissionsInDomain(t, e, "alice", "domain1", [][]string{})
	testGetPermissionsInDomain(t, e, "bob", "domain1", [][]string{})
	testGetPermissionsInDomain(t, e, "admin", "domain1", [][]string{{"admin", "domain1", "data1", "read"}, {"admin", "domain1", "data1", "write"}})
	testGetPermissionsInDomain(t, e, "non_exist", "domain1", [][]string{})

	testGetPermissionsInDomain(t, e, "alice", "domain2", [][]string{})
	testGetPermissionsInDomain(t, e, "bob", "domain2", [][]string{})
	testGetPermissionsInDomain(t, e, "admin", "domain2", [][]string{{"admin", "domain2", "data2", "read"}, {"admin", "domain2", "data2", "write"}})
	testGetPermissionsInDomain(t, e, "non_exist", "domain2", [][]string{})
}
