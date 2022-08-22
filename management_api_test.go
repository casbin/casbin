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

	"github.com/casbin/casbin/v2/util"
)

func testStringList(t *testing.T, title string, f func() []string, res []string) {
	t.Helper()
	myRes := f()
	t.Log(title+": ", myRes)

	if !util.ArrayEquals(res, myRes) {
		t.Error(title+": ", myRes, ", supposed to be ", res)
	}
}

func TestGetList(t *testing.T) {
	e, _ := NewEnforcer("examples/rbac_model.conf", "examples/rbac_policy.csv")

	testStringList(t, "Subjects", e.GetAllSubjects, []string{"alice", "bob", "data2_admin"})
	testStringList(t, "Objects", e.GetAllObjects, []string{"data1", "data2"})
	testStringList(t, "Actions", e.GetAllActions, []string{"read", "write"})
	testStringList(t, "Roles", e.GetAllRoles, []string{"data2_admin"})
}

func testGetPolicy(t *testing.T, e *Enforcer, res [][]string) {
	t.Helper()
	myRes := e.GetPolicy()
	t.Log("Policy: ", myRes)

	if !util.Array2DEquals(res, myRes) {
		t.Error("Policy: ", myRes, ", supposed to be ", res)
	}
}

func testGetFilteredPolicy(t *testing.T, e *Enforcer, fieldIndex int, res [][]string, fieldValues ...string) {
	t.Helper()
	myRes := e.GetFilteredPolicy(fieldIndex, fieldValues...)
	t.Log("Policy for ", util.ParamsToString(fieldValues...), ": ", myRes)

	if !util.Array2DEquals(res, myRes) {
		t.Error("Policy for ", util.ParamsToString(fieldValues...), ": ", myRes, ", supposed to be ", res)
	}
}

func testGetFilteredNamedPolicyWithMatcher(t *testing.T, e *Enforcer, ptype string, matcher string, res [][]string) {
	t.Helper()
	myRes, err := e.GetFilteredNamedPolicyWithMatcher(ptype, matcher)
	t.Log("Policy for", matcher, ": ", myRes)

	if err != nil {
		t.Error(err)
	}

	if !util.Array2DEquals(res, myRes) {
		t.Error("Policy for ", matcher, ": ", myRes, ", supposed to be ", res)
	}
}

func testGetGroupingPolicy(t *testing.T, e *Enforcer, res [][]string) {
	t.Helper()
	myRes := e.GetGroupingPolicy()
	t.Log("Grouping policy: ", myRes)

	if !util.Array2DEquals(res, myRes) {
		t.Error("Grouping policy: ", myRes, ", supposed to be ", res)
	}
}

func testGetFilteredGroupingPolicy(t *testing.T, e *Enforcer, fieldIndex int, res [][]string, fieldValues ...string) {
	t.Helper()
	myRes := e.GetFilteredGroupingPolicy(fieldIndex, fieldValues...)
	t.Log("Grouping policy for ", util.ParamsToString(fieldValues...), ": ", myRes)

	if !util.Array2DEquals(res, myRes) {
		t.Error("Grouping policy for ", util.ParamsToString(fieldValues...), ": ", myRes, ", supposed to be ", res)
	}
}

func testHasPolicy(t *testing.T, e *Enforcer, policy []string, res bool) {
	t.Helper()
	myRes := e.HasPolicy(policy)
	t.Log("Has policy ", util.ArrayToString(policy), ": ", myRes)

	if res != myRes {
		t.Error("Has policy ", util.ArrayToString(policy), ": ", myRes, ", supposed to be ", res)
	}
}

func testHasGroupingPolicy(t *testing.T, e *Enforcer, policy []string, res bool) {
	t.Helper()
	myRes := e.HasGroupingPolicy(policy)
	t.Log("Has grouping policy ", util.ArrayToString(policy), ": ", myRes)

	if res != myRes {
		t.Error("Has grouping policy ", util.ArrayToString(policy), ": ", myRes, ", supposed to be ", res)
	}
}

func TestGetPolicyAPI(t *testing.T) {
	e, _ := NewEnforcer("examples/rbac_model.conf", "examples/rbac_policy.csv")

	testGetPolicy(t, e, [][]string{
		{"alice", "data1", "read"},
		{"bob", "data2", "write"},
		{"data2_admin", "data2", "read"},
		{"data2_admin", "data2", "write"}})

	testGetFilteredPolicy(t, e, 0, [][]string{{"alice", "data1", "read"}}, "alice")
	testGetFilteredPolicy(t, e, 0, [][]string{{"bob", "data2", "write"}}, "bob")
	testGetFilteredPolicy(t, e, 0, [][]string{{"data2_admin", "data2", "read"}, {"data2_admin", "data2", "write"}}, "data2_admin")
	testGetFilteredPolicy(t, e, 1, [][]string{{"alice", "data1", "read"}}, "data1")
	testGetFilteredPolicy(t, e, 1, [][]string{{"bob", "data2", "write"}, {"data2_admin", "data2", "read"}, {"data2_admin", "data2", "write"}}, "data2")
	testGetFilteredPolicy(t, e, 2, [][]string{{"alice", "data1", "read"}, {"data2_admin", "data2", "read"}}, "read")
	testGetFilteredPolicy(t, e, 2, [][]string{{"bob", "data2", "write"}, {"data2_admin", "data2", "write"}}, "write")

	testGetFilteredNamedPolicyWithMatcher(t, e, "p", "'alice' == p.sub", [][]string{{"alice", "data1", "read"}})
	testGetFilteredNamedPolicyWithMatcher(t, e, "p", "keyMatch2(p.sub, '*')", [][]string{
		{"alice", "data1", "read"},
		{"bob", "data2", "write"},
		{"data2_admin", "data2", "read"},
		{"data2_admin", "data2", "write"}})

	testGetFilteredPolicy(t, e, 0, [][]string{{"data2_admin", "data2", "read"}, {"data2_admin", "data2", "write"}}, "data2_admin", "data2")
	// Note: "" (empty string) in fieldValues means matching all values.
	testGetFilteredPolicy(t, e, 0, [][]string{{"data2_admin", "data2", "read"}}, "data2_admin", "", "read")
	testGetFilteredPolicy(t, e, 1, [][]string{{"bob", "data2", "write"}, {"data2_admin", "data2", "write"}}, "data2", "write")

	testHasPolicy(t, e, []string{"alice", "data1", "read"}, true)
	testHasPolicy(t, e, []string{"bob", "data2", "write"}, true)
	testHasPolicy(t, e, []string{"alice", "data2", "read"}, false)
	testHasPolicy(t, e, []string{"bob", "data3", "write"}, false)

	testGetGroupingPolicy(t, e, [][]string{{"alice", "data2_admin"}})

	testGetFilteredGroupingPolicy(t, e, 0, [][]string{{"alice", "data2_admin"}}, "alice")
	testGetFilteredGroupingPolicy(t, e, 0, [][]string{}, "bob")
	testGetFilteredGroupingPolicy(t, e, 1, [][]string{}, "data1_admin")
	testGetFilteredGroupingPolicy(t, e, 1, [][]string{{"alice", "data2_admin"}}, "data2_admin")
	// Note: "" (empty string) in fieldValues means matching all values.
	testGetFilteredGroupingPolicy(t, e, 0, [][]string{{"alice", "data2_admin"}}, "", "data2_admin")

	testHasGroupingPolicy(t, e, []string{"alice", "data2_admin"}, true)
	testHasGroupingPolicy(t, e, []string{"bob", "data2_admin"}, false)
}

func TestModifyPolicyAPI(t *testing.T) {
	e, _ := NewEnforcer("examples/rbac_model.conf", "examples/rbac_policy.csv")

	testGetPolicy(t, e, [][]string{
		{"alice", "data1", "read"},
		{"bob", "data2", "write"},
		{"data2_admin", "data2", "read"},
		{"data2_admin", "data2", "write"}})

	_, _ = e.RemovePolicy("alice", "data1", "read")
	_, _ = e.RemovePolicy("bob", "data2", "write")
	_, _ = e.RemovePolicy("alice", "data1", "read")
	_, _ = e.AddPolicy("eve", "data3", "read")
	_, _ = e.AddPolicy("eve", "data3", "read")

	rules := [][]string{
		{"jack", "data4", "read"},
		{"jack", "data4", "read"},
		{"jack", "data4", "read"},
		{"katy", "data4", "write"},
		{"leyo", "data4", "read"},
		{"katy", "data4", "write"},
		{"katy", "data4", "write"},
		{"ham", "data4", "write"},
	}

	_, _ = e.AddPolicies(rules)
	_, _ = e.AddPolicies(rules)

	testGetPolicy(t, e, [][]string{
		{"data2_admin", "data2", "read"},
		{"data2_admin", "data2", "write"},
		{"eve", "data3", "read"},
		{"jack", "data4", "read"},
		{"katy", "data4", "write"},
		{"leyo", "data4", "read"},
		{"ham", "data4", "write"}})

	_, _ = e.RemovePolicies(rules)
	_, _ = e.RemovePolicies(rules)

	namedPolicy := []string{"eve", "data3", "read"}
	_, _ = e.RemoveNamedPolicy("p", namedPolicy)
	_, _ = e.AddNamedPolicy("p", namedPolicy)

	testGetPolicy(t, e, [][]string{
		{"data2_admin", "data2", "read"},
		{"data2_admin", "data2", "write"},
		{"eve", "data3", "read"}})

	_, _ = e.RemoveFilteredPolicy(1, "data2")

	testGetPolicy(t, e, [][]string{{"eve", "data3", "read"}})

	_, _ = e.UpdatePolicy([]string{"eve", "data3", "read"}, []string{"eve", "data3", "write"})

	testGetPolicy(t, e, [][]string{{"eve", "data3", "write"}})

	// This test shows a rollback effect.
	// _, _ = e.UpdatePolicies([][]string{{"eve", "data3", "write"}, {"jack", "data4", "read"}}, [][]string{{"eve", "data3", "read"}, {"jack", "data4", "write"}})
	// testGetPolicy(t, e, [][]string{{"eve", "data3", "read"}, {"jack", "data4", "write"}})

	_, _ = e.AddPolicies(rules)
	_, _ = e.UpdatePolicies([][]string{{"eve", "data3", "write"}, {"leyo", "data4", "read"}, {"katy", "data4", "write"}},
		[][]string{{"eve", "data3", "read"}, {"leyo", "data4", "write"}, {"katy", "data1", "write"}})
	testGetPolicy(t, e, [][]string{{"eve", "data3", "read"}, {"jack", "data4", "read"}, {"katy", "data1", "write"}, {"leyo", "data4", "write"}, {"ham", "data4", "write"}})
}

func TestModifyGroupingPolicyAPI(t *testing.T) {
	e, _ := NewEnforcer("examples/rbac_model.conf", "examples/rbac_policy.csv")

	testGetRoles(t, e, []string{"data2_admin"}, "alice")
	testGetRoles(t, e, []string{}, "bob")
	testGetRoles(t, e, []string{}, "eve")
	testGetRoles(t, e, []string{}, "non_exist")

	_, _ = e.RemoveGroupingPolicy("alice", "data2_admin")
	_, _ = e.AddGroupingPolicy("bob", "data1_admin")
	_, _ = e.AddGroupingPolicy("eve", "data3_admin")

	groupingRules := [][]string{
		{"ham", "data4_admin"},
		{"jack", "data5_admin"},
	}

	_, _ = e.AddGroupingPolicies(groupingRules)
	testGetRoles(t, e, []string{"data4_admin"}, "ham")
	testGetRoles(t, e, []string{"data5_admin"}, "jack")
	_, _ = e.RemoveGroupingPolicies(groupingRules)

	testGetRoles(t, e, []string{}, "alice")
	namedGroupingPolicy := []string{"alice", "data2_admin"}
	testGetRoles(t, e, []string{}, "alice")
	_, _ = e.AddNamedGroupingPolicy("g", namedGroupingPolicy)
	testGetRoles(t, e, []string{"data2_admin"}, "alice")
	_, _ = e.RemoveNamedGroupingPolicy("g", namedGroupingPolicy)

	_, _ = e.AddNamedGroupingPolicies("g", groupingRules)
	_, _ = e.AddNamedGroupingPolicies("g", groupingRules)
	testGetRoles(t, e, []string{"data4_admin"}, "ham")
	testGetRoles(t, e, []string{"data5_admin"}, "jack")
	_, _ = e.RemoveNamedGroupingPolicies("g", groupingRules)
	_, _ = e.RemoveNamedGroupingPolicies("g", groupingRules)

	testGetRoles(t, e, []string{}, "alice")
	testGetRoles(t, e, []string{"data1_admin"}, "bob")
	testGetRoles(t, e, []string{"data3_admin"}, "eve")
	testGetRoles(t, e, []string{}, "non_exist")

	testGetUsers(t, e, []string{"bob"}, "data1_admin")
	testGetUsers(t, e, []string{}, "data2_admin")
	testGetUsers(t, e, []string{"eve"}, "data3_admin")

	_, _ = e.RemoveFilteredGroupingPolicy(0, "bob")

	testGetRoles(t, e, []string{}, "alice")
	testGetRoles(t, e, []string{}, "bob")
	testGetRoles(t, e, []string{"data3_admin"}, "eve")
	testGetRoles(t, e, []string{}, "non_exist")

	testGetUsers(t, e, []string{}, "data1_admin")
	testGetUsers(t, e, []string{}, "data2_admin")
	testGetUsers(t, e, []string{"eve"}, "data3_admin")
	_, _ = e.AddGroupingPolicy("data3_admin", "data4_admin")
	_, _ = e.UpdateGroupingPolicy([]string{"eve", "data3_admin"}, []string{"eve", "admin"})
	_, _ = e.UpdateGroupingPolicy([]string{"data3_admin", "data4_admin"}, []string{"admin", "data4_admin"})
	testGetUsers(t, e, []string{"admin"}, "data4_admin")
	testGetUsers(t, e, []string{"eve"}, "admin")

	testGetRoles(t, e, []string{"admin"}, "eve")
	testGetRoles(t, e, []string{"data4_admin"}, "admin")

	_, _ = e.UpdateGroupingPolicies([][]string{{"eve", "admin"}}, [][]string{{"eve", "admin_groups"}})
	_, _ = e.UpdateGroupingPolicies([][]string{{"admin", "data4_admin"}}, [][]string{{"admin", "data5_admin"}})
	testGetUsers(t, e, []string{"admin"}, "data5_admin")
	testGetUsers(t, e, []string{"eve"}, "admin_groups")

	testGetRoles(t, e, []string{"data5_admin"}, "admin")
	testGetRoles(t, e, []string{"admin_groups"}, "eve")

}

func assert(t *testing.T, res, exp interface{}) {
	ok := reflect.DeepEqual(res, exp)
	if !ok {
		t.Errorf("%v != %v", res, exp)
	}
}

func TestExistedOrMissedPolicy(t *testing.T) {
	var (
		ok       bool
		err      error
		affected [][]string
		e        *Enforcer
	)

	// test for adding a policy that already exists
	e, _ = NewEnforcer("examples/rbac_model.conf", "examples/rbac_policy.csv")

	testGetPolicy(t, e, [][]string{
		{"alice", "data1", "read"},
		{"bob", "data2", "write"},
		{"data2_admin", "data2", "read"},
		{"data2_admin", "data2", "write"}})

	ok = e.model.AddPolicy("p", "p", []string{"alice", "data1", "read"})
	assert(t, ok, false)

	ok = e.model.AddPolicies("p", "p", [][]string{
		{"alice", "data1", "write"},
		{"alice", "data1", "read"}, // duplicate
	})
	assert(t, ok, false)
	testGetPolicy(t, e, [][]string{
		{"alice", "data1", "read"},
		{"bob", "data2", "write"},
		{"data2_admin", "data2", "read"},
		{"data2_admin", "data2", "write"},

		//{"alice", "data1", "write"},
	})

	affected = e.model.AddPoliciesWithAffected("p", "p", [][]string{
		{"alice", "data1", "write"}, // new
		{"alice", "data1", "read"},  // duplicate
	})
	assert(t, affected, [][]string{
		{"alice", "data1", "write"}, // new
		//{"alice", "data1", "read"}, // duplicate
	})
	testGetPolicy(t, e, [][]string{
		{"alice", "data1", "read"},
		{"bob", "data2", "write"},
		{"data2_admin", "data2", "read"},
		{"data2_admin", "data2", "write"},

		{"alice", "data1", "write"}, //new
	})

	ok, err = e.AddPolicy([]string{"alice", "data1", "read"})
	assert(t, ok, false)
	assert(t, err, nil)

	ok, err = e.AddPolicies([][]string{
		{"alice", "data1", "read"},
		{"alice", "data2", "read"}, // new
	})
	assert(t, ok, true)
	assert(t, err, nil)
	testGetPolicy(t, e, [][]string{
		{"alice", "data1", "read"},
		{"bob", "data2", "write"},
		{"data2_admin", "data2", "read"},
		{"data2_admin", "data2", "write"},

		{"alice", "data1", "write"},
		{"alice", "data2", "read"}, // new
	})

	ok, err = e.AddPolicies([][]string{
		{"alice", "data1", "read"}, // duplicate
		{"alice", "data2", "read"}, // duplicate
	})
	assert(t, ok, false)
	assert(t, err, nil)

	// test for removing a policy that does not exist
	e, _ = NewEnforcer("examples/rbac_model.conf", "examples/rbac_policy.csv")

	testGetPolicy(t, e, [][]string{
		{"alice", "data1", "read"},
		{"bob", "data2", "write"},
		{"data2_admin", "data2", "read"},
		{"data2_admin", "data2", "write"}})

	ok = e.model.RemovePolicy("p", "p", []string{"eva", "data1", "read"})
	assert(t, ok, false)

	ok = e.model.RemovePolicies("p", "p", [][]string{
		{"alice", "data1", "read"},
		{"eva", "data1", "read"}, // missing
	})
	assert(t, ok, false)
	testGetPolicy(t, e, [][]string{
		{"alice", "data1", "read"},
		{"bob", "data2", "write"},
		{"data2_admin", "data2", "read"},
		{"data2_admin", "data2", "write"},
	})

	affected = e.model.RemovePoliciesWithAffected("p", "p", [][]string{
		{"alice", "data1", "read"},
		{"eva", "data1", "read"}, // missing
	})
	assert(t, affected, [][]string{
		{"alice", "data1", "read"},
		//{"eva", "data1", "read"}, // missing
	})
	testGetPolicy(t, e, [][]string{
		//{"alice", "data1", "read"}, // deleted
		{"bob", "data2", "write"},
		{"data2_admin", "data2", "read"},
		{"data2_admin", "data2", "write"},
	})

	ok, err = e.RemovePolicy([]string{"alice", "data1", "read"})
	assert(t, ok, false)
	assert(t, err, nil)

	ok, err = e.RemovePolicies([][]string{
		{"bob", "data2", "write"},
		{"eva", "data1", "read"}, // missing
	})
	assert(t, ok, true)
	assert(t, err, nil)
	testGetPolicy(t, e, [][]string{
		//{"alice", "data1", "read"},
		//{"bob", "data2", "write"}, // deleted
		{"data2_admin", "data2", "read"},
		{"data2_admin", "data2", "write"},
	})

	ok, err = e.RemovePolicies([][]string{
		{"bob", "data2", "write"}, // missing
		{"eva", "data1", "read"},  // missing
	})
	assert(t, ok, false)
	assert(t, err, nil)

	// test for update policies
	e, _ = NewEnforcer("examples/rbac_model.conf", "examples/rbac_policy.csv")

	testGetPolicy(t, e, [][]string{
		{"alice", "data1", "read"},
		{"bob", "data2", "write"},
		{"data2_admin", "data2", "read"},
		{"data2_admin", "data2", "write"}})

	ok = e.model.UpdatePolicy("p", "p",
		[]string{"carol", "data1", "get"}, // missing
		[]string{"carol", "data1", "read"},
	)
	assert(t, ok, false)

	ok = e.model.UpdatePolicy("p", "p",
		[]string{"bob", "data2", "write"},
		[]string{"alice", "data1", "read"}, // existing
	)
	assert(t, ok, false)

	ok = e.model.UpdatePolicies("p", "p", [][]string{
		{"alice", "data1", "read"},
		{"eva", "data1", "read"}, // missing
	},
		[][]string{
			{"eva", "data2", "read"},
			{"eva", "data2", "write"},
		})
	assert(t, ok, false)
	testGetPolicy(t, e, [][]string{
		{"alice", "data1", "read"},
		{"bob", "data2", "write"},
		{"data2_admin", "data2", "read"},
		{"data2_admin", "data2", "write"},
	})

	ok = e.model.UpdatePolicies("p", "p", [][]string{
		{"bob", "data2", "write"},
		{"data2_admin", "data2", "write"},
	},
		[][]string{
			{"eva", "data1", "read"},
			{"alice", "data1", "read"}, // existing
		})
	assert(t, ok, false)
	testGetPolicy(t, e, [][]string{
		{"alice", "data1", "read"},
		{"bob", "data2", "write"},
		{"data2_admin", "data2", "read"},
		{"data2_admin", "data2", "write"},
	})

	affected = e.model.UpdatePoliciesWithAffected("p", "p", [][]string{
		{"alice", "data1", "read"},
		{"bob", "data2", "write"},
	},
		[][]string{
			{"eva", "data1", "read"},
			{"data2_admin", "data2", "read"}, // existing
		})
	assert(t, affected, [][]string{
		{"alice", "data1", "read"},
		//{"bob", "data2", "write"},
	})
	testGetPolicy(t, e, [][]string{
		{"eva", "data1", "read"}, // updated
		{"bob", "data2", "write"},
		{"data2_admin", "data2", "read"},
		{"data2_admin", "data2", "write"},
	})

	affected = e.model.UpdatePoliciesWithAffected("p", "p", [][]string{
		{"eva", "data1", "read"},
		{"carol", "data2", "write"}, // missing
	},
		[][]string{
			{"alice", "data1", "read"},
			{"alice", "data2", "write"},
		})
	assert(t, affected, [][]string{
		{"eva", "data1", "read"},
		//{"bob", "data2", "write"},
	})
	testGetPolicy(t, e, [][]string{
		{"alice", "data1", "read"}, // updated
		{"bob", "data2", "write"},
		{"data2_admin", "data2", "read"},
		{"data2_admin", "data2", "write"},
	})

	ok, err = e.UpdatePolicy(
		[]string{"carol", "data1", "get"}, // missing
		[]string{"carol", "data1", "read"},
	)
	assert(t, ok, false)
	assert(t, err, nil)

	ok, err = e.UpdatePolicy(
		[]string{"bob", "data2", "write"},
		[]string{"alice", "data1", "read"}, // existing
	)
	assert(t, ok, false)
	assert(t, err, nil)

	ok, err = e.UpdatePolicies([][]string{
		{"alice", "data1", "read"},
		{"bob", "data2", "write"},
	},
		[][]string{
			{"eva", "data1", "read"},
			{"data2_admin", "data2", "read"}, // existing
		})
	assert(t, ok, true)
	assert(t, err, nil)
	testGetPolicy(t, e, [][]string{
		{"eva", "data1", "read"}, // updated
		{"bob", "data2", "write"},
		{"data2_admin", "data2", "read"},
		{"data2_admin", "data2", "write"},
	})

	ok, err = e.UpdatePolicies([][]string{
		{"eva", "data1", "read"},
		{"carol", "data2", "write"}, // missing
	},
		[][]string{
			{"alice", "data1", "read"},
			{"alice", "data2", "write"},
		})
	assert(t, ok, true)
	assert(t, err, nil)
	testGetPolicy(t, e, [][]string{
		{"alice", "data1", "read"}, // updated
		{"bob", "data2", "write"},
		{"data2_admin", "data2", "read"},
		{"data2_admin", "data2", "write"},
	})
}
