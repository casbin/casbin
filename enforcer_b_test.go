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
)

func BenchmarkBatchEnforce(b *testing.B) {
	e, _ := NewEnforcer("examples/rbac_model.conf", "examples/rbac_policy.csv", false)
	
	// Create a batch of 100 requests
	requests := make([][]interface{}, 100)
	for i := 0; i < 100; i++ {
		if i%3 == 0 {
			requests[i] = []interface{}{"alice", "data1", "read"}
		} else if i%3 == 1 {
			requests[i] = []interface{}{"alice", "data1", "write"}
		} else {
			requests[i] = []interface{}{"bob", "data2", "read"}
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = e.BatchEnforce(requests)
	}
}

func BenchmarkBatchEnforceSmall(b *testing.B) {
	e, _ := NewEnforcer("examples/rbac_model.conf", "examples/rbac_policy.csv", false)
	
	// Create a batch of 10 requests
	requests := make([][]interface{}, 10)
	for i := 0; i < 10; i++ {
		if i%3 == 0 {
			requests[i] = []interface{}{"alice", "data1", "read"}
		} else if i%3 == 1 {
			requests[i] = []interface{}{"alice", "data1", "write"}
		} else {
			requests[i] = []interface{}{"bob", "data2", "read"}
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = e.BatchEnforce(requests)
	}
}

func BenchmarkBatchEnforceLarge(b *testing.B) {
	e, _ := NewEnforcer("examples/rbac_model.conf", "examples/rbac_policy.csv", false)
	
	// Create a batch of 1000 requests
	requests := make([][]interface{}, 1000)
	for i := 0; i < 1000; i++ {
		if i%3 == 0 {
			requests[i] = []interface{}{"alice", "data1", "read"}
		} else if i%3 == 1 {
			requests[i] = []interface{}{"alice", "data1", "write"}
		} else {
			requests[i] = []interface{}{"bob", "data2", "read"}
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = e.BatchEnforce(requests)
	}
}

// Baseline comparison: individual Enforce calls
func BenchmarkEnforceLoop(b *testing.B) {
	e, _ := NewEnforcer("examples/rbac_model.conf", "examples/rbac_policy.csv", false)
	
	requests := make([][]interface{}, 100)
	for i := 0; i < 100; i++ {
		if i%3 == 0 {
			requests[i] = []interface{}{"alice", "data1", "read"}
		} else if i%3 == 1 {
			requests[i] = []interface{}{"alice", "data1", "write"}
		} else {
			requests[i] = []interface{}{"bob", "data2", "read"}
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, req := range requests {
			_, _ = e.Enforce(req...)
		}
	}
}
