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
	"fmt"
	"testing"

	"github.com/casbin/casbin/v3/model"
)

// BenchmarkIAMWithoutEffectField benchmarks IAM-like policies without p_eft field
// (Option 1: separate allow/deny policies)
func BenchmarkIAMWithoutEffectField(b *testing.B) {
	m := model.NewModel()
	_ = m.LoadModelFromText(`
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act
`)
	e, _ := NewEnforcer(m)
	
	// Add 1000 allow policies
	for i := 0; i < 1000; i++ {
		_, _ = e.AddPolicy(fmt.Sprintf("role%d", i), fmt.Sprintf("resource%d", i), "read")
	}
	
	// Add groupings
	for i := 0; i < 100; i++ {
		_, _ = e.AddGroupingPolicy(fmt.Sprintf("user%d", i), fmt.Sprintf("role%d", i*10))
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = e.Enforce("user50", "resource500", "read")
	}
}

// BenchmarkIAMWithEffectField benchmarks IAM-like policies with p_eft field
// (Option 2: single policy with allow/deny in the effect field)
func BenchmarkIAMWithEffectField(b *testing.B) {
	m := model.NewModel()
	_ = m.LoadModelFromText(`
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act, eft

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow)) && !some(where (p.eft == deny))

[matchers]
m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act
`)
	e, _ := NewEnforcer(m)
	
	// Add 1000 allow policies
	for i := 0; i < 1000; i++ {
		_, _ = e.AddPolicy(fmt.Sprintf("role%d", i), fmt.Sprintf("resource%d", i), "read", "allow")
	}
	
	// Add groupings
	for i := 0; i < 100; i++ {
		_, _ = e.AddGroupingPolicy(fmt.Sprintf("user%d", i), fmt.Sprintf("role%d", i*10))
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = e.Enforce("user50", "resource500", "read")
	}
}

// BenchmarkIAMWithEffectFieldLarge benchmarks with larger policy set
func BenchmarkIAMWithEffectFieldLarge(b *testing.B) {
	m := model.NewModel()
	_ = m.LoadModelFromText(`
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act, eft

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow)) && !some(where (p.eft == deny))

[matchers]
m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act
`)
	e, _ := NewEnforcer(m)
	
	// Add 5000 allow policies
	for i := 0; i < 5000; i++ {
		_, _ = e.AddPolicy(fmt.Sprintf("role%d", i), fmt.Sprintf("resource%d", i), "read", "allow")
	}
	
	// Add some deny policies
	for i := 0; i < 50; i++ {
		_, _ = e.AddPolicy(fmt.Sprintf("role%d", i), fmt.Sprintf("resource%d", i), "write", "deny")
	}
	
	// Add groupings
	for i := 0; i < 500; i++ {
		_, _ = e.AddGroupingPolicy(fmt.Sprintf("user%d", i), fmt.Sprintf("role%d", i*10))
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = e.Enforce("user250", "resource2500", "read")
	}
}

// BenchmarkIAMWithoutEffectFieldLarge benchmarks with larger policy set without effect field
func BenchmarkIAMWithoutEffectFieldLarge(b *testing.B) {
	m := model.NewModel()
	_ = m.LoadModelFromText(`
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act
`)
	e, _ := NewEnforcer(m)
	
	// Add 5000 allow policies
	for i := 0; i < 5000; i++ {
		_, _ = e.AddPolicy(fmt.Sprintf("role%d", i), fmt.Sprintf("resource%d", i), "read")
	}
	
	// Add groupings
	for i := 0; i < 500; i++ {
		_, _ = e.AddGroupingPolicy(fmt.Sprintf("user%d", i), fmt.Sprintf("role%d", i*10))
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = e.Enforce("user250", "resource2500", "read")
	}
}
