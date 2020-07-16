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

	"github.com/casbin/casbin/v3/util"
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
}
