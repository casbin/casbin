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
	"fmt"
	"testing"
)

// BenchmarkEnforcementPerformance tests the performance improvements
// from precompiling matcher expressions and caching token maps
func BenchmarkEnforcementPerformance(b *testing.B) {
	e, _ := NewEnforcer("examples/rbac_model.conf", "examples/rbac_policy.csv", false)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = e.Enforce("alice", "data2", "read")
	}
}

// BenchmarkEnforcementWithMultipleContexts tests enforcement with different request types
func BenchmarkEnforcementWithMultipleContexts(b *testing.B) {
	e, _ := NewEnforcer("examples/rbac_model.conf", "examples/rbac_policy.csv", false)

	requests := [][]interface{}{
		{"alice", "data1", "read"},
		{"alice", "data2", "read"},
		{"bob", "data1", "read"},
		{"bob", "data2", "write"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := requests[i%len(requests)]
		_, _ = e.Enforce(req...)
	}
}

// BenchmarkEnforcementScalability tests enforcement performance at different scales
func BenchmarkEnforcementScalability(b *testing.B) {
	scales := []struct {
		name   string
		groups int
		users  int
	}{
		{"Small", 10, 100},
		{"Medium", 100, 1000},
		{"Large", 1000, 10000},
	}

	for _, scale := range scales {
		b.Run(scale.name, func(b *testing.B) {
			e, _ := NewEnforcer("examples/rbac_model.conf", false)

			// Add policies
			for i := 0; i < scale.groups; i++ {
				_, _ = e.AddPolicy(fmt.Sprintf("group%d", i), fmt.Sprintf("data%d", i%10), "read")
			}

			// Add users
			for i := 0; i < scale.users; i++ {
				_, _ = e.AddGroupingPolicy(fmt.Sprintf("user%d", i), fmt.Sprintf("group%d", i%scale.groups))
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = e.Enforce(fmt.Sprintf("user%d", i%scale.users), "data5", "read")
			}
		})
	}
}
