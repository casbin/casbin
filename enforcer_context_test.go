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
	"context"
	"testing"
	"time"
)

func TestIEnforcerContext_BasicOperations(t *testing.T) {
	e, err := NewContextEnforcer("examples/rbac_model.conf", "examples/rbac_policy.csv")
	if err != nil {
		t.Fatalf("NewContextEnforcer failed: %v", err)
	}

	ctx := context.Background()

	err = e.LoadPolicyCtx(ctx)
	if err != nil {
		t.Fatalf("LoadPolicyCtx failed: %v", err)
	}

	added, err := e.AddPolicyCtx(ctx, "eve", "data3", "read")
	if err != nil {
		t.Fatalf("AddPolicyCtx failed: %v", err)
	}
	if !added {
		t.Error("AddPolicyCtx should return true for new policy")
	}

	removed, err := e.RemovePolicyCtx(ctx, "eve", "data3", "read")
	if err != nil {
		t.Fatalf("RemovePolicyCtx failed: %v", err)
	}
	if !removed {
		t.Error("RemovePolicyCtx should return true for existing policy")
	}

	err = e.SavePolicyCtx(ctx)
	if err != nil {
		t.Fatalf("SavePolicyCtx failed: %v", err)
	}
}

func TestIEnforcerContext_RBACOperations(t *testing.T) {
	e, err := NewContextEnforcer("examples/rbac_model.conf", "examples/rbac_policy.csv")
	if err != nil {
		t.Fatalf("NewContextEnforcer failed: %v", err)
	}

	ctx := context.Background()

	added, err := e.AddRoleForUserCtx(ctx, "eve", "data1_admin")
	if err != nil {
		t.Fatalf("AddRoleForUserCtx failed: %v", err)
	}
	if !added {
		t.Error("AddRoleForUserCtx should return true for new role")
	}

	added, err = e.AddPermissionForUserCtx(ctx, "eve", "data3", "write")
	if err != nil {
		t.Fatalf("AddPermissionForUserCtx failed: %v", err)
	}
	if !added {
		t.Error("AddPermissionForUserCtx should return true for new permission")
	}

	deleted, err := e.DeleteRoleForUserCtx(ctx, "eve", "data1_admin")
	if err != nil {
		t.Fatalf("DeleteRoleForUserCtx failed: %v", err)
	}
	if !deleted {
		t.Error("DeleteRoleForUserCtx should return true for existing role")
	}

	deleted, err = e.DeletePermissionForUserCtx(ctx, "eve", "data3", "write")
	if err != nil {
		t.Fatalf("DeletePermissionForUserCtx failed: %v", err)
	}
	if !deleted {
		t.Error("DeletePermissionForUserCtx should return true for existing permission")
	}
}

func TestIEnforcerContext_BatchOperations(t *testing.T) {
	e, err := NewContextEnforcer("examples/rbac_model.conf", "examples/rbac_policy.csv")
	if err != nil {
		t.Fatalf("NewContextEnforcer failed: %v", err)
	}

	ctx := context.Background()

	rules := [][]string{
		{"eve", "data3", "read"},
		{"eve", "data3", "write"},
	}
	added, err := e.AddPoliciesCtx(ctx, rules)
	if err != nil {
		t.Fatalf("AddPoliciesCtx failed: %v", err)
	}
	if !added {
		t.Error("AddPoliciesCtx should return true for new policies")
	}

	removed, err := e.RemovePoliciesCtx(ctx, rules)
	if err != nil {
		t.Fatalf("RemovePoliciesCtx failed: %v", err)
	}
	if !removed {
		t.Error("RemovePoliciesCtx should return true for existing policies")
	}
}

func TestIEnforcerContext_ContextCancellation(t *testing.T) {
	e, err := NewContextEnforcer("examples/rbac_model.conf", "examples/rbac_policy.csv")
	if err != nil {
		t.Fatalf("NewContextEnforcer failed: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err = e.LoadPolicyCtx(ctx)
	if err != nil {
		t.Logf("LoadPolicyCtx with cancelled context returned error: %v", err)
	}
}

func TestIEnforcerContext_ContextTimeout(t *testing.T) {
	e, err := NewContextEnforcer("examples/rbac_model.conf", "examples/rbac_policy.csv")
	if err != nil {
		t.Fatalf("NewContextEnforcer failed: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	time.Sleep(2 * time.Millisecond)

	_, err = e.AddPolicyCtx(ctx, "test", "data", "read")
	if err != nil {
		t.Logf("AddPolicyCtx with timeout context returned error: %v", err)
	}
}

func TestIEnforcerContext_GroupingPolicyOperations(t *testing.T) {
	e, err := NewContextEnforcer("examples/rbac_model.conf", "examples/rbac_policy.csv")
	if err != nil {
		t.Fatalf("NewContextEnforcer failed: %v", err)
	}

	ctx := context.Background()

	added, err := e.AddGroupingPolicyCtx(ctx, "eve", "data3_admin")
	if err != nil {
		t.Fatalf("AddGroupingPolicyCtx failed: %v", err)
	}
	if !added {
		t.Error("AddGroupingPolicyCtx should return true for new grouping policy")
	}

	removed, err := e.RemoveGroupingPolicyCtx(ctx, "eve", "data3_admin")
	if err != nil {
		t.Fatalf("RemoveGroupingPolicyCtx failed: %v", err)
	}
	if !removed {
		t.Error("RemoveGroupingPolicyCtx should return true for existing grouping policy")
	}
}

func TestIEnforcerContext_UpdateOperations(t *testing.T) {
	e, err := NewContextEnforcer("examples/rbac_model.conf", "examples/rbac_policy.csv")
	if err != nil {
		t.Fatalf("NewContextEnforcer failed: %v", err)
	}

	ctx := context.Background()

	_, err = e.AddPolicyCtx(ctx, "eve", "data3", "read")
	if err != nil {
		t.Fatalf("AddPolicyCtx failed: %v", err)
	}

	updated, err := e.UpdatePolicyCtx(ctx, []string{"eve", "data3", "read"}, []string{"eve", "data3", "write"})
	if err != nil {
		t.Fatalf("UpdatePolicyCtx failed: %v", err)
	}
	if !updated {
		t.Error("UpdatePolicyCtx should return true for successful update")
	}

	_, _ = e.RemovePolicyCtx(ctx, "eve", "data3", "write")
}

func TestIEnforcerContext_SelfMethods(t *testing.T) {
	e, err := NewContextEnforcer("examples/rbac_model.conf", "examples/rbac_policy.csv")
	if err != nil {
		t.Fatalf("NewContextEnforcer failed: %v", err)
	}

	ctx := context.Background()

	added, err := e.SelfAddPolicyCtx(ctx, "p", "p", []string{"eve", "data3", "read"})
	if err != nil {
		t.Fatalf("SelfAddPolicyCtx failed: %v", err)
	}
	if !added {
		t.Error("SelfAddPolicyCtx should return true for new policy")
	}

	removed, err := e.SelfRemovePolicyCtx(ctx, "p", "p", []string{"eve", "data3", "read"})
	if err != nil {
		t.Fatalf("SelfRemovePolicyCtx failed: %v", err)
	}
	if !removed {
		t.Error("SelfRemovePolicyCtx should return true for existing policy")
	}
}
