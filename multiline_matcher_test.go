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

func TestMultiLineMatcherWithLetStatements(t *testing.T) {
	e, err := NewEnforcer("examples/rbac_with_hierarchy_multiline_model.conf", "examples/rbac_with_hierarchy_multiline_policy.csv")
	if err != nil {
		t.Fatalf("Failed to create enforcer: %v", err)
	}

	// alice has direct permission on data1 for read
	testEnforce(t, e, "alice", "data1", "read", true)
	
	// alice doesn't have direct permission on data1 for write, but has via role and resource hierarchy
	testEnforce(t, e, "alice", "data1", "write", true)
	
	// bob has direct permission on data2 for write
	testEnforce(t, e, "bob", "data2", "write", true)
	
	// bob doesn't have direct permission on data1
	testEnforce(t, e, "bob", "data1", "read", false)
	testEnforce(t, e, "bob", "data1", "write", false)
	
	// Test with inherited permissions through data_group
	testEnforce(t, e, "alice", "data2", "write", true)
}

func TestMultiLineMatcherWithEarlyReturn(t *testing.T) {
	e, err := NewEnforcer("examples/rbac_with_early_return_model.conf", "examples/rbac_with_hierarchy_multiline_policy.csv")
	if err != nil {
		t.Fatalf("Failed to create enforcer: %v", err)
	}

	// alice has direct permission on data1 for read
	testEnforce(t, e, "alice", "data1", "read", true)
	
	// alice doesn't have direct permission on data1 for write, but has via role
	testEnforce(t, e, "alice", "data1", "write", true)
	
	// bob has direct permission on data2 for write
	testEnforce(t, e, "bob", "data2", "write", true)
	
	// bob doesn't have permission on data1
	testEnforce(t, e, "bob", "data1", "read", false)
	testEnforce(t, e, "bob", "data1", "write", false)
	
	// alice can write to data2 through role and resource hierarchy
	testEnforce(t, e, "alice", "data2", "write", true)
}

func TestMultiLineMatcherInMemory(t *testing.T) {
	m := model.NewModel()
	m.AddDef("r", "r", "sub, obj, act")
	m.AddDef("p", "p", "sub, obj, act")
	m.AddDef("g", "g", "_, _")
	m.AddDef("e", "e", "some(where (p.eft == allow))")
	m.AddDef("m", "m", `{
		let role_match = g(r.sub, p.sub)
		let obj_match = r.obj == p.obj
		let act_match = r.act == p.act
		return role_match && obj_match && act_match
	}`)

	e, err := NewEnforcer(m)
	if err != nil {
		t.Fatalf("Failed to create enforcer: %v", err)
	}

	// Add policies
	_, _ = e.AddPolicy("alice", "data1", "read")
	_, _ = e.AddPolicy("data_admin", "data2", "write")
	_, _ = e.AddGroupingPolicy("bob", "data_admin")

	// Test enforcement
	testEnforce(t, e, "alice", "data1", "read", true)
	testEnforce(t, e, "alice", "data1", "write", false)
	testEnforce(t, e, "bob", "data2", "write", true)
	testEnforce(t, e, "bob", "data1", "read", false)
}

func TestSimpleBlockMatcher(t *testing.T) {
	m := model.NewModel()
	m.AddDef("r", "r", "sub, obj, act")
	m.AddDef("p", "p", "sub, obj, act")
	m.AddDef("e", "e", "some(where (p.eft == allow))")
	m.AddDef("m", "m", `{
		return r.sub == p.sub && r.obj == p.obj && r.act == p.act
	}`)

	e, err := NewEnforcer(m)
	if err != nil {
		t.Fatalf("Failed to create enforcer: %v", err)
	}

	_, _ = e.AddPolicy("alice", "data1", "read")
	_, _ = e.AddPolicy("bob", "data2", "write")

	testEnforce(t, e, "alice", "data1", "read", true)
	testEnforce(t, e, "alice", "data1", "write", false)
	testEnforce(t, e, "bob", "data2", "write", true)
	testEnforce(t, e, "bob", "data1", "read", false)
}
