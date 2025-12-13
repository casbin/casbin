// Copyright 2024 The casbin Authors. All Rights Reserved.
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

	"github.com/casbin/casbin/v3/errors"
	"github.com/casbin/casbin/v3/model"
)

func TestConstraintSOD(t *testing.T) {
	modelText := `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[constraint_definition]
c = sod("role1", "role2")

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act
`

	m, err := model.NewModelFromString(modelText)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	e, err := NewEnforcer(m)
	if err != nil {
		t.Fatalf("Failed to create enforcer: %v", err)
	}

	// Add a user to role1 should succeed
	_, err = e.AddRoleForUser("alice", "role1")
	if err != nil {
		t.Fatalf("Failed to add role1 to alice: %v", err)
	}

	// Add a different user to role2 should succeed
	_, err = e.AddRoleForUser("bob", "role2")
	if err != nil {
		t.Fatalf("Failed to add role2 to bob: %v", err)
	}

	// Try to add role2 to alice should fail (SOD violation)
	_, err = e.AddRoleForUser("alice", "role2")
	if err == nil {
		t.Fatal("Expected constraint violation error, got nil")
	}
	if !strings.Contains(err.Error(), "constraint violation") {
		t.Fatalf("Expected constraint violation error, got: %v", err)
	}
}

func TestConstraintSODMax(t *testing.T) {
	modelText := `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[constraint_definition]
c = sodMax(["role1", "role2", "role3"], 1)

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act
`

	m, err := model.NewModelFromString(modelText)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	e, err := NewEnforcer(m)
	if err != nil {
		t.Fatalf("Failed to create enforcer: %v", err)
	}

	// Add user to one role should succeed
	_, err = e.AddRoleForUser("alice", "role1")
	if err != nil {
		t.Fatalf("Failed to add role1 to alice: %v", err)
	}

	// Try to add user to another role from the set should fail
	_, err = e.AddRoleForUser("alice", "role2")
	if err == nil {
		t.Fatal("Expected constraint violation error, got nil")
	}
	if !strings.Contains(err.Error(), "constraint violation") {
		t.Fatalf("Expected constraint violation error, got: %v", err)
	}
}

func TestConstraintRoleMax(t *testing.T) {
	modelText := `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[constraint_definition]
c = roleMax("admin", 2)

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act
`

	m, err := model.NewModelFromString(modelText)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	e, err := NewEnforcer(m)
	if err != nil {
		t.Fatalf("Failed to create enforcer: %v", err)
	}

	// Add first user to admin role should succeed
	_, err = e.AddRoleForUser("alice", "admin")
	if err != nil {
		t.Fatalf("Failed to add admin to alice: %v", err)
	}

	// Add second user to admin role should succeed
	_, err = e.AddRoleForUser("bob", "admin")
	if err != nil {
		t.Fatalf("Failed to add admin to bob: %v", err)
	}

	// Try to add third user to admin role should fail (exceeds max)
	_, err = e.AddRoleForUser("charlie", "admin")
	if err == nil {
		t.Fatal("Expected constraint violation error, got nil")
	}
	if !strings.Contains(err.Error(), "constraint violation") {
		t.Fatalf("Expected constraint violation error, got: %v", err)
	}
}

func TestConstraintRolePre(t *testing.T) {
	modelText := `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[constraint_definition]
c = rolePre("db_admin", "security_trained")

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act
`

	m, err := model.NewModelFromString(modelText)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	e, err := NewEnforcer(m)
	if err != nil {
		t.Fatalf("Failed to create enforcer: %v", err)
	}

	// Try to add db_admin without prerequisite should fail
	_, err = e.AddRoleForUser("alice", "db_admin")
	if err == nil {
		t.Fatal("Expected constraint violation error, got nil")
	}
	if !strings.Contains(err.Error(), "constraint violation") {
		t.Fatalf("Expected constraint violation error, got: %v", err)
	}

	// Add prerequisite role first
	_, err = e.AddRoleForUser("alice", "security_trained")
	if err != nil {
		t.Fatalf("Failed to add security_trained to alice: %v", err)
	}

	// Now adding db_admin should succeed
	_, err = e.AddRoleForUser("alice", "db_admin")
	if err != nil {
		t.Fatalf("Failed to add db_admin to alice: %v", err)
	}
}

func TestConstraintWithoutRBAC(t *testing.T) {
	modelText := `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[constraint_definition]
c = sod("role1", "role2")

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = r.sub == p.sub && r.obj == p.obj && r.act == p.act
`

	_, err := model.NewModelFromString(modelText)
	if err == nil {
		t.Fatal("Expected error for constraints without RBAC, got nil")
	}
	if err != errors.ErrConstraintRequiresRBAC {
		t.Fatalf("Expected ErrConstraintRequiresRBAC, got: %v", err)
	}
}

func TestConstraintParsingError(t *testing.T) {
	modelText := `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[constraint_definition]
c = invalidFunction("role1", "role2")

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act
`

	_, err := model.NewModelFromString(modelText)
	if err == nil {
		t.Fatal("Expected parsing error for invalid constraint, got nil")
	}
	if !strings.Contains(err.Error(), "constraint parsing error") {
		t.Fatalf("Expected constraint parsing error, got: %v", err)
	}
}

func TestConstraintRollback(t *testing.T) {
	modelText := `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[constraint_definition]
c = sod("role1", "role2")

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act
`

	m, err := model.NewModelFromString(modelText)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	e, err := NewEnforcer(m)
	if err != nil {
		t.Fatalf("Failed to create enforcer: %v", err)
	}

	// Add alice to role1
	_, err = e.AddRoleForUser("alice", "role1")
	if err != nil {
		t.Fatalf("Failed to add role1 to alice: %v", err)
	}

	// Try to add alice to role2 (should fail with constraint violation)
	_, err = e.AddRoleForUser("alice", "role2")
	if err == nil {
		t.Fatal("Expected constraint violation error, got nil")
	}
	if !strings.Contains(err.Error(), "constraint violation") {
		t.Fatalf("Expected constraint violation error, got: %v", err)
	}
}
