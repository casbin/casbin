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
	"testing"

	"github.com/casbin/casbin/v3/model"
	fileadapter "github.com/casbin/casbin/v3/persist/file-adapter"
)

// TestBuilderRBACModel tests building an RBAC model programmatically.
func TestBuilderRBACModel(t *testing.T) {
	// Build RBAC model using the builder
	m, err := model.NewBuilder().
		Request("sub", "obj", "act").
		Policy("sub", "obj", "act").
		RoleDefinition("_", "_").
		Effect(model.AllowOverride).
		Matcher(model.And(model.G("r.sub", "p.sub"), model.Eq("obj"), model.Eq("act"))).
		Build()

	if err != nil {
		t.Fatalf("Failed to build model: %v", err)
	}

	// Create enforcer with the built model and policy file adapter
	e, err := NewEnforcer(m, fileadapter.NewAdapter("examples/rbac_policy.csv"))
	if err != nil {
		t.Fatalf("Failed to create enforcer: %v", err)
	}

	// Run the same tests as TestRBACModel
	testEnforce(t, e, "alice", "data1", "read", true)
	testEnforce(t, e, "alice", "data1", "write", false)
	testEnforce(t, e, "alice", "data2", "read", true)
	testEnforce(t, e, "alice", "data2", "write", true)
	testEnforce(t, e, "bob", "data1", "read", false)
	testEnforce(t, e, "bob", "data1", "write", false)
	testEnforce(t, e, "bob", "data2", "read", false)
	testEnforce(t, e, "bob", "data2", "write", true)
}

// TestBuilderABACModel tests building an ABAC model programmatically.
func TestBuilderABACModel(t *testing.T) {
	// Build ABAC model using the builder
	m, err := model.NewBuilder().
		Request("sub", "obj", "act").
		Policy("sub", "obj", "act").
		Effect(model.AllowOverride).
		Matcher("r.sub == r.obj.Owner").
		Build()

	if err != nil {
		t.Fatalf("Failed to build model: %v", err)
	}

	// Create enforcer with the built model
	e, err := NewEnforcer(m)
	if err != nil {
		t.Fatalf("Failed to create enforcer: %v", err)
	}

	// Run the same tests as TestABACModel
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
