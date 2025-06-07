// Copyright 2025 The casbin Authors. All Rights Reserved.
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
	"encoding/json"
	"fmt"
	"testing"

	"github.com/casbin/casbin/v2/util"
)

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

func TestABACMapRequest(t *testing.T) {
	e, _ := NewEnforcer("examples/abac_model.conf")

	data1 := map[string]interface{}{
		"Name":  "data1",
		"Owner": "alice",
	}
	data2 := map[string]interface{}{
		"Name":  "data2",
		"Owner": "bob",
	}

	testEnforce(t, e, "alice", data1, "read", true)
	testEnforce(t, e, "alice", data1, "write", true)
	testEnforce(t, e, "alice", data2, "read", false)
	testEnforce(t, e, "alice", data2, "write", false)
	testEnforce(t, e, "bob", data1, "read", false)
	testEnforce(t, e, "bob", data1, "write", false)
	testEnforce(t, e, "bob", data2, "read", true)
	testEnforce(t, e, "bob", data2, "write", true)
}

func TestABACTypes(t *testing.T) {
	e, _ := NewEnforcer("examples/abac_model.conf")
	matcher := `"moderator" IN r.sub.Roles && r.sub.Enabled == true && r.sub.Age >= 21 && r.sub.Name != "foo"`
	e.GetModel()["m"]["m"].Value = util.RemoveComments(util.EscapeAssertion(matcher))

	structRequest := struct {
		Roles   []interface{}
		Enabled bool
		Age     int
		Name    string
	}{
		Roles:   []interface{}{"user", "moderator"},
		Enabled: true,
		Age:     30,
		Name:    "alice",
	}
	testEnforce(t, e, structRequest, "", "", true)

	mapRequest := map[string]interface{}{
		"Roles":   []interface{}{"user", "moderator"},
		"Enabled": true,
		"Age":     30,
		"Name":    "alice",
	}
	testEnforce(t, e, mapRequest, nil, "", true)

	e.EnableAcceptJsonRequest(true)
	jsonRequest, _ := json.Marshal(mapRequest)
	testEnforce(t, e, string(jsonRequest), "", "", true)
}

func TestABACJsonRequest(t *testing.T) {
	e, _ := NewEnforcer("examples/abac_model.conf")
	e.EnableAcceptJsonRequest(true)

	data1Json := `{ "Name": "data1", "Owner": "alice"}`
	data2Json := `{ "Name": "data2", "Owner": "bob"}`

	testEnforce(t, e, "alice", data1Json, "read", true)
	testEnforce(t, e, "alice", data1Json, "write", true)
	testEnforce(t, e, "alice", data2Json, "read", false)
	testEnforce(t, e, "alice", data2Json, "write", false)
	testEnforce(t, e, "bob", data1Json, "read", false)
	testEnforce(t, e, "bob", data1Json, "write", false)
	testEnforce(t, e, "bob", data2Json, "read", true)
	testEnforce(t, e, "bob", data2Json, "write", true)

	e, _ = NewEnforcer("examples/abac_not_using_policy_model.conf", "examples/abac_rule_effect_policy.csv")
	e.EnableAcceptJsonRequest(true)

	testEnforce(t, e, "alice", data1Json, "read", true)
	testEnforce(t, e, "alice", data1Json, "write", true)
	testEnforce(t, e, "alice", data2Json, "read", false)
	testEnforce(t, e, "alice", data2Json, "write", false)

	e, _ = NewEnforcer("examples/abac_rule_model.conf", "examples/abac_rule_policy.csv")
	e.EnableAcceptJsonRequest(true)
	sub1Json := `{"Name": "alice", "Age": 16}`
	sub2Json := `{"Name": "alice", "Age": 20}`
	sub3Json := `{"Name": "alice", "Age": 65}`

	testEnforce(t, e, sub1Json, "/data1", "read", false)
	testEnforce(t, e, sub1Json, "/data2", "read", false)
	testEnforce(t, e, sub1Json, "/data1", "write", false)
	testEnforce(t, e, sub1Json, "/data2", "write", true)
	testEnforce(t, e, sub2Json, "/data1", "read", true)
	testEnforce(t, e, sub2Json, "/data2", "read", false)
	testEnforce(t, e, sub2Json, "/data1", "write", false)
	testEnforce(t, e, sub2Json, "/data2", "write", true)
	testEnforce(t, e, sub3Json, "/data1", "read", true)
	testEnforce(t, e, sub3Json, "/data2", "read", false)
	testEnforce(t, e, sub3Json, "/data1", "write", false)
	testEnforce(t, e, sub3Json, "/data2", "write", false)
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

func TestABACNotUsingPolicy(t *testing.T) {
	e, _ := NewEnforcer("examples/abac_not_using_policy_model.conf", "examples/abac_rule_effect_policy.csv")
	data1 := newTestResource("data1", "alice")
	data2 := newTestResource("data2", "bob")

	testEnforce(t, e, "alice", data1, "read", true)
	testEnforce(t, e, "alice", data1, "write", true)
	testEnforce(t, e, "alice", data2, "read", false)
	testEnforce(t, e, "alice", data2, "write", false)
}

func TestABACPolicy(t *testing.T) {
	e, _ := NewEnforcer("examples/abac_rule_model.conf", "examples/abac_rule_policy.csv")
	m := e.GetModel()
	for sec, ast := range m {
		fmt.Println(sec)
		for ptype, p := range ast {
			fmt.Println(ptype, p)
		}
	}
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
