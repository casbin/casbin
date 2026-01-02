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

	"github.com/casbin/casbin/v3/model"
)

func TestBuilderWithEnforcer(t *testing.T) {
	// Build a model programmatically
	m, err := model.NewBuilder().
		Request("sub", "obj", "act").
		Policy("sub", "obj", "act").
		Effect(model.AllowOverride).
		Matcher(model.And(model.Eq("sub"), model.Eq("obj"), model.Eq("act"))).
		Build()

	if err != nil {
		t.Fatalf("Build() failed: %v", err)
	}

	// Create an enforcer with the built model
	e, err := NewEnforcer(m)
	if err != nil {
		t.Fatalf("Failed to create enforcer: %v", err)
	}

	// Add some policies
	_, err = e.AddPolicy("alice", "data1", "read")
	if err != nil {
		t.Fatalf("Failed to add policy: %v", err)
	}
	_, err = e.AddPolicy("bob", "data2", "write")
	if err != nil {
		t.Fatalf("Failed to add policy: %v", err)
	}

	// Test enforcement
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
	if ok {
		t.Error("alice should not be able to read data2")
	}

	ok, err = e.Enforce("bob", "data2", "write")
	if err != nil {
		t.Fatalf("Enforce failed: %v", err)
	}
	if !ok {
		t.Error("bob should be able to write data2")
	}
}

func TestBuilderRBACWithEnforcer(t *testing.T) {
	// Build an RBAC model programmatically
	m, err := model.NewBuilder().
		Request("sub", "obj", "act").
		Policy("sub", "obj", "act").
		RoleDefinition("_", "_").
		Effect(model.AllowOverride).
		Matcher(model.And(model.G("r.sub", "p.sub"), model.Eq("obj"), model.Eq("act"))).
		Build()

	if err != nil {
		t.Fatalf("Build() failed: %v", err)
	}

	// Create an enforcer with the built RBAC model
	e, err := NewEnforcer(m)
	if err != nil {
		t.Fatalf("Failed to create enforcer: %v", err)
	}

	// Add policies and roles
	_, err = e.AddPolicy("admin", "data", "read")
	if err != nil {
		t.Fatalf("Failed to add policy: %v", err)
	}
	_, err = e.AddPolicy("admin", "data", "write")
	if err != nil {
		t.Fatalf("Failed to add policy: %v", err)
	}
	_, err = e.AddGroupingPolicy("alice", "admin")
	if err != nil {
		t.Fatalf("Failed to add grouping policy: %v", err)
	}

	// Test enforcement with role inheritance
	ok, err := e.Enforce("alice", "data", "read")
	if err != nil {
		t.Fatalf("Enforce failed: %v", err)
	}
	if !ok {
		t.Error("alice should be able to read data through admin role")
	}

	ok, err = e.Enforce("alice", "data", "write")
	if err != nil {
		t.Fatalf("Enforce failed: %v", err)
	}
	if !ok {
		t.Error("alice should be able to write data through admin role")
	}

	ok, err = e.Enforce("bob", "data", "read")
	if err != nil {
		t.Fatalf("Enforce failed: %v", err)
	}
	if ok {
		t.Error("bob should not be able to read data")
	}
}
