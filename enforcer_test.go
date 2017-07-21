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

	"github.com/casbin/casbin/file-adapter"
)

func TestGetAndSetModel(t *testing.T) {
	e := NewEnforcer("examples/basic_model.conf", "examples/basic_policy.csv")

	model := e.GetModel()
	e.SetModel(model)
}

func TestKeyMatchModelInMemory(t *testing.T) {
	m := NewModel()
	m.AddDef("r", "r", "sub, obj, act")
	m.AddDef("p", "p", "sub, obj, act")
	m.AddDef("e", "e", "some(where (p.eft == allow))")
	m.AddDef("m", "m", "r.sub == p.sub && keyMatch(r.obj, p.obj) && regexMatch(r.act, p.act)")

	a := fileadapter.NewAdapter("examples/keymatch_policy.csv")

	e := NewEnforcer(m, a)

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

	e = NewEnforcer(m)
	a.LoadPolicy(e.GetModel())

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

func TestRBACModelInMemory(t *testing.T) {
	m := NewModel()
	m.AddDef("r", "r", "sub, obj, act")
	m.AddDef("p", "p", "sub, obj, act")
	m.AddDef("g", "g", "_, _")
	m.AddDef("e", "e", "some(where (p.eft == allow))")
	m.AddDef("m", "m", "g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act")

	e := NewEnforcer(m)

	e.AddPermissionForUser("alice", "data1", "read")
	e.AddPermissionForUser("bob", "data2", "write")
	e.AddPermissionForUser("data2_admin", "data2", "read")
	e.AddPermissionForUser("data2_admin", "data2", "write")
	e.AddRoleForUser("alice", "data2_admin")

	testEnforce(t, e, "alice", "data1", "read", true)
	testEnforce(t, e, "alice", "data1", "write", false)
	testEnforce(t, e, "alice", "data2", "read", true)
	testEnforce(t, e, "alice", "data2", "write", true)
	testEnforce(t, e, "bob", "data1", "read", false)
	testEnforce(t, e, "bob", "data1", "write", false)
	testEnforce(t, e, "bob", "data2", "read", false)
	testEnforce(t, e, "bob", "data2", "write", true)
}

func TestRBACModelInMemory2(t *testing.T) {
	text :=
		`
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act
`
	m := NewModel(text)
	// The above is the same as:
	// m := NewModel()
	// m.LoadModelFromText(text)

	e := NewEnforcer(m)

	e.AddPermissionForUser("alice", "data1", "read")
	e.AddPermissionForUser("bob", "data2", "write")
	e.AddPermissionForUser("data2_admin", "data2", "read")
	e.AddPermissionForUser("data2_admin", "data2", "write")
	e.AddRoleForUser("alice", "data2_admin")

	testEnforce(t, e, "alice", "data1", "read", true)
	testEnforce(t, e, "alice", "data1", "write", false)
	testEnforce(t, e, "alice", "data2", "read", true)
	testEnforce(t, e, "alice", "data2", "write", true)
	testEnforce(t, e, "bob", "data1", "read", false)
	testEnforce(t, e, "bob", "data1", "write", false)
	testEnforce(t, e, "bob", "data2", "read", false)
	testEnforce(t, e, "bob", "data2", "write", true)
}

func TestNotUsedRBACModelInMemory(t *testing.T) {
	m := NewModel()
	m.AddDef("r", "r", "sub, obj, act")
	m.AddDef("p", "p", "sub, obj, act")
	m.AddDef("g", "g", "_, _")
	m.AddDef("e", "e", "some(where (p.eft == allow))")
	m.AddDef("m", "m", "g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act")

	e := NewEnforcer(m)

	e.AddPermissionForUser("alice", "data1", "read")
	e.AddPermissionForUser("bob", "data2", "write")

	testEnforce(t, e, "alice", "data1", "read", true)
	testEnforce(t, e, "alice", "data1", "write", false)
	testEnforce(t, e, "alice", "data2", "read", false)
	testEnforce(t, e, "alice", "data2", "write", false)
	testEnforce(t, e, "bob", "data1", "read", false)
	testEnforce(t, e, "bob", "data1", "write", false)
	testEnforce(t, e, "bob", "data2", "read", false)
	testEnforce(t, e, "bob", "data2", "write", true)
}

func TestReloadPolicy(t *testing.T) {
	e := NewEnforcer("examples/rbac_model.conf", "examples/rbac_policy.csv")

	e.LoadPolicy()
	testGetPolicy(t, e, [][]string{{"alice", "data1", "read"}, {"bob", "data2", "write"}, {"data2_admin", "data2", "read"}, {"data2_admin", "data2", "write"}})
}

func TestSavePolicy(t *testing.T) {
	e := NewEnforcer("examples/rbac_model.conf", "examples/rbac_policy.csv")

	e.SavePolicy()
}

func TestEnableEnforce(t *testing.T) {
	e := NewEnforcer("examples/basic_model.conf", "examples/basic_policy.csv")

	e.EnableEnforce(false)
	testEnforce(t, e, "alice", "data1", "read", true)
	testEnforce(t, e, "alice", "data1", "write", true)
	testEnforce(t, e, "alice", "data2", "read", true)
	testEnforce(t, e, "alice", "data2", "write", true)
	testEnforce(t, e, "bob", "data1", "read", true)
	testEnforce(t, e, "bob", "data1", "write", true)
	testEnforce(t, e, "bob", "data2", "read", true)
	testEnforce(t, e, "bob", "data2", "write", true)

	e.EnableEnforce(true)
	testEnforce(t, e, "alice", "data1", "read", true)
	testEnforce(t, e, "alice", "data1", "write", false)
	testEnforce(t, e, "alice", "data2", "read", false)
	testEnforce(t, e, "alice", "data2", "write", false)
	testEnforce(t, e, "bob", "data1", "read", false)
	testEnforce(t, e, "bob", "data1", "write", false)
	testEnforce(t, e, "bob", "data2", "read", false)
	testEnforce(t, e, "bob", "data2", "write", true)
}

func TestEnableLog(t *testing.T) {
	e := NewEnforcer("examples/basic_model.conf", "examples/basic_policy.csv", true)
	// The log is enabled by default, so the above is the same with:
	// e := NewEnforcer("examples/basic_model.conf", "examples/basic_policy.csv")

	testEnforce(t, e, "alice", "data1", "read", true)
	testEnforce(t, e, "alice", "data1", "write", false)
	testEnforce(t, e, "alice", "data2", "read", false)
	testEnforce(t, e, "alice", "data2", "write", false)
	testEnforce(t, e, "bob", "data1", "read", false)
	testEnforce(t, e, "bob", "data1", "write", false)
	testEnforce(t, e, "bob", "data2", "read", false)
	testEnforce(t, e, "bob", "data2", "write", true)

	// The log can also be enabled or disabled at run-time.
	e.EnableLog(false)
	testEnforce(t, e, "alice", "data1", "read", true)
	testEnforce(t, e, "alice", "data1", "write", false)
	testEnforce(t, e, "alice", "data2", "read", false)
	testEnforce(t, e, "alice", "data2", "write", false)
	testEnforce(t, e, "bob", "data1", "read", false)
	testEnforce(t, e, "bob", "data1", "write", false)
	testEnforce(t, e, "bob", "data2", "read", false)
	testEnforce(t, e, "bob", "data2", "write", true)
}

func TestEnableAutoSave(t *testing.T) {
	e := NewEnforcer("examples/basic_model.conf", "examples/basic_policy.csv")

	e.EnableAutoSave(false)
	// Because AutoSave is disabled, the policy change only affects the policy in Casbin enforcer,
	// it doesn't affect the policy in the storage.
	e.RemovePolicy("alice", "data1", "read")
	// Reload the policy from the storage to see the effect.
	e.LoadPolicy()
	testEnforce(t, e, "alice", "data1", "read", true)
	testEnforce(t, e, "alice", "data1", "write", false)
	testEnforce(t, e, "alice", "data2", "read", false)
	testEnforce(t, e, "alice", "data2", "write", false)
	testEnforce(t, e, "bob", "data1", "read", false)
	testEnforce(t, e, "bob", "data1", "write", false)
	testEnforce(t, e, "bob", "data2", "read", false)
	testEnforce(t, e, "bob", "data2", "write", true)

	e.EnableAutoSave(true)
	// Because AutoSave is enabled, the policy change not only affects the policy in Casbin enforcer,
	// but also affects the policy in the storage.
	e.RemovePolicy("alice", "data1", "read")

	// However, the file adapter doesn't implement the AutoSave feature, so enabling it has no effect at all here.

	// Reload the policy from the storage to see the effect.
	e.LoadPolicy()
	testEnforce(t, e, "alice", "data1", "read", true) // Will not be false here.
	testEnforce(t, e, "alice", "data1", "write", false)
	testEnforce(t, e, "alice", "data2", "read", false)
	testEnforce(t, e, "alice", "data2", "write", false)
	testEnforce(t, e, "bob", "data1", "read", false)
	testEnforce(t, e, "bob", "data1", "write", false)
	testEnforce(t, e, "bob", "data2", "read", false)
	testEnforce(t, e, "bob", "data2", "write", true)
}

func TestInitWithAdapter(t *testing.T) {
	adapter := fileadapter.NewAdapter("examples/basic_policy.csv")
	e := NewEnforcer("examples/basic_model.conf", adapter)

	testEnforce(t, e, "alice", "data1", "read", true)
	testEnforce(t, e, "alice", "data1", "write", false)
	testEnforce(t, e, "alice", "data2", "read", false)
	testEnforce(t, e, "alice", "data2", "write", false)
	testEnforce(t, e, "bob", "data1", "read", false)
	testEnforce(t, e, "bob", "data1", "write", false)
	testEnforce(t, e, "bob", "data2", "read", false)
	testEnforce(t, e, "bob", "data2", "write", true)
}
