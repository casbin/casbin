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

package util

import (
	"fmt"
	"testing"
)

// BenchmarkKeyMatch2_LiteralDomains mimics DomainManager comparing many
// concrete domains (plus a bare "*" admin wildcard) during BuildRoleLinks.
func BenchmarkKeyMatch2_LiteralDomains(b *testing.B) {
	const n = 1200
	domains := make([]string, n)
	for i := 0; i < n; i++ {
		domains[i] = fmt.Sprintf("team/%d", i)
	}
	domains = append(domains, "*")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, d1 := range domains {
			for _, d2 := range domains {
				_ = KeyMatch2(d1, d2)
			}
		}
	}
}

func BenchmarkKeyMatch2_RESTPatterns(b *testing.B) {
	pairs := [][2]string{
		{"/proxy/myid/res", "/proxy/:id/*"},
		{"/alice/all", "/:id/all"},
		{"/resource1", "/:resource"},
		{"/foo/bar", "/foo/*"},
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, p := range pairs {
			_ = KeyMatch2(p[0], p[1])
		}
	}
}
