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

func testStringList(t *testing.T, title string, f func() []string, res []string) {
	t.Helper()
	myRes := f()
	log.Print(title+": ", myRes)

	if !util.ArrayEquals(res, myRes) {
		t.Error(title+": ", myRes, ", supposed to be ", res)
	}
}

func TestGetList(t *testing.T) {
	e := NewEnforcer("examples/rbac_model.conf", "examples/rbac_policy.csv")

	testStringList(t, "Subjects", e.GetAllSubjects, []string{"alice", "bob", "data2_admin"})
	testStringList(t, "Objeccts", e.GetAllObjects, []string{"data1", "data2"})
	testStringList(t, "Actions", e.GetAllActions, []string{"read", "write"})
	testStringList(t, "Roles", e.GetAllRoles, []string{"data2_admin"})
}

func testGetPolicy(t *testing.T, e *Enforcer, res [][]string) {
	t.Helper()
	myRes := e.GetPolicy()
	log.Print("Policy: ", myRes)

	if !util.Array2DEquals(res, myRes) {
		t.Error("Policy: ", myRes, ", supposed to be ", res)
	}
}

func testGetFilteredPolicy(t *testing.T, e *Enforcer, fieldIndex int, res [][]string, fieldValues ...string) {
	t.Helper()
	myRes := e.GetFilteredPolicy(fieldIndex, fieldValues...)
	log.Print("Policy for ", util.ParamsToString(fieldValues...), ": ", myRes)

	if !util.Array2DEquals(res, myRes) {
		t.Error("Policy for ", util.ParamsToString(fieldValues...), ": ", myRes, ", supposed to be ", res)
	}
}

func testGetGroupingPolicy(t *testing.T, e *Enforcer, res [][]string) {
	t.Helper()
	myRes := e.GetGroupingPolicy()
	log.Print("Grouping policy: ", myRes)

	if !util.Array2DEquals(res, myRes) {
		t.Error("Grouping policy: ", myRes, ", supposed to be ", res)
	}
}

func testGetFilteredGroupingPolicy(t *testing.T, e *Enforcer, fieldIndex int, res [][]string, fieldValues ...string) {
	t.Helper()
	myRes := e.GetFilteredGroupingPolicy(fieldIndex, fieldValues...)
	log.Print("Grouping policy for ", util.ParamsToString(fieldValues...), ": ", myRes)

	if !util.Array2DEquals(res, myRes) {
		t.Error("Grouping policy for ", util.ParamsToString(fieldValues...), ": ", myRes, ", supposed to be ", res)
	}
}

func testHasPolicy(t *testing.T, e *Enforcer, policy []string, res bool) {
	t.Helper()
	myRes := e.HasPolicy(policy)
	log.Print("Has policy ", util.ArrayToString(policy), ": ", myRes)

	if res != myRes {
		t.Error("Has policy ", util.ArrayToString(policy), ": ", myRes, ", supposed to be ", res)
	}
}

func testHasGroupingPolicy(t *testing.T, e *Enforcer, policy []string, res bool) {
	t.Helper()
	myRes := e.HasGroupingPolicy(policy)
	log.Print("Has grouping policy ", util.ArrayToString(policy), ": ", myRes)

	if res != myRes {
		t.Error("Has grouping policy ", util.ArrayToString(policy), ": ", myRes, ", supposed to be ", res)
	}
}

func testHasGroupingPolicyStringInput(t *testing.T, e *Enforcer, policy1 string, policy2 string, res bool) {
	t.Helper()
	myRes := e.HasGroupingPolicy(policy1, policy2)
	log.Print("Has grouping policy ", policy1, policy2, ": ", myRes)

	if res != myRes {
		t.Error("Has grouping policy ", policy1, policy2, ": ", myRes, ", supposed to be ", res)
	}
}

func TestGetPolicy(t *testing.T) {
	e := NewEnforcer("examples/rbac_model.conf", "examples/rbac_policy.csv")

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

	testGetGroupingPolicy(t, e, [][]string{
		{"alice", "data2_admin"}})

	testGetFilteredGroupingPolicy(t, e, 0, [][]string{{"alice", "data2_admin"}}, "alice")
	testGetFilteredGroupingPolicy(t, e, 0, [][]string{}, "bob")
	testGetFilteredGroupingPolicy(t, e, 1, [][]string{}, "data1_admin")
	testGetFilteredGroupingPolicy(t, e, 1, [][]string{{"alice", "data2_admin"}}, "data2_admin")
	// Note: "" (empty string) in fieldValues means matching all values.
	testGetFilteredGroupingPolicy(t, e, 0, [][]string{{"alice", "data2_admin"}}, "", "data2_admin")

	testHasGroupingPolicy(t, e, []string{"alice", "data2_admin"}, true)
	testHasGroupingPolicy(t, e, []string{"bob", "data2_admin"}, false)

	testHasGroupingPolicyStringInput(t, e, "alice", "data2_admin", true)
	testHasGroupingPolicyStringInput(t, e, "bob", "data2_admin", false)
}

func TestModifyPolicy(t *testing.T) {
	e := NewEnforcer("examples/rbac_model.conf", "examples/rbac_policy.csv")

	testGetPolicy(t, e, [][]string{
		{"alice", "data1", "read"},
		{"bob", "data2", "write"},
		{"data2_admin", "data2", "read"},
		{"data2_admin", "data2", "write"}})

	e.RemovePolicy("alice", "data1", "read")
	e.RemovePolicy("bob", "data2", "write")
	e.RemovePolicy("alice", "data1", "read")
	e.AddPolicy("eve", "data3", "read")
	e.AddPolicy("eve", "data3", "read")

	namedPolicy := []string{"eve", "data3", "read"}
	e.RemoveNamedPolicy("p", namedPolicy)
	e.AddNamedPolicy("p", namedPolicy)

	testGetPolicy(t, e, [][]string{
		{"data2_admin", "data2", "read"},
		{"data2_admin", "data2", "write"},
		{"eve", "data3", "read"}})

	e.RemoveFilteredPolicy(1, "data2")

	testGetPolicy(t, e, [][]string{{"eve", "data3", "read"}})
}

func TestModifyGroupingPolicy(t *testing.T) {
	e := NewEnforcer("examples/rbac_model.conf", "examples/rbac_policy.csv")

	testGetRoles(t, e, "alice", []string{"data2_admin"})
	testGetRoles(t, e, "bob", []string{})
	testGetRoles(t, e, "eve", []string{})
	testGetRoles(t, e, "non_exist", []string{})

	e.RemoveGroupingPolicy("alice", "data2_admin")
	e.AddGroupingPolicy("bob", "data1_admin")
	e.AddGroupingPolicy("eve", "data3_admin")

	namedGroupingPolicy := []string{"alice", "data2_admin"}
	testGetRoles(t, e, "alice", []string{})
	e.AddNamedGroupingPolicy("g", namedGroupingPolicy)
	testGetRoles(t, e, "alice", []string{"data2_admin"})
	e.RemoveNamedGroupingPolicy("g", namedGroupingPolicy)

	testGetRoles(t, e, "alice", []string{})
	testGetRoles(t, e, "bob", []string{"data1_admin"})
	testGetRoles(t, e, "eve", []string{"data3_admin"})
	testGetRoles(t, e, "non_exist", []string{})

	testGetUsers(t, e, "data1_admin", []string{"bob"})
	testGetUsers(t, e, "data2_admin", []string{})
	testGetUsers(t, e, "data3_admin", []string{"eve"})

	e.RemoveFilteredGroupingPolicy(0, "bob")

	testGetRoles(t, e, "alice", []string{})
	testGetRoles(t, e, "bob", []string{})
	testGetRoles(t, e, "eve", []string{"data3_admin"})
	testGetRoles(t, e, "non_exist", []string{})

	testGetUsers(t, e, "data1_admin", []string{})
	testGetUsers(t, e, "data2_admin", []string{})
	testGetUsers(t, e, "data3_admin", []string{"eve"})
}
