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
	"testing"

	fileadapter "github.com/casbin/casbin/v3/persist/file-adapter"
	"github.com/casbin/casbin/v3/util"
)

func TestAIPolicyLoad(t *testing.T) {
	e, err := NewEnforcer("examples/ai_policy_model.conf", "examples/ai_policy.csv")
	if err != nil {
		t.Fatal(err)
	}

	// Test that regular policies are loaded
	policies, err := e.GetPolicy()
	if err != nil {
		t.Fatal(err)
	}

	expectedPolicies := [][]string{
		{"alice", "data1", "read", "09:00", "18:00"},
		{"bob", "data2", "write", "13:00", "16:00"},
	}

	if !util.Array2DEquals(expectedPolicies, policies) {
		t.Errorf("Policies = %v, want %v", policies, expectedPolicies)
	}

	// Test that grouping policies are loaded
	groupingPolicies, err := e.GetGroupingPolicy()
	if err != nil {
		t.Fatal(err)
	}

	expectedGrouping := [][]string{
		{"cathy", "alice"},
	}

	if !util.Array2DEquals(expectedGrouping, groupingPolicies) {
		t.Errorf("Grouping policies = %v, want %v", groupingPolicies, expectedGrouping)
	}

	// Test that AI policies are loaded
	aiPolicies, err := e.model.GetPolicy("a", "ai")
	if err != nil {
		t.Fatal(err)
	}

	expectedAI := [][]string{
		{`if the request object contains anything like credential/secret leak, then deny`},
	}

	if !util.Array2DEquals(expectedAI, aiPolicies) {
		t.Errorf("AI policies = %v, want %v", aiPolicies, expectedAI)
	}
}

func TestAIPolicySave(t *testing.T) {
	// Create a temporary file for testing
	tmpFile := t.TempDir() + "/ai_policy_test.csv"

	e, err := NewEnforcer("examples/ai_policy_model.conf", "examples/ai_policy.csv")
	if err != nil {
		t.Fatal(err)
	}

	// Update adapter to save to temp file
	e.SetAdapter(fileadapter.NewAdapter(tmpFile))
	
	// Save to the temporary file
	err = e.SavePolicy()
	if err != nil {
		t.Fatal(err)
	}

	// Load from the saved file
	e2, err := NewEnforcer("examples/ai_policy_model.conf", tmpFile)
	if err != nil {
		t.Fatal(err)
	}

	// Verify AI policies are preserved
	aiPolicies, err := e2.model.GetPolicy("a", "ai")
	if err != nil {
		t.Fatal(err)
	}

	expectedAI := [][]string{
		{`if the request object contains anything like credential/secret leak, then deny`},
	}

	if !util.Array2DEquals(expectedAI, aiPolicies) {
		t.Errorf("AI policies after save/load = %v, want %v", aiPolicies, expectedAI)
	}
}
