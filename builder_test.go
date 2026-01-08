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

// TestBuilderRBAC tests the builder version of the RBAC model.
func TestBuilderRBAC(t *testing.T) {
	// Build the RBAC model programmatically
	m, _ := model.New().
		Request("sub", "obj", "act").
		Policy("sub", "obj", "act").
		Role("_", "_").
		Effect("some(where (p.eft == allow))").
		Matcher("g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act").
		Build()

	// Use the same policy file as the original test
	a := fileadapter.NewAdapter("examples/rbac_policy.csv")
	e, _ := NewEnforcer(m, a)

	// Same test cases as TestRBACModel
	testEnforce(t, e, "alice", "data1", "read", true)
	testEnforce(t, e, "alice", "data1", "write", false)
	testEnforce(t, e, "alice", "data2", "read", true)
	testEnforce(t, e, "alice", "data2", "write", true)
	testEnforce(t, e, "bob", "data1", "read", false)
	testEnforce(t, e, "bob", "data1", "write", false)
	testEnforce(t, e, "bob", "data2", "read", false)
	testEnforce(t, e, "bob", "data2", "write", true)
}

// TestBuilderABAC tests the builder version of the ABAC model.
func TestBuilderABAC(t *testing.T) {
	// Build the ABAC model programmatically
	m, _ := model.New().
		Request("sub", "obj", "act").
		Policy("sub", "obj", "act").
		Effect("some(where (p.eft == allow))").
		Matcher("r.sub == r.obj.Owner").
		Build()

	e, _ := NewEnforcer(m)

	data1 := newTestResource("data1", "alice")
	data2 := newTestResource("data2", "bob")

	// Same test cases as TestABACModel
	testEnforce(t, e, "alice", data1, "read", true)
	testEnforce(t, e, "alice", data1, "write", true)
	testEnforce(t, e, "alice", data2, "read", false)
	testEnforce(t, e, "alice", data2, "write", false)
	testEnforce(t, e, "bob", data1, "read", false)
	testEnforce(t, e, "bob", data1, "write", false)
	testEnforce(t, e, "bob", data2, "read", true)
	testEnforce(t, e, "bob", data2, "write", true)
}
