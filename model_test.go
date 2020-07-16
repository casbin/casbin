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

	fileadapter "github.com/casbin/casbin/v3/persist/file-adapter"
	"github.com/casbin/casbin/v3/rbac"
	defaultrolemanager "github.com/casbin/casbin/v3/rbac/default-role-manager"
	"github.com/casbin/casbin/v3/util"
)

func testEnforce(t *testing.T, e *Enforcer, sub interface{}, obj interface{}, act string, res bool) {
	t.Helper()
	if myRes, _ := e.Enforce(sub, obj, act); myRes != res {
		t.Errorf("%s, %v, %s: %t, supposed to be %t", sub, obj, act, myRes, res)
	}
}

func testEnforceWithoutUsers(t *testing.T, e *Enforcer, obj string, act string, res bool) {
	t.Helper()
	if myRes, _ := e.Enforce(obj, act); myRes != res {
		t.Errorf("%s, %s: %t, supposed to be %t", obj, act, myRes, res)
	}
}

func testDomainEnforce(t *testing.T, e *Enforcer, sub string, dom string, obj string, act string, res bool) {
	t.Helper()
	if myRes, _ := e.Enforce(sub, dom, obj, act); myRes != res {
		t.Errorf("%s, %s, %s, %s: %t, supposed to be %t", sub, dom, obj, act, myRes, res)
	}
}

func TestBasicModel(t *testing.T) {
	e, _ := NewEnforcer("examples/basic_model.conf", "examples/basic_policy.csv")

	testEnforce(t, e, "alice", "data1", "read", true)
	testEnforce(t, e, "alice", "data1", "write", false)
	testEnforce(t, e, "alice", "data2", "read", false)
	testEnforce(t, e, "alice", "data2", "write", false)
	testEnforce(t, e, "bob", "data1", "read", false)
	testEnforce(t, e, "bob", "data1", "write", false)
	testEnforce(t, e, "bob", "data2", "read", false)
	testEnforce(t, e, "bob", "data2", "write", true)
}

func TestBasicModelWithoutSpaces(t *testing.T) {
	e, _ := NewEnforcer("examples/basic_model_without_spaces.conf", "examples/basic_policy.csv")

	testEnforce(t, e, "alice", "data1", "read", true)
	testEnforce(t, e, "alice", "data1", "write", false)
	testEnforce(t, e, "alice", "data2", "read", false)
	testEnforce(t, e, "alice", "data2", "write", false)
	testEnforce(t, e, "bob", "data1", "read", false)
	testEnforce(t, e, "bob", "data1", "write", false)
	testEnforce(t, e, "bob", "data2", "read", false)
	testEnforce(t, e, "bob", "data2", "write", true)
}

func TestBasicModelNoPolicy(t *testing.T) {
	e, _ := NewEnforcer("examples/basic_model.conf")

	testEnforce(t, e, "alice", "data1", "read", false)
	testEnforce(t, e, "alice", "data1", "write", false)
	testEnforce(t, e, "alice", "data2", "read", false)
	testEnforce(t, e, "alice", "data2", "write", false)
	testEnforce(t, e, "bob", "data1", "read", false)
	testEnforce(t, e, "bob", "data1", "write", false)
	testEnforce(t, e, "bob", "data2", "read", false)
	testEnforce(t, e, "bob", "data2", "write", false)
}

func TestBasicModelWithRoot(t *testing.T) {
	e, _ := NewEnforcer("examples/basic_with_root_model.conf", "examples/basic_policy.csv")

	testEnforce(t, e, "alice", "data1", "read", true)
	testEnforce(t, e, "alice", "data1", "write", false)
	testEnforce(t, e, "alice", "data2", "read", false)
	testEnforce(t, e, "alice", "data2", "write", false)
	testEnforce(t, e, "bob", "data1", "read", false)
	testEnforce(t, e, "bob", "data1", "write", false)
	testEnforce(t, e, "bob", "data2", "read", false)
	testEnforce(t, e, "bob", "data2", "write", true)
	testEnforce(t, e, "root", "data1", "read", true)
	testEnforce(t, e, "root", "data1", "write", true)
	testEnforce(t, e, "root", "data2", "read", true)
	testEnforce(t, e, "root", "data2", "write", true)
}

func TestBasicModelWithRootNoPolicy(t *testing.T) {
	e, _ := NewEnforcer("examples/basic_with_root_model.conf")

	testEnforce(t, e, "alice", "data1", "read", false)
	testEnforce(t, e, "alice", "data1", "write", false)
	testEnforce(t, e, "alice", "data2", "read", false)
	testEnforce(t, e, "alice", "data2", "write", false)
	testEnforce(t, e, "bob", "data1", "read", false)
	testEnforce(t, e, "bob", "data1", "write", false)
	testEnforce(t, e, "bob", "data2", "read", false)
	testEnforce(t, e, "bob", "data2", "write", false)
	testEnforce(t, e, "root", "data1", "read", true)
	testEnforce(t, e, "root", "data1", "write", true)
	testEnforce(t, e, "root", "data2", "read", true)
	testEnforce(t, e, "root", "data2", "write", true)
}

func TestBasicModelWithoutUsers(t *testing.T) {
	e, _ := NewEnforcer("examples/basic_without_users_model.conf", "examples/basic_without_users_policy.csv")

	testEnforceWithoutUsers(t, e, "data1", "read", true)
	testEnforceWithoutUsers(t, e, "data1", "write", false)
	testEnforceWithoutUsers(t, e, "data2", "read", false)
	testEnforceWithoutUsers(t, e, "data2", "write", true)
}

func TestBasicModelWithoutResources(t *testing.T) {
	e, _ := NewEnforcer("examples/basic_without_resources_model.conf", "examples/basic_without_resources_policy.csv")

	testEnforceWithoutUsers(t, e, "alice", "read", true)
	testEnforceWithoutUsers(t, e, "alice", "write", false)
	testEnforceWithoutUsers(t, e, "bob", "read", false)
	testEnforceWithoutUsers(t, e, "bob", "write", true)
}

func TestRBACModel(t *testing.T) {
	e, _ := NewEnforcer("examples/rbac_model.conf", "examples/rbac_policy.csv")

	testEnforce(t, e, "alice", "data1", "read", true)
	testEnforce(t, e, "alice", "data1", "write", false)
	testEnforce(t, e, "alice", "data2", "read", true)
	testEnforce(t, e, "alice", "data2", "write", true)
	testEnforce(t, e, "bob", "data1", "read", false)
	testEnforce(t, e, "bob", "data1", "write", false)
	testEnforce(t, e, "bob", "data2", "read", false)
	testEnforce(t, e, "bob", "data2", "write", true)
}

func TestRBACModelWithResourceRoles(t *testing.T) {
	e, _ := NewEnforcer("examples/rbac_with_resource_roles_model.conf", "examples/rbac_with_resource_roles_policy.csv")

	testEnforce(t, e, "alice", "data1", "read", true)
	testEnforce(t, e, "alice", "data1", "write", true)
	testEnforce(t, e, "alice", "data2", "read", false)
	testEnforce(t, e, "alice", "data2", "write", true)
	testEnforce(t, e, "bob", "data1", "read", false)
	testEnforce(t, e, "bob", "data1", "write", false)
	testEnforce(t, e, "bob", "data2", "read", false)
	testEnforce(t, e, "bob", "data2", "write", true)
}

func TestRBACModelWithDomains(t *testing.T) {
	e, _ := NewEnforcer("examples/rbac_with_domains_model.conf", "examples/rbac_with_domains_policy.csv")

	testDomainEnforce(t, e, "alice", "domain1", "data1", "read", true)
	testDomainEnforce(t, e, "alice", "domain1", "data1", "write", true)
	testDomainEnforce(t, e, "alice", "domain1", "data2", "read", false)
	testDomainEnforce(t, e, "alice", "domain1", "data2", "write", false)
	testDomainEnforce(t, e, "bob", "domain2", "data1", "read", false)
	testDomainEnforce(t, e, "bob", "domain2", "data1", "write", false)
	testDomainEnforce(t, e, "bob", "domain2", "data2", "read", true)
	testDomainEnforce(t, e, "bob", "domain2", "data2", "write", true)
}

func TestRBACModelWithDomainsAtRuntime(t *testing.T) {
	e, _ := NewEnforcer("examples/rbac_with_domains_model.conf")

	_, _ = e.AddPolicy("admin", "domain1", "data1", "read")
	_, _ = e.AddPolicy("admin", "domain1", "data1", "write")
	_, _ = e.AddPolicy("admin", "domain2", "data2", "read")
	_, _ = e.AddPolicy("admin", "domain2", "data2", "write")

	_, _ = e.AddGroupingPolicy("alice", "admin", "domain1")
	_, _ = e.AddGroupingPolicy("bob", "admin", "domain2")

	testDomainEnforce(t, e, "alice", "domain1", "data1", "read", true)
	testDomainEnforce(t, e, "alice", "domain1", "data1", "write", true)
	testDomainEnforce(t, e, "alice", "domain1", "data2", "read", false)
	testDomainEnforce(t, e, "alice", "domain1", "data2", "write", false)
	testDomainEnforce(t, e, "bob", "domain2", "data1", "read", false)
	testDomainEnforce(t, e, "bob", "domain2", "data1", "write", false)
	testDomainEnforce(t, e, "bob", "domain2", "data2", "read", true)
	testDomainEnforce(t, e, "bob", "domain2", "data2", "write", true)

	// Remove all policy rules related to domain1 and data1.
	_, _ = e.RemoveFilteredPolicy(1, "domain1", "data1")

	testDomainEnforce(t, e, "alice", "domain1", "data1", "read", false)
	testDomainEnforce(t, e, "alice", "domain1", "data1", "write", false)
	testDomainEnforce(t, e, "alice", "domain1", "data2", "read", false)
	testDomainEnforce(t, e, "alice", "domain1", "data2", "write", false)
	testDomainEnforce(t, e, "bob", "domain2", "data1", "read", false)
	testDomainEnforce(t, e, "bob", "domain2", "data1", "write", false)
	testDomainEnforce(t, e, "bob", "domain2", "data2", "read", true)
	testDomainEnforce(t, e, "bob", "domain2", "data2", "write", true)

	// Remove the specified policy rule.
	_, _ = e.RemovePolicy("admin", "domain2", "data2", "read")

	testDomainEnforce(t, e, "alice", "domain1", "data1", "read", false)
	testDomainEnforce(t, e, "alice", "domain1", "data1", "write", false)
	testDomainEnforce(t, e, "alice", "domain1", "data2", "read", false)
	testDomainEnforce(t, e, "alice", "domain1", "data2", "write", false)
	testDomainEnforce(t, e, "bob", "domain2", "data1", "read", false)
	testDomainEnforce(t, e, "bob", "domain2", "data1", "write", false)
	testDomainEnforce(t, e, "bob", "domain2", "data2", "read", false)
	testDomainEnforce(t, e, "bob", "domain2", "data2", "write", true)
}

func TestRBACModelWithDomainsAtRuntimeMockAdapter(t *testing.T) {
	adapter := fileadapter.NewAdapterMock("examples/rbac_with_domains_policy.csv")
	e, _ := NewEnforcer("examples/rbac_with_domains_model.conf", adapter)

	_, _ = e.AddPolicy("admin", "domain3", "data1", "read")
	_, _ = e.AddGroupingPolicy("alice", "admin", "domain3")

	testDomainEnforce(t, e, "alice", "domain3", "data1", "read", true)

	testDomainEnforce(t, e, "alice", "domain1", "data1", "read", true)
	_, _ = e.RemoveFilteredPolicy(1, "domain1", "data1")
	testDomainEnforce(t, e, "alice", "domain1", "data1", "read", false)

	testDomainEnforce(t, e, "bob", "domain2", "data2", "read", true)
	_, _ = e.RemovePolicy("admin", "domain2", "data2", "read")
	testDomainEnforce(t, e, "bob", "domain2", "data2", "read", false)
}

func TestRBACModelWithDeny(t *testing.T) {
	e, _ := NewEnforcer("examples/rbac_with_deny_model.conf", "examples/rbac_with_deny_policy.csv")

	testEnforce(t, e, "alice", "data1", "read", true)
	testEnforce(t, e, "alice", "data1", "write", false)
	testEnforce(t, e, "alice", "data2", "read", true)
	testEnforce(t, e, "alice", "data2", "write", false)
	testEnforce(t, e, "bob", "data1", "read", false)
	testEnforce(t, e, "bob", "data1", "write", false)
	testEnforce(t, e, "bob", "data2", "read", false)
	testEnforce(t, e, "bob", "data2", "write", true)
}

func TestRBACModelWithOnlyDeny(t *testing.T) {
	e, _ := NewEnforcer("examples/rbac_with_not_deny_model.conf", "examples/rbac_with_deny_policy.csv")

	testEnforce(t, e, "alice", "data2", "write", false)
}

func TestRBACModelWithCustomData(t *testing.T) {
	e, _ := NewEnforcer("examples/rbac_model.conf", "examples/rbac_policy.csv")

	// You can add custom data to a grouping policy, Casbin will ignore it. It is only meaningful to the caller.
	// This feature can be used to store information like whether "bob" is an end user (so no subject will inherit "bob")
	// For Casbin, it is equivalent to: e.AddGroupingPolicy("bob", "data2_admin")
	_, _ = e.AddGroupingPolicy("bob", "data2_admin", "custom_data")

	testEnforce(t, e, "alice", "data1", "read", true)
	testEnforce(t, e, "alice", "data1", "write", false)
	testEnforce(t, e, "alice", "data2", "read", true)
	testEnforce(t, e, "alice", "data2", "write", true)
	testEnforce(t, e, "bob", "data1", "read", false)
	testEnforce(t, e, "bob", "data1", "write", false)
	testEnforce(t, e, "bob", "data2", "read", true)
	testEnforce(t, e, "bob", "data2", "write", true)

	// You should also take the custom data as a parameter when deleting a grouping policy.
	// e.RemoveGroupingPolicy("bob", "data2_admin") won't work.
	// Or you can remove it by using RemoveFilteredGroupingPolicy().
	_, _ = e.RemoveGroupingPolicy("bob", "data2_admin", "custom_data")

	testEnforce(t, e, "alice", "data1", "read", true)
	testEnforce(t, e, "alice", "data1", "write", false)
	testEnforce(t, e, "alice", "data2", "read", true)
	testEnforce(t, e, "alice", "data2", "write", true)
	testEnforce(t, e, "bob", "data1", "read", false)
	testEnforce(t, e, "bob", "data1", "write", false)
	testEnforce(t, e, "bob", "data2", "read", false)
	testEnforce(t, e, "bob", "data2", "write", true)
}

func TestRBACModelWithPattern(t *testing.T) {
	e, _ := NewEnforcer("examples/rbac_with_pattern_model.conf", "examples/rbac_with_pattern_policy.csv")

	// Here's a little confusing: the matching function here is not the custom function used in matcher.
	// It is the matching function used by "g" (and "g2", "g3" if any..)
	// You can see in policy that: "g2, /book/:id, book_group", so in "g2()" function in the matcher, instead
	// of checking whether "/book/:id" equals the obj: "/book/1", it checks whether the pattern matches.
	// You can see it as normal RBAC: "/book/:id" == "/book/1" becomes KeyMatch2("/book/:id", "/book/1")
	e.rm.(*defaultrolemanager.RoleManager).AddMatchingFunc("KeyMatch2", util.KeyMatch2)
	testEnforce(t, e, "alice", "/book/1", "GET", true)
	testEnforce(t, e, "alice", "/book/2", "GET", true)
	testEnforce(t, e, "alice", "/pen/1", "GET", true)
	testEnforce(t, e, "alice", "/pen/2", "GET", false)
	testEnforce(t, e, "bob", "/book/1", "GET", false)
	testEnforce(t, e, "bob", "/book/2", "GET", false)
	testEnforce(t, e, "bob", "/pen/1", "GET", true)
	testEnforce(t, e, "bob", "/pen/2", "GET", true)

	// AddMatchingFunc() is actually setting a function because only one function is allowed,
	// so when we set "KeyMatch3", we are actually replacing "KeyMatch2" with "KeyMatch3".
	e.rm.(*defaultrolemanager.RoleManager).AddMatchingFunc("KeyMatch3", util.KeyMatch3)
	testEnforce(t, e, "alice", "/book2/1", "GET", true)
	testEnforce(t, e, "alice", "/book2/2", "GET", true)
	testEnforce(t, e, "alice", "/pen2/1", "GET", true)
	testEnforce(t, e, "alice", "/pen2/2", "GET", false)
	testEnforce(t, e, "bob", "/book2/1", "GET", false)
	testEnforce(t, e, "bob", "/book2/2", "GET", false)
	testEnforce(t, e, "bob", "/pen2/1", "GET", true)
	testEnforce(t, e, "bob", "/pen2/2", "GET", true)
}

type testCustomRoleManager struct{}

func NewRoleManager() rbac.RoleManager {
	return &testCustomRoleManager{}
}
func (rm *testCustomRoleManager) Clear() error { return nil }
func (rm *testCustomRoleManager) AddLink(name1 string, name2 string, domain ...string) error {
	return nil
}
func (rm *testCustomRoleManager) DeleteLink(name1 string, name2 string, domain ...string) error {
	return nil
}
func (rm *testCustomRoleManager) HasLink(name1 string, name2 string, domain ...string) (bool, error) {
	if name1 == "alice" && name2 == "alice" {
		return true, nil
	} else if name1 == "alice" && name2 == "data2_admin" {
		return true, nil
	} else if name1 == "bob" && name2 == "bob" {
		return true, nil
	}
	return false, nil
}
func (rm *testCustomRoleManager) GetRoles(name string, domain ...string) ([]string, error) {
	return []string{}, nil
}
func (rm *testCustomRoleManager) GetUsers(name string, domain ...string) ([]string, error) {
	return []string{}, nil
}
func (rm *testCustomRoleManager) PrintRoles() error { return nil }

func TestRBACModelWithCustomRoleManager(t *testing.T) {
	e, _ := NewEnforcer("examples/rbac_model.conf", "examples/rbac_policy.csv")
	e.SetRoleManager(NewRoleManager())
	_ = e.LoadModel()
	_ = e.LoadPolicy()

	testEnforce(t, e, "alice", "data1", "read", true)
	testEnforce(t, e, "alice", "data1", "write", false)
	testEnforce(t, e, "alice", "data2", "read", true)
	testEnforce(t, e, "alice", "data2", "write", true)
	testEnforce(t, e, "bob", "data1", "read", false)
	testEnforce(t, e, "bob", "data1", "write", false)
	testEnforce(t, e, "bob", "data2", "read", false)
	testEnforce(t, e, "bob", "data2", "write", true)
}

type testResource struct {
	Name  string
	Owner string
}

func newTestResource(name string, owner string) testResource {
	r := testResource{}
	r.Name = name
	r.Owner = owner
	return r
}

func TestABACModel(t *testing.T) {
	e, _ := NewEnforcer("examples/abac_model.conf")

	data1 := newTestResource("data1", "alice")
	data2 := newTestResource("data2", "bob")

	testEnforce(t, e, "alice", data1, "read", true)
	testEnforce(t, e, "alice", data1, "write", true)
	testEnforce(t, e, "alice", data2, "read", false)
	testEnforce(t, e, "alice", data2, "write", false)
	testEnforce(t, e, "bob", data1, "read", false)
	testEnforce(t, e, "bob", data1, "write", false)
	testEnforce(t, e, "bob", data2, "read", true)
	testEnforce(t, e, "bob", data2, "write", true)
}

func TestKeyMatchModel(t *testing.T) {
	e, _ := NewEnforcer("examples/keymatch_model.conf", "examples/keymatch_policy.csv")

	testEnforce(t, e, "alice", "/alice_data/resource1", "GET", true)
	testEnforce(t, e, "alice", "/alice_data/resource1", "POST", true)
	testEnforce(t, e, "alice", "/alice_data/resource2", "GET", true)
	testEnforce(t, e, "alice", "/alice_data/resource2", "POST", false)
	testEnforce(t, e, "alice", "/bob_data/resource1", "GET", false)
	testEnforce(t, e, "alice", "/bob_data/resource1", "POST", false)
	testEnforce(t, e, "alice", "/bob_data/resource2", "GET", false)
	testEnforce(t, e, "alice", "/bob_data/resource2", "POST", false)

	testEnforce(t, e, "bob", "/alice_data/resource1", "GET", false)
	testEnforce(t, e, "bob", "/alice_data/resource1", "POST", false)
	testEnforce(t, e, "bob", "/alice_data/resource2", "GET", true)
	testEnforce(t, e, "bob", "/alice_data/resource2", "POST", false)
	testEnforce(t, e, "bob", "/bob_data/resource1", "GET", false)
	testEnforce(t, e, "bob", "/bob_data/resource1", "POST", true)
	testEnforce(t, e, "bob", "/bob_data/resource2", "GET", false)
	testEnforce(t, e, "bob", "/bob_data/resource2", "POST", true)

	testEnforce(t, e, "cathy", "/cathy_data", "GET", true)
	testEnforce(t, e, "cathy", "/cathy_data", "POST", true)
	testEnforce(t, e, "cathy", "/cathy_data", "DELETE", false)
}

func TestKeyMatch2Model(t *testing.T) {
	e, _ := NewEnforcer("examples/keymatch2_model.conf", "examples/keymatch2_policy.csv")

	testEnforce(t, e, "alice", "/alice_data", "GET", false)
	testEnforce(t, e, "alice", "/alice_data/resource1", "GET", true)
	testEnforce(t, e, "alice", "/alice_data2/myid", "GET", false)
	testEnforce(t, e, "alice", "/alice_data2/myid/using/res_id", "GET", true)
}

func CustomFunction(key1 string, key2 string) bool {
	if key1 == "/alice_data2/myid/using/res_id" && key2 == "/alice_data/:resource" {
		return true
	} else if key1 == "/alice_data2/myid/using/res_id" && key2 == "/alice_data2/:id/using/:resId" {
		return true
	} else {
		return false
	}
}

func CustomFunctionWrapper(args ...interface{}) (interface{}, error) {
	key1 := args[0].(string)
	key2 := args[1].(string)

	return bool(CustomFunction(key1, key2)), nil
}

func TestKeyMatchCustomModel(t *testing.T) {
	e, _ := NewEnforcer("examples/keymatch_custom_model.conf", "examples/keymatch2_policy.csv")

	e.AddFunction("keyMatchCustom", CustomFunctionWrapper)

	testEnforce(t, e, "alice", "/alice_data2/myid", "GET", false)
	testEnforce(t, e, "alice", "/alice_data2/myid/using/res_id", "GET", true)
}

func TestIPMatchModel(t *testing.T) {
	e, _ := NewEnforcer("examples/ipmatch_model.conf", "examples/ipmatch_policy.csv")

	testEnforce(t, e, "192.168.2.123", "data1", "read", true)
	testEnforce(t, e, "192.168.2.123", "data1", "write", false)
	testEnforce(t, e, "192.168.2.123", "data2", "read", false)
	testEnforce(t, e, "192.168.2.123", "data2", "write", false)

	testEnforce(t, e, "192.168.0.123", "data1", "read", false)
	testEnforce(t, e, "192.168.0.123", "data1", "write", false)
	testEnforce(t, e, "192.168.0.123", "data2", "read", false)
	testEnforce(t, e, "192.168.0.123", "data2", "write", false)

	testEnforce(t, e, "10.0.0.5", "data1", "read", false)
	testEnforce(t, e, "10.0.0.5", "data1", "write", false)
	testEnforce(t, e, "10.0.0.5", "data2", "read", false)
	testEnforce(t, e, "10.0.0.5", "data2", "write", true)

	testEnforce(t, e, "192.168.0.1", "data1", "read", false)
	testEnforce(t, e, "192.168.0.1", "data1", "write", false)
	testEnforce(t, e, "192.168.0.1", "data2", "read", false)
	testEnforce(t, e, "192.168.0.1", "data2", "write", false)
}

func TestGlobMatchModel(t *testing.T) {
	e, _ := NewEnforcer("examples/glob_model.conf", "examples/glob_policy.csv")
	testEnforce(t, e, "u1", "/foo/", "read", true)
	testEnforce(t, e, "u1", "/foo", "read", false)
	testEnforce(t, e, "u1", "/foo/subprefix", "read", true)
	testEnforce(t, e, "u1", "foo", "read", false)

	testEnforce(t, e, "u2", "/foosubprefix", "read", true)
	testEnforce(t, e, "u2", "/foo/subprefix", "read", false)
	testEnforce(t, e, "u2", "foo", "read", false)

	testEnforce(t, e, "u3", "/prefix/foo/subprefix", "read", true)
	testEnforce(t, e, "u3", "/prefix/foo/", "read", true)
	testEnforce(t, e, "u3", "/prefix/foo", "read", false)

	testEnforce(t, e, "u4", "/foo", "read", false)
	testEnforce(t, e, "u4", "foo", "read", true)
}

func TestPriorityModel(t *testing.T) {
	e, _ := NewEnforcer("examples/priority_model.conf", "examples/priority_policy.csv")

	testEnforce(t, e, "alice", "data1", "read", true)
	testEnforce(t, e, "alice", "data1", "write", false)
	testEnforce(t, e, "alice", "data2", "read", false)
	testEnforce(t, e, "alice", "data2", "write", false)
	testEnforce(t, e, "bob", "data1", "read", false)
	testEnforce(t, e, "bob", "data1", "write", false)
	testEnforce(t, e, "bob", "data2", "read", true)
	testEnforce(t, e, "bob", "data2", "write", false)
}

func TestPriorityModelIndeterminate(t *testing.T) {
	e, _ := NewEnforcer("examples/priority_model.conf", "examples/priority_indeterminate_policy.csv")

	testEnforce(t, e, "alice", "data1", "read", false)
}

func TestRBACModelInMultiLines(t *testing.T) {
	e, _ := NewEnforcer("examples/rbac_model_in_multi_line.conf", "examples/rbac_policy.csv")

	testEnforce(t, e, "alice", "data1", "read", true)
	testEnforce(t, e, "alice", "data1", "write", false)
	testEnforce(t, e, "alice", "data2", "read", true)
	testEnforce(t, e, "alice", "data2", "write", true)
	testEnforce(t, e, "bob", "data1", "read", false)
	testEnforce(t, e, "bob", "data1", "write", false)
	testEnforce(t, e, "bob", "data2", "read", false)
	testEnforce(t, e, "bob", "data2", "write", true)
}

type testSub struct {
	Name string
	Age  int
}

func newTestSubject(name string, age int) testSub {
	s := testSub{}
	s.Name = name
	s.Age = age
	return s
}

func TestABACPolicy(t *testing.T) {
	e, _ := NewEnforcer("examples/abac_rule_model.conf", "examples/abac_rule_policy.csv")
	sub1 := newTestSubject("alice", 16)
	sub2 := newTestSubject("alice", 20)
	sub3 := newTestSubject("alice", 65)

	testEnforce(t, e, sub1, "/data1", "read", false)
	testEnforce(t, e, sub1, "/data2", "read", false)
	testEnforce(t, e, sub1, "/data1", "write", false)
	testEnforce(t, e, sub1, "/data2", "write", true)
	testEnforce(t, e, sub2, "/data1", "read", true)
	testEnforce(t, e, sub2, "/data2", "read", false)
	testEnforce(t, e, sub2, "/data1", "write", false)
	testEnforce(t, e, sub2, "/data2", "write", true)
	testEnforce(t, e, sub3, "/data1", "read", true)
	testEnforce(t, e, sub3, "/data2", "read", false)
	testEnforce(t, e, sub3, "/data1", "write", false)
	testEnforce(t, e, sub3, "/data2", "write", false)
}

func TestCommentModel(t *testing.T) {
	e, _ := NewEnforcer("examples/comment_model.conf", "examples/basic_policy.csv")
	testEnforce(t, e, "alice", "data1", "read", true)
	testEnforce(t, e, "alice", "data1", "write", false)
	testEnforce(t, e, "alice", "data2", "read", false)
	testEnforce(t, e, "alice", "data2", "write", false)
	testEnforce(t, e, "bob", "data1", "read", false)
	testEnforce(t, e, "bob", "data1", "write", false)
	testEnforce(t, e, "bob", "data2", "read", false)
	testEnforce(t, e, "bob", "data2", "write", true)
}

func TestDomainMatchModel(t *testing.T) {
	e, _ := NewEnforcer("examples/rbac_with_domain_pattern_model.conf", "examples/rbac_with_domain_pattern_policy.csv")
	e.rm.(*defaultrolemanager.RoleManager).AddDomainMatchingFunc("keyMatch2", util.KeyMatch2)

	testDomainEnforce(t, e, "alice", "domain1", "data1", "read", true)
	testDomainEnforce(t, e, "alice", "domain1", "data1", "write", true)
	testDomainEnforce(t, e, "alice", "domain1", "data2", "read", false)
	testDomainEnforce(t, e, "alice", "domain1", "data2", "write", false)
	testDomainEnforce(t, e, "alice", "domain2", "data2", "read", true)
	testDomainEnforce(t, e, "alice", "domain2", "data2", "write", true)
	testDomainEnforce(t, e, "bob", "domain2", "data1", "read", false)
	testDomainEnforce(t, e, "bob", "domain2", "data1", "write", false)
	testDomainEnforce(t, e, "bob", "domain2", "data2", "read", true)
	testDomainEnforce(t, e, "bob", "domain2", "data2", "write", true)
}

func TestAllMatchModel(t *testing.T) {
	e, _ := NewEnforcer("examples/rbac_with_all_pattern_model.conf", "examples/rbac_with_all_pattern_policy.csv")
	e.rm.(*defaultrolemanager.RoleManager).AddMatchingFunc("keyMatch2", util.KeyMatch2)
	e.rm.(*defaultrolemanager.RoleManager).AddDomainMatchingFunc("keyMatch2", util.KeyMatch2)

	testDomainEnforce(t, e, "alice", "domain1", "/book/1", "read", true)
	testDomainEnforce(t, e, "alice", "domain1", "/book/1", "write", false)
	testDomainEnforce(t, e, "alice", "domain2", "/book/1", "read", false)
	testDomainEnforce(t, e, "alice", "domain2", "/book/1", "write", true)
}
