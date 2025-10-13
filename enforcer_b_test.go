// Copyright 2025 The casbin Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package casbin

import (
	"fmt"
	"testing"
)

func BenchmarkBatchEnforceSmall(b *testing.B) {
	e, _ := NewEnforcer("examples/rbac_model.conf", "examples/rbac_policy.csv")
	
	// Create 10 batch requests
	requests := make([][]interface{}, 10)
	for i := 0; i < 10; i++ {
		requests[i] = []interface{}{"alice", fmt.Sprintf("data%d", i%3), "read"}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = e.BatchEnforce(requests)
	}
}

func BenchmarkBatchEnforceMedium(b *testing.B) {
	e, _ := NewEnforcer("examples/rbac_model.conf", false)
	
	// 100 roles, 10 resources
	for i := 0; i < 100; i++ {
		_, _ = e.AddPolicy(fmt.Sprintf("group%d", i), fmt.Sprintf("data%d", i/10), "read")
	}
	// 1000 users
	for i := 0; i < 1000; i++ {
		_, _ = e.AddGroupingPolicy(fmt.Sprintf("user%d", i), fmt.Sprintf("group%d", i/10))
	}

	// Create 100 batch requests
	requests := make([][]interface{}, 100)
	for i := 0; i < 100; i++ {
		requests[i] = []interface{}{fmt.Sprintf("user%d", i), fmt.Sprintf("data%d", i/10), "read"}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = e.BatchEnforce(requests)
	}
}

func BenchmarkBatchEnforceLarge(b *testing.B) {
	e, _ := NewEnforcer("examples/rbac_model.conf", false)
	
	// 1000 roles, 100 resources
	pPolicies := make([][]string, 0)
	for i := 0; i < 1000; i++ {
		pPolicies = append(pPolicies, []string{fmt.Sprintf("group%d", i), fmt.Sprintf("data%d", i/10), "read"})
	}
	_, _ = e.AddPolicies(pPolicies)
	
	// 10000 users
	gPolicies := make([][]string, 0)
	for i := 0; i < 10000; i++ {
		gPolicies = append(gPolicies, []string{fmt.Sprintf("user%d", i), fmt.Sprintf("group%d", i/10)})
	}
	_, _ = e.AddGroupingPolicies(gPolicies)

	// Create 1000 batch requests
	requests := make([][]interface{}, 1000)
	for i := 0; i < 1000; i++ {
		requests[i] = []interface{}{fmt.Sprintf("user%d", i), fmt.Sprintf("data%d", i/100), "read"}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = e.BatchEnforce(requests)
	}
}

// BenchmarkEnforceVsBatchEnforce compares individual Enforce vs BatchEnforce
func BenchmarkEnforceVsBatchEnforce(b *testing.B) {
	e, _ := NewEnforcer("examples/rbac_model.conf", false)
	
	// Setup: 100 roles, 10 resources
	for i := 0; i < 100; i++ {
		_, _ = e.AddPolicy(fmt.Sprintf("group%d", i), fmt.Sprintf("data%d", i/10), "read")
	}
	for i := 0; i < 1000; i++ {
		_, _ = e.AddGroupingPolicy(fmt.Sprintf("user%d", i), fmt.Sprintf("group%d", i/10))
	}

	// Create 100 requests for testing
	requests := make([][]interface{}, 100)
	for i := 0; i < 100; i++ {
		requests[i] = []interface{}{fmt.Sprintf("user%d", i), fmt.Sprintf("data%d", i/10), "read"}
	}

	b.Run("IndividualEnforce", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, request := range requests {
				_, _ = e.Enforce(request...)
			}
		}
	})

	b.Run("BatchEnforce", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = e.BatchEnforce(requests)
		}
	})
}
