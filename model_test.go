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
	"reflect"
	"testing"
)

func testEnforce(t *testing.T, e *Enforcer, sub string, obj string, act string, res bool) {
	if e.Enforce(sub, obj, act) != res {
		t.Errorf("%s, %s, %s: %t, supposed to be %t", sub, obj, act, !res, res)
	}
}

func testEnforceWithoutUsers(t *testing.T, e *Enforcer, obj string, act string, res bool) {
	if e.Enforce(obj, act) != res {
		t.Errorf("%s, %s: %t, supposed to be %t", obj, act, !res, res)
	}
}

func testDomainEnforce(t *testing.T, e *Enforcer, sub string, dom string, obj string, act string, res bool) {
	if e.Enforce(sub, dom, obj, act) != res {
		t.Errorf("%s, %s, %s, %s: %t, supposed to be %t", sub, dom, obj, act, !res, res)
	}
}

func TestBasicModel(t *testing.T) {
	e := NewEnforcer("examples/basic_model.conf", "examples/basic_policy.csv")

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
	e := NewEnforcer("examples/basic_model.conf", "")

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
	e := NewEnforcer("examples/basic_model_with_root.conf", "examples/basic_policy.csv")

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
	e := NewEnforcer("examples/basic_model_with_root.conf", "")

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
	e := NewEnforcer("examples/basic_model_without_users.conf", "examples/basic_policy_without_users.csv")

	testEnforceWithoutUsers(t, e, "data1", "read", true)
	testEnforceWithoutUsers(t, e, "data1", "write", false)
	testEnforceWithoutUsers(t, e, "data2", "read", false)
	testEnforceWithoutUsers(t, e, "data2", "write", true)
}

func TestBasicModelWithoutResources(t *testing.T) {
	e := NewEnforcer("examples/basic_model_without_resources.conf", "examples/basic_policy_without_resources.csv")

	testEnforceWithoutUsers(t, e, "alice", "read", true)
	testEnforceWithoutUsers(t, e, "alice", "write", false)
	testEnforceWithoutUsers(t, e, "bob", "read", false)
	testEnforceWithoutUsers(t, e, "bob", "write", true)
}

func TestRBACModel(t *testing.T) {
	e := NewEnforcer("examples/rbac_model.conf", "examples/rbac_policy.csv")

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
	e := NewEnforcer("examples/rbac_model_with_resource_roles.conf", "examples/rbac_policy_with_resource_roles.csv")

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
	e := NewEnforcer("examples/rbac_model_with_domains.conf", "examples/rbac_policy_with_domains.csv")

	testDomainEnforce(t, e, "alice", "domain1", "data1", "read", true)
	testDomainEnforce(t, e, "alice", "domain1", "data1", "write", true)
	testDomainEnforce(t, e, "alice", "domain1", "data2", "read", false)
	testDomainEnforce(t, e, "alice", "domain1", "data2", "write", false)
	testDomainEnforce(t, e, "bob", "domain2", "data1", "read", false)
	testDomainEnforce(t, e, "bob", "domain2", "data1", "write", false)
	testDomainEnforce(t, e, "bob", "domain2", "data2", "read", true)
	testDomainEnforce(t, e, "bob", "domain2", "data2", "write", true)
}

func TestRBACModelWithDeny(t *testing.T) {
	e := NewEnforcer("examples/rbac_model_with_deny.conf", "examples/rbac_policy_with_deny.csv")

	testEnforce(t, e, "alice", "data1", "read", true)
	testEnforce(t, e, "alice", "data1", "write", false)
	testEnforce(t, e, "alice", "data2", "read", true)
	testEnforce(t, e, "alice", "data2", "write", false)
	testEnforce(t, e, "bob", "data1", "read", false)
	testEnforce(t, e, "bob", "data1", "write", false)
	testEnforce(t, e, "bob", "data2", "read", false)
	testEnforce(t, e, "bob", "data2", "write", true)
}

func getAttr(name string, attr string) string {
	// This is the same as:
	//
	// alice.domain = domain1
	// bob.domain = domain2
	// data1.domain = domain1
	// data2.domain = domain2

	if attr != "domain" {
		return "unknown"
	}

	if name == "alice" || name == "data1" {
		return "domain1"
	} else if name == "bob" || name == "data2" {
		return "domain2"
	} else {
		return "unknown"
	}
}

func getAttrFunc(args ...interface{}) (interface{}, error) {
	name := args[0].(string)
	attr := args[1].(string)

	return (string)(getAttr(name, attr)), nil
}

func TestABACModel(t *testing.T) {
	e := NewEnforcer("examples/abac_model.conf", "")

	e.AddSubjectAttributeFunction(getAttrFunc)
	e.AddObjectAttributeFunction(getAttrFunc)

	testEnforce(t, e, "alice", "data1", "read", true)
	testEnforce(t, e, "alice", "data1", "write", true)
	testEnforce(t, e, "alice", "data2", "read", false)
	testEnforce(t, e, "alice", "data2", "write", false)
	testEnforce(t, e, "bob", "data1", "read", false)
	testEnforce(t, e, "bob", "data1", "write", false)
	testEnforce(t, e, "bob", "data2", "read", true)
	testEnforce(t, e, "bob", "data2", "write", true)
}

type testUser struct {
	name   string
	domain string
}

func newTestUser(name string, domain string) *testUser {
	u := testUser{}
	u.name = name
	u.domain = domain
	return &u
}

func (u *testUser) getAttribute(attributeName string) string {
	ru := reflect.ValueOf(u)
	f := reflect.Indirect(ru).FieldByName(attributeName)
	return f.String()
}

type testResource struct {
	name   string
	domain string
}

func newTestResource(name string, domain string) *testResource {
	r := testResource{}
	r.name = name
	r.domain = domain
	return &r
}

func (u *testResource) getAttribute(attributeName string) string {
	ru := reflect.ValueOf(u)
	f := reflect.Indirect(ru).FieldByName(attributeName)
	return f.String()
}

//func TestABACModel2(t *testing.T) {
//	e := NewEnforcer("examples/abac_model.conf", "")
//
//	alice := newTestUser("alice", "domain1")
//	bob := newTestUser("bob", "domain2")
//	data1 := newTestResource("data1", "domain1")
//	data2 := newTestResource("data2", "domain2")
//
//	log.Println(alice.getAttribute("domain"))
//	log.Println(bob.getAttribute("domain"))
//	log.Println(data1.getAttribute("domain"))
//	log.Println(data2.getAttribute("domain"))
//}

func TestKeymatchModel(t *testing.T) {
	e := NewEnforcer("examples/keymatch_model.conf", "examples/keymatch_policy.csv")

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
