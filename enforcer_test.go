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
	"sync"
	"testing"

	"github.com/casbin/casbin/v3/model"
	fileadapter "github.com/casbin/casbin/v3/persist/file-adapter"
	"github.com/casbin/casbin/v3/util"
)

func TestKeyMatchModelInMemory(t *testing.T) {
	m := model.NewModel()
	m.AddDef("r", "r", "sub, obj, act")
	m.AddDef("p", "p", "sub, obj, act")
	m.AddDef("e", "e", "some(where (p.eft == allow))")
	m.AddDef("m", "m", "r.sub == p.sub && keyMatch(r.obj, p.obj) && regexMatch(r.act, p.act)")

	a := fileadapter.NewAdapter("examples/keymatch_policy.csv")

	e, _ := NewEnforcer(m, a)

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

	e, _ = NewEnforcer(m)
	_ = a.LoadPolicy(e.GetModel())

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

func TestKeyMatchModelInMemoryDeny(t *testing.T) {
	m := model.NewModel()
	m.AddDef("r", "r", "sub, obj, act")
	m.AddDef("p", "p", "sub, obj, act")
	m.AddDef("e", "e", "!some(where (p.eft == deny))")
	m.AddDef("m", "m", "r.sub == p.sub && keyMatch(r.obj, p.obj) && regexMatch(r.act, p.act)")

	a := fileadapter.NewAdapter("examples/keymatch_policy.csv")

	e, _ := NewEnforcer(m, a)

	testEnforce(t, e, "alice", "/alice_data/resource2", "POST", true)
}

func TestRBACModelInMemoryIndeterminate(t *testing.T) {
	m := model.NewModel()
	m.AddDef("r", "r", "sub, obj, act")
	m.AddDef("p", "p", "sub, obj, act")
	m.AddDef("g", "g", "_, _")
	m.AddDef("e", "e", "some(where (p.eft == allow))")
	m.AddDef("m", "m", "g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act")

	e, _ := NewEnforcer(m)

	_, _ = e.AddPermissionForUser("alice", "data1", "invalid")

	testEnforce(t, e, "alice", "data1", "read", false)
}

func TestRBACModelInMemory(t *testing.T) {
	m := model.NewModel()
	m.AddDef("r", "r", "sub, obj, act")
	m.AddDef("p", "p", "sub, obj, act")
	m.AddDef("g", "g", "_, _")
	m.AddDef("e", "e", "some(where (p.eft == allow))")
	m.AddDef("m", "m", "g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act")

	e, _ := NewEnforcer(m)

	_, _ = e.AddPermissionForUser("alice", "data1", "read")
	_, _ = e.AddPermissionForUser("bob", "data2", "write")
	_, _ = e.AddPermissionForUser("data2_admin", "data2", "read")
	_, _ = e.AddPermissionForUser("data2_admin", "data2", "write")
	_, _ = e.AddRoleForUser("alice", "data2_admin")

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
	m, _ := model.NewModelFromString(text)
	// The above is the same as:
	// m := NewModel()
	// m.LoadModelFromText(text)

	e, _ := NewEnforcer(m)

	_, _ = e.AddPermissionForUser("alice", "data1", "read")
	_, _ = e.AddPermissionForUser("bob", "data2", "write")
	_, _ = e.AddPermissionForUser("data2_admin", "data2", "read")
	_, _ = e.AddPermissionForUser("data2_admin", "data2", "write")
	_, _ = e.AddRoleForUser("alice", "data2_admin")

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
	m := model.NewModel()
	m.AddDef("r", "r", "sub, obj, act")
	m.AddDef("p", "p", "sub, obj, act")
	m.AddDef("g", "g", "_, _")
	m.AddDef("e", "e", "some(where (p.eft == allow))")
	m.AddDef("m", "m", "g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act")

	e, _ := NewEnforcer(m)

	_, _ = e.AddPermissionForUser("alice", "data1", "read")
	_, _ = e.AddPermissionForUser("bob", "data2", "write")

	testEnforce(t, e, "alice", "data1", "read", true)
	testEnforce(t, e, "alice", "data1", "write", false)
	testEnforce(t, e, "alice", "data2", "read", false)
	testEnforce(t, e, "alice", "data2", "write", false)
	testEnforce(t, e, "bob", "data1", "read", false)
	testEnforce(t, e, "bob", "data1", "write", false)
	testEnforce(t, e, "bob", "data2", "read", false)
	testEnforce(t, e, "bob", "data2", "write", true)
}

func TestMatcherUsingInOperator(t *testing.T) {
	// From file config
	e, _ := NewEnforcer("examples/rbac_model_matcher_using_in_op.conf")
	_, _ = e.AddPermissionForUser("alice", "data1", "read")

	testEnforce(t, e, "alice", "data1", "read", true)
	testEnforce(t, e, "alice", "data2", "read", true)
	testEnforce(t, e, "alice", "data3", "read", true)
	testEnforce(t, e, "anyone", "data1", "read", false)
	testEnforce(t, e, "anyone", "data2", "read", true)
	testEnforce(t, e, "anyone", "data3", "read", true)
}

func TestReloadPolicy(t *testing.T) {
	e, _ := NewEnforcer("examples/rbac_model.conf", "examples/rbac_policy.csv")

	_ = e.LoadPolicy()
	testGetPolicy(t, e, [][]string{{"alice", "data1", "read"}, {"bob", "data2", "write"}, {"data2_admin", "data2", "read"}, {"data2_admin", "data2", "write"}})
}

func TestSavePolicy(t *testing.T) {
	e, _ := NewEnforcer("examples/rbac_model.conf", "examples/rbac_policy.csv")

	_ = e.SavePolicy()
}

func TestClearPolicy(t *testing.T) {
	e, _ := NewEnforcer("examples/rbac_model.conf", "examples/rbac_policy.csv")

	e.ClearPolicy()
}

func TestEnableEnforce(t *testing.T) {
	e, _ := NewEnforcer("examples/basic_model.conf", "examples/basic_policy.csv")

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
	e, _ := NewEnforcer("examples/basic_model.conf", "examples/basic_policy.csv", true)
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
	e, _ := NewEnforcer("examples/basic_model.conf", "examples/basic_policy.csv")

	e.EnableAutoSave(false)
	// Because AutoSave is disabled, the policy change only affects the policy in Casbin enforcer,
	// it doesn't affect the policy in the storage.
	_, _ = e.RemovePolicy("alice", "data1", "read")
	// Reload the policy from the storage to see the effect.
	_ = e.LoadPolicy()
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
	_, _ = e.RemovePolicy("alice", "data1", "read")

	// However, the file adapter doesn't implement the AutoSave feature, so enabling it has no effect at all here.

	// Reload the policy from the storage to see the effect.
	_ = e.LoadPolicy()
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
	e, _ := NewEnforcer("examples/basic_model.conf", adapter)

	testEnforce(t, e, "alice", "data1", "read", true)
	testEnforce(t, e, "alice", "data1", "write", false)
	testEnforce(t, e, "alice", "data2", "read", false)
	testEnforce(t, e, "alice", "data2", "write", false)
	testEnforce(t, e, "bob", "data1", "read", false)
	testEnforce(t, e, "bob", "data1", "write", false)
	testEnforce(t, e, "bob", "data2", "read", false)
	testEnforce(t, e, "bob", "data2", "write", true)
}

func TestRoleLinks(t *testing.T) {
	e, _ := NewEnforcer("examples/rbac_model.conf")
	e.EnableAutoBuildRoleLinks(false)
	_ = e.BuildRoleLinks()
	_, _ = e.Enforce("user501", "data9", "read")
}

func TestEnforceConcurrency(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Enforce is not concurrent")
		}
	}()

	e, _ := NewEnforcer("examples/rbac_model.conf")
	_ = e.LoadModel()

	var wg sync.WaitGroup

	// Simulate concurrency (maybe use a timer?)
	for i := 1; i <= 10000; i++ {
		wg.Add(1)
		go func() {
			_, _ = e.Enforce("user501", "data9", "read")
			wg.Done()
		}()
	}

	wg.Wait()
}

func TestGetAndSetModel(t *testing.T) {
	e, _ := NewEnforcer("examples/basic_model.conf", "examples/basic_policy.csv")
	e2, _ := NewEnforcer("examples/basic_with_root_model.conf", "examples/basic_policy.csv")

	testEnforce(t, e, "root", "data1", "read", false)

	e.SetModel(e2.GetModel())

	testEnforce(t, e, "root", "data1", "read", true)
}

func TestGetAndSetAdapterInMem(t *testing.T) {
	e, _ := NewEnforcer("examples/basic_model.conf", "examples/basic_policy.csv")
	e2, _ := NewEnforcer("examples/basic_model.conf", "examples/basic_inverse_policy.csv")

	testEnforce(t, e, "alice", "data1", "read", true)
	testEnforce(t, e, "alice", "data1", "write", false)

	a2 := e2.GetAdapter()
	e.SetAdapter(a2)
	_ = e.LoadPolicy()

	testEnforce(t, e, "alice", "data1", "read", false)
	testEnforce(t, e, "alice", "data1", "write", true)
}

func TestSetAdapterFromFile(t *testing.T) {
	e, _ := NewEnforcer("examples/basic_model.conf")

	testEnforce(t, e, "alice", "data1", "read", false)

	a := fileadapter.NewAdapter("examples/basic_policy.csv")
	e.SetAdapter(a)
	_ = e.LoadPolicy()

	testEnforce(t, e, "alice", "data1", "read", true)
}

func TestInitEmpty(t *testing.T) {
	e, _ := NewEnforcer()

	m := model.NewModel()
	m.AddDef("r", "r", "sub, obj, act")
	m.AddDef("p", "p", "sub, obj, act")
	m.AddDef("e", "e", "some(where (p.eft == allow))")
	m.AddDef("m", "m", "r.sub == p.sub && keyMatch(r.obj, p.obj) && regexMatch(r.act, p.act)")

	a := fileadapter.NewAdapter("examples/keymatch_policy.csv")

	e.SetModel(m)
	e.SetAdapter(a)
	_ = e.LoadPolicy()

	testEnforce(t, e, "alice", "/alice_data/resource1", "GET", true)
}
func testEnforceEx(t *testing.T, e *Enforcer, sub string, obj string, act string, res []string) {
	t.Helper()
	_, myRes, _ := e.EnforceEx(sub, obj, act)

	if ok := util.ArrayEquals(res, myRes); !ok {
		t.Error("Key: ", myRes, ", supposed to be ", res)
	}
}

func TestEnforceEx(t *testing.T) {
	e, _ := NewEnforcer("examples/basic_model.conf", "examples/basic_policy.csv")

	testEnforceEx(t, e, "alice", "data1", "read", []string{"alice", "data1", "read"})
	testEnforceEx(t, e, "alice", "data1", "write", []string{})
	testEnforceEx(t, e, "alice", "data2", "read", []string{})
	testEnforceEx(t, e, "alice", "data2", "write", []string{})
	testEnforceEx(t, e, "bob", "data1", "read", []string{})
	testEnforceEx(t, e, "bob", "data1", "write", []string{})
	testEnforceEx(t, e, "bob", "data2", "read", []string{})
	testEnforceEx(t, e, "bob", "data2", "write", []string{"bob", "data2", "write"})

	e, _ = NewEnforcer("examples/rbac_model.conf", "examples/rbac_policy.csv")

	testEnforceEx(t, e, "alice", "data1", "read", []string{"alice", "data1", "read"})
	testEnforceEx(t, e, "alice", "data1", "write", []string{})
	testEnforceEx(t, e, "alice", "data2", "read", []string{"data2_admin", "data2", "read"})
	testEnforceEx(t, e, "alice", "data2", "write", []string{"data2_admin", "data2", "write"})
	testEnforceEx(t, e, "bob", "data1", "read", []string{})
	testEnforceEx(t, e, "bob", "data1", "write", []string{})
	testEnforceEx(t, e, "bob", "data2", "read", []string{})
	testEnforceEx(t, e, "bob", "data2", "write", []string{"bob", "data2", "write"})

	e, _ = NewEnforcer("examples/priority_model.conf", "examples/priority_policy.csv")

	testEnforceEx(t, e, "alice", "data1", "read", []string{"alice", "data1", "read", "allow"})
	testEnforceEx(t, e, "alice", "data1", "write", []string{"data1_deny_group", "data1", "write", "deny"})
	testEnforceEx(t, e, "alice", "data2", "read", []string{})
	testEnforceEx(t, e, "alice", "data2", "write", []string{})
	testEnforceEx(t, e, "bob", "data1", "write", []string{})
	testEnforceEx(t, e, "bob", "data2", "read", []string{"data2_allow_group", "data2", "read", "allow"})
	testEnforceEx(t, e, "bob", "data2", "write", []string{"bob", "data2", "write", "deny"})
}

func TestEnforceExLog(t *testing.T) {
	e, _ := NewEnforcer("examples/basic_model.conf", "examples/basic_policy.csv", true)

	testEnforceEx(t, e, "alice", "data1", "read", []string{"alice", "data1", "read"})
	testEnforceEx(t, e, "alice", "data1", "write", []string{})
	testEnforceEx(t, e, "alice", "data2", "read", []string{})
	testEnforceEx(t, e, "alice", "data2", "write", []string{})
	testEnforceEx(t, e, "bob", "data1", "read", []string{})
	testEnforceEx(t, e, "bob", "data1", "write", []string{})
	testEnforceEx(t, e, "bob", "data2", "read", []string{})
	testEnforceEx(t, e, "bob", "data2", "write", []string{"bob", "data2", "write"})
}
