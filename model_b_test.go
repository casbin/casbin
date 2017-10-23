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

func BenchmarkBasicModel(b *testing.B) {
	e := NewEnforcer("examples/basic_model.conf", "examples/basic_policy.csv")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.Enforce("alice", "data1", "read")
	}
}

func BenchmarkRBACModel(b *testing.B) {
	e := NewEnforcer("examples/rbac_model.conf", "examples/rbac_policy.csv")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.Enforce("alice", "data2", "read")
	}
}
