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

package stringadapter

import (
	"testing"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
)

func Test_KeyMatchRbac(t *testing.T) {
	conf := `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _ , _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub)  && keyMatch(r.obj, p.obj) && regexMatch(r.act, p.act)
`
	line := `
p, alice, /alice_data/*, (GET)|(POST)
p, alice, /alice_data/resource1, POST
p, data_group_admin, /admin/*, POST
p, data_group_admin, /bob_data/*, POST
g, alice, data_group_admin
`
	a := NewAdapter(line)
	m := model.NewModel()
	err := m.LoadModelFromText(conf)
	if err != nil {
		t.Errorf("load model from text failed: %v", err.Error())
		return
	}
	e, _ := casbin.NewEnforcer(m, a)
	sub := "alice"
	obj := "/alice_data/login"
	act := "POST"
	if res, _ := e.Enforce(sub, obj, act); !res {
		t.Error("unexpected enforce result")
	}
}

func Test_StringRbac(t *testing.T) {
	conf := `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _ , _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act
`
	line := `
p, alice, data1, read
p, data_group_admin, data3, read
p, data_group_admin, data3, write
g, alice, data_group_admin
`
	a := NewAdapter(line)
	m := model.NewModel()
	err := m.LoadModelFromText(conf)
	if err != nil {
		t.Errorf("load model from text failed: %v", err.Error())
		return
	}
	e, _ := casbin.NewEnforcer(m, a)
	sub := "alice" // the user that wants to access a resource.
	obj := "data1" // the resource that is going to be accessed.
	act := "read"  // the operation that the user performs on the resource.
	if res, _ := e.Enforce(sub, obj, act); !res {
		t.Error("unexpected enforce result")
	}
}
