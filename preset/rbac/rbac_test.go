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

package rbac

import (
	"testing"

	casbin "github.com/casbin/casbin/v3"
	"github.com/casbin/casbin/v3/preset"
)

func TestAssignRole(t *testing.T) {
	m := preset.RBAC()
	e, err := casbin.NewEnforcer(m)
	if err != nil {
		t.Fatalf("Failed to create enforcer: %v", err)
	}

	// Assign a role to alice
	ok, err := AssignRole(e, "alice", "admin")
	if err != nil {
		t.Fatalf("AssignRole failed: %v", err)
	}
	if !ok {
		t.Error("AssignRole should return true when adding a new role")
	}

	// Verify the role was assigned
	roles, err := e.GetRolesForUser("alice")
	if err != nil {
		t.Fatalf("GetRolesForUser failed: %v", err)
	}
	if len(roles) != 1 || roles[0] != "admin" {
		t.Errorf("Expected alice to have admin role, got %v", roles)
	}

	// Try to assign the same role again
	ok, err = AssignRole(e, "alice", "admin")
	if err != nil {
		t.Fatalf("AssignRole failed: %v", err)
	}
	if ok {
		t.Error("AssignRole should return false when role already exists")
	}
}

func TestGrant(t *testing.T) {
	m := preset.RBAC()
	e, err := casbin.NewEnforcer(m)
	if err != nil {
		t.Fatalf("Failed to create enforcer: %v", err)
	}

	// Grant permission to admin role
	ok, err := Grant(e, "admin", "data1", "read")
	if err != nil {
		t.Fatalf("Grant failed: %v", err)
	}
	if !ok {
		t.Error("Grant should return true when adding a new permission")
	}

	// Verify the permission was granted
	ok, err = e.Enforce("admin", "data1", "read")
	if err != nil {
		t.Fatalf("Enforce failed: %v", err)
	}
	if !ok {
		t.Error("admin should be able to read data1 after grant")
	}

	// Try to grant the same permission again
	ok, err = Grant(e, "admin", "data1", "read")
	if err != nil {
		t.Fatalf("Grant failed: %v", err)
	}
	if ok {
		t.Error("Grant should return false when permission already exists")
	}
}

func TestIntegrationAssignRoleAndGrant(t *testing.T) {
	m := preset.RBAC()
	e, err := casbin.NewEnforcer(m)
	if err != nil {
		t.Fatalf("Failed to create enforcer: %v", err)
	}

	// Grant permission to admin role
	_, err = Grant(e, "admin", "data1", "read")
	if err != nil {
		t.Fatalf("Grant failed: %v", err)
	}

	// Assign admin role to alice
	_, err = AssignRole(e, "alice", "admin")
	if err != nil {
		t.Fatalf("AssignRole failed: %v", err)
	}

	// Alice should be able to read data1 through the admin role
	ok, err := e.Enforce("alice", "data1", "read")
	if err != nil {
		t.Fatalf("Enforce failed: %v", err)
	}
	if !ok {
		t.Error("alice should be able to read data1 (via admin role)")
	}

	// Alice should not have direct permission
	hasPolicy, err := e.HasPolicy("alice", "data1", "read")
	if err != nil {
		t.Fatalf("HasPolicy failed: %v", err)
	}
	if hasPolicy {
		t.Error("alice should not have direct policy")
	}

	// But she should have it through the role
	hasRole, err := e.HasRoleForUser("alice", "admin")
	if err != nil {
		t.Fatalf("HasRoleForUser failed: %v", err)
	}
	if !hasRole {
		t.Error("alice should have admin role")
	}
}

func TestGrantToUserDirectly(t *testing.T) {
	m := preset.RBAC()
	e, err := casbin.NewEnforcer(m)
	if err != nil {
		t.Fatalf("Failed to create enforcer: %v", err)
	}

	// Grant permission directly to a user (not a role)
	_, err = Grant(e, "alice", "data1", "read")
	if err != nil {
		t.Fatalf("Grant failed: %v", err)
	}

	// Alice should be able to read data1
	ok, err := e.Enforce("alice", "data1", "read")
	if err != nil {
		t.Fatalf("Enforce failed: %v", err)
	}
	if !ok {
		t.Error("alice should be able to read data1 after direct grant")
	}

	// Alice should not have any roles
	roles, err := e.GetRolesForUser("alice")
	if err != nil {
		t.Fatalf("GetRolesForUser failed: %v", err)
	}
	if len(roles) != 0 {
		t.Errorf("alice should not have any roles, got %v", roles)
	}
}

func TestExampleFromIssue(t *testing.T) {
	// This test demonstrates the example usage from the issue description
	m := preset.RBAC()
	e, err := casbin.NewEnforcer(m)
	if err != nil {
		t.Fatalf("Failed to create enforcer: %v", err)
	}

	// Use the optional helpers
	_, err = AssignRole(e, "alice", "admin")
	if err != nil {
		t.Fatalf("AssignRole failed: %v", err)
	}

	_, err = Grant(e, "admin", "data1", "read")
	if err != nil {
		t.Fatalf("Grant failed: %v", err)
	}

	// Test enforcement
	ok, err := e.Enforce("alice", "data1", "read")
	if err != nil {
		t.Fatalf("Enforce failed: %v", err)
	}
	if !ok {
		t.Error("alice should be able to read data1 (as shown in the issue example)")
	}
}
