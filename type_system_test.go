// Copyright 2026 The casbin Authors. All Rights Reserved.
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
	"strings"
	"testing"

	"github.com/casbin/casbin/v3/model"
	"github.com/casbin/casbin/v3/persist"
	stringadapter "github.com/casbin/casbin/v3/persist/string-adapter"
)

const typedRBACModel = `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[type_definition]
user = user:
role = role:

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act
`

func TestTypedRoleListsAndEnforce(t *testing.T) {
	policy := strings.TrimSpace(`
p, role:data2_admin, data2, read
p, role:data2_admin, data2, write
p, user:bob, data2, write
g, user:alice, role:data2_admin
`)

	m, err := model.NewModelFromString(typedRBACModel)
	if err != nil {
		t.Fatalf("load model failed: %v", err)
	}

	e, err := NewEnforcer(m, stringadapter.NewAdapter(policy))
	if err != nil {
		t.Fatalf("new enforcer failed: %v", err)
	}

	testStringList(t, "Roles", e.GetAllRoles, []string{"role:data2_admin"})
	testStringList(t, "Users", e.GetAllUsers, []string{"user:alice", "user:bob"})

	testEnforce(t, e, "user:alice", "data2", "read", true)
	testEnforce(t, e, "user:bob", "data2", "write", true)

	users, err := e.GetImplicitUsersForPermission("data2", "write")
	if err != nil {
		t.Fatalf("GetImplicitUsersForPermission failed: %v", err)
	}
	if len(users) != 2 || users[0] != "user:alice" || users[1] != "user:bob" {
		t.Fatalf("GetImplicitUsersForPermission got %v", users)
	}
}

func TestTypedRoleValidationOnPolicyMutation(t *testing.T) {
	m, err := model.NewModelFromString(typedRBACModel)
	if err != nil {
		t.Fatalf("load model failed: %v", err)
	}

	e, err := NewEnforcer(m)
	if err != nil {
		t.Fatalf("new enforcer failed: %v", err)
	}

	if _, err := e.AddRoleForUser("user:alice", "role:admin"); err != nil {
		t.Fatalf("AddRoleForUser should succeed: %v", err)
	}

	if _, err := e.AddRoleForUser("user:alice", "user:bob"); err == nil || !strings.Contains(err.Error(), "expected role") {
		t.Fatalf("expected role type error, got %v", err)
	}

	if _, err := e.AddPermissionForUser("group:ops", "data1", "read"); err == nil || !strings.Contains(err.Error(), "does not match any configured user/role prefix") {
		t.Fatalf("expected typed subject error, got %v", err)
	}
}

func TestTypedRoleValidationOnPolicyLoad(t *testing.T) {
	m, err := model.NewModelFromString(typedRBACModel)
	if err != nil {
		t.Fatalf("load model failed: %v", err)
	}

	if err := persist.LoadPolicyLine("g, user:alice, user:bob", m); err == nil || !strings.Contains(err.Error(), "expected role") {
		t.Fatalf("expected invalid grouping policy error, got %v", err)
	}

	if err := persist.LoadPolicyLine("p, team:ops, data1, read", m); err == nil || !strings.Contains(err.Error(), "does not match any configured user/role prefix") {
		t.Fatalf("expected invalid subject policy error, got %v", err)
	}
}

func TestTypedRoleDefinitionValidation(t *testing.T) {
	invalidModel := `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[type_definition]
user = actor:
role = actor:

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act
`

	if _, err := model.NewModelFromString(invalidModel); err == nil || !strings.Contains(err.Error(), "prefixes must be different") {
		t.Fatalf("expected invalid type definition error, got %v", err)
	}
}
