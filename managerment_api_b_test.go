// Copyright 2020 The casbin Authors. All Rights Reserved.
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
	"math/rand"
	"testing"
)

func BenchmarkHasPolicySmall(b *testing.B) {
	e, _ := NewEnforcer("examples/basic_model.conf", false)

	// 100 roles, 10 resources.
	for i := 0; i < 100; i++ {
		_, _ = e.AddPolicy(fmt.Sprintf("user%d", i), fmt.Sprintf("data%d", i/10), "read")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.HasPolicy(fmt.Sprintf("user%d", rand.Intn(100)), fmt.Sprintf("data%d", rand.Intn(100)/10), "read")
	}
}

func BenchmarkHasPolicyMedium(b *testing.B) {
	e, _ := NewEnforcer("examples/basic_model.conf", false)

	// 1000 roles, 100 resources.
	pPolicies := make([][]string, 0)
	for i := 0; i < 1000; i++ {
		pPolicies = append(pPolicies, []string{fmt.Sprintf("user%d", i), fmt.Sprintf("data%d", i/10), "read"})
	}
	_, err := e.AddPolicies(pPolicies)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.HasPolicy(fmt.Sprintf("user%d", rand.Intn(1000)), fmt.Sprintf("data%d", rand.Intn(1000)/10), "read")
	}
}

func BenchmarkHasPolicyLarge(b *testing.B) {
	e, _ := NewEnforcer("examples/basic_model.conf", false)

	// 10000 roles, 1000 resources.
	pPolicies := make([][]string, 0)
	for i := 0; i < 10000; i++ {
		pPolicies = append(pPolicies, []string{fmt.Sprintf("user%d", i), fmt.Sprintf("data%d", i/10), "read"})
	}
	_, err := e.AddPolicies(pPolicies)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.HasPolicy(fmt.Sprintf("user%d", rand.Intn(10000)), fmt.Sprintf("data%d", rand.Intn(10000)/10), "read")
	}
}

func BenchmarkAddPolicySmall(b *testing.B) {
	e, _ := NewEnforcer("examples/basic_model.conf", false)

	// 100 roles, 10 resources.
	for i := 0; i < 100; i++ {
		_, _ = e.AddPolicy(fmt.Sprintf("user%d", i), fmt.Sprintf("data%d", i/10), "read")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = e.AddPolicy(fmt.Sprintf("user%d", rand.Intn(100)+100), fmt.Sprintf("data%d", (rand.Intn(100)+100)/10), "read")
	}
}

func BenchmarkAddPolicyMedium(b *testing.B) {
	e, _ := NewEnforcer("examples/basic_model.conf", false)

	// 1000 roles, 100 resources.
	pPolicies := make([][]string, 0)
	for i := 0; i < 1000; i++ {
		pPolicies = append(pPolicies, []string{fmt.Sprintf("user%d", i), fmt.Sprintf("data%d", i/10), "read"})
	}
	_, err := e.AddPolicies(pPolicies)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = e.AddPolicy(fmt.Sprintf("user%d", rand.Intn(1000)+1000), fmt.Sprintf("data%d", (rand.Intn(1000)+1000)/10), "read")
	}
}

func BenchmarkAddPolicyLarge(b *testing.B) {
	e, _ := NewEnforcer("examples/basic_model.conf", false)

	// 10000 roles, 1000 resources.
	pPolicies := make([][]string, 0)
	for i := 0; i < 10000; i++ {
		pPolicies = append(pPolicies, []string{fmt.Sprintf("user%d", i), fmt.Sprintf("data%d", i/10), "read"})
	}
	_, err := e.AddPolicies(pPolicies)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = e.AddPolicy(fmt.Sprintf("user%d", rand.Intn(10000)+10000), fmt.Sprintf("data%d", (rand.Intn(10000)+10000)/10), "read")
	}
}

func BenchmarkRemovePolicySmall(b *testing.B) {
	e, _ := NewEnforcer("examples/basic_model.conf", false)

	// 100 roles, 10 resources.
	for i := 0; i < 100; i++ {
		_, _ = e.AddPolicy(fmt.Sprintf("user%d", i), fmt.Sprintf("data%d", i/10), "read")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = e.RemovePolicy(fmt.Sprintf("user%d", rand.Intn(100)), fmt.Sprintf("data%d", rand.Intn(100)/10), "read")
	}
}

func BenchmarkRemovePolicyMedium(b *testing.B) {
	e, _ := NewEnforcer("examples/basic_model.conf", false)

	// 1000 roles, 100 resources.
	pPolicies := make([][]string, 0)
	for i := 0; i < 1000; i++ {
		pPolicies = append(pPolicies, []string{fmt.Sprintf("user%d", i), fmt.Sprintf("data%d", i/10), "read"})
	}
	_, err := e.AddPolicies(pPolicies)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = e.RemovePolicy(fmt.Sprintf("user%d", rand.Intn(1000)), fmt.Sprintf("data%d", rand.Intn(1000)/10), "read")
	}
}

func BenchmarkRemovePolicyLarge(b *testing.B) {
	e, _ := NewEnforcer("examples/basic_model.conf", false)

	// 10000 roles, 1000 resources.
	pPolicies := make([][]string, 0)
	for i := 0; i < 10000; i++ {
		pPolicies = append(pPolicies, []string{fmt.Sprintf("user2%d", i), fmt.Sprintf("data%d", i/10), "read"})
	}
	_, err := e.AddPolicies(pPolicies)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = e.RemovePolicy(fmt.Sprintf("user2%d", rand.Intn(10000)), fmt.Sprintf("data%d", rand.Intn(10000)/10), "read")
	}
}
