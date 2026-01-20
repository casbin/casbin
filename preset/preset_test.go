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

package preset

import (
	"testing"

	"github.com/casbin/casbin/v3"
	fileadapter "github.com/casbin/casbin/v3/persist/file-adapter"
)

func TestRBACPreset(t *testing.T) {
	m := RBAC()
	if m == nil {
		t.Fatal("RBAC() returned nil model")
	}

	// Create enforcer with preset model and a file adapter
	adapter := fileadapter.NewAdapter("../examples/rbac_policy.csv")
	e, err := casbin.NewEnforcer(m, adapter)
	if err != nil {
		t.Fatalf("Failed to create enforcer: %v", err)
	}

	// Test basic enforcement
	ok, err := e.Enforce("alice", "data1", "read")
	if err != nil {
		t.Fatalf("Enforce failed: %v", err)
	}
	if !ok {
		t.Error("alice should be able to read data1")
	}

	ok, err = e.Enforce("alice", "data2", "read")
	if err != nil {
		t.Fatalf("Enforce failed: %v", err)
	}
	if !ok {
		t.Error("alice should be able to read data2 (via data2_admin role)")
	}

	ok, err = e.Enforce("bob", "data2", "write")
	if err != nil {
		t.Fatalf("Enforce failed: %v", err)
	}
	if !ok {
		t.Error("bob should be able to write data2")
	}

	ok, err = e.Enforce("bob", "data1", "read")
	if err != nil {
		t.Fatalf("Enforce failed: %v", err)
	}
	if ok {
		t.Error("bob should not be able to read data1")
	}
}

func TestRBACPresetWithoutPolicies(t *testing.T) {
	m := RBAC()
	e, err := casbin.NewEnforcer(m)
	if err != nil {
		t.Fatalf("Failed to create enforcer: %v", err)
	}

	// Initially, no permissions should be granted
	ok, err := e.Enforce("alice", "data1", "read")
	if err != nil {
		t.Fatalf("Enforce failed: %v", err)
	}
	if ok {
		t.Error("alice should not have any permissions initially")
	}

	// Add a policy
	_, err = e.AddPolicy("alice", "data1", "read")
	if err != nil {
		t.Fatalf("AddPolicy failed: %v", err)
	}

	// Now alice should have permission
	ok, err = e.Enforce("alice", "data1", "read")
	if err != nil {
		t.Fatalf("Enforce failed: %v", err)
	}
	if !ok {
		t.Error("alice should be able to read data1 after adding policy")
	}

	// Add role
	_, err = e.AddRoleForUser("bob", "alice")
	if err != nil {
		t.Fatalf("AddRoleForUser failed: %v", err)
	}

	// Bob should inherit alice's permission
	ok, err = e.Enforce("bob", "data1", "read")
	if err != nil {
		t.Fatalf("Enforce failed: %v", err)
	}
	if !ok {
		t.Error("bob should be able to read data1 (via alice role)")
	}
}

func TestRBACPresetModelEquivalence(t *testing.T) {
	// Test that the preset model is equivalent to the standard rbac_model.conf
	presetModel := RBAC()
	presetAdapter := fileadapter.NewAdapter("../examples/rbac_policy.csv")
	presetEnforcer, err := casbin.NewEnforcer(presetModel, presetAdapter)
	if err != nil {
		t.Fatalf("Failed to create preset enforcer: %v", err)
	}

	fileEnforcer, err := casbin.NewEnforcer("../examples/rbac_model.conf", "../examples/rbac_policy.csv")
	if err != nil {
		t.Fatalf("Failed to create file enforcer: %v", err)
	}

	// Test cases that should behave the same
	testCases := []struct {
		sub string
		obj string
		act string
	}{
		{"alice", "data1", "read"},
		{"alice", "data2", "read"},
		{"alice", "data2", "write"},
		{"bob", "data2", "write"},
		{"bob", "data1", "read"},
		{"data2_admin", "data2", "read"},
	}

	for _, tc := range testCases {
		presetOk, err := presetEnforcer.Enforce(tc.sub, tc.obj, tc.act)
		if err != nil {
			t.Fatalf("Preset enforce failed for %v: %v", tc, err)
		}

		fileOk, err := fileEnforcer.Enforce(tc.sub, tc.obj, tc.act)
		if err != nil {
			t.Fatalf("File enforce failed for %v: %v", tc, err)
		}

		if presetOk != fileOk {
			t.Errorf("Mismatch for %v: preset=%v, file=%v", tc, presetOk, fileOk)
		}
	}
}
