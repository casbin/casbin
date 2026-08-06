// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package util

import "testing"

// The key2 of each benchmark holds a pattern character, so the keyMatchShortcut
// fast path does not apply and the regexp pipeline is what is being measured.

func BenchmarkKeyMatch2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		KeyMatch2("/alice_data/resource1", "/:name/resource1")
	}
}

func BenchmarkKeyMatch3(b *testing.B) {
	for i := 0; i < b.N; i++ {
		KeyMatch3("/proxy/myid/res/1", "/proxy/{id}/res/{rid}")
	}
}

func BenchmarkKeyMatch4(b *testing.B) {
	for i := 0; i < b.N; i++ {
		KeyMatch4("/parent/123/child/123", "/parent/{id}/child/{id}")
	}
}

func BenchmarkKeyMatch5(b *testing.B) {
	for i := 0; i < b.N; i++ {
		KeyMatch5("/parent/child1?status=1", "/parent/{id}")
	}
}

func BenchmarkKeyGet2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		KeyGet2("/alice_data/resource1", "/:name/resource1", "name")
	}
}

func BenchmarkKeyGet3(b *testing.B) {
	for i := 0; i < b.N; i++ {
		KeyGet3("project/proj_project1_admin/", "project/proj_{project}_admin/", "project")
	}
}

func BenchmarkRegexMatch(b *testing.B) {
	for i := 0; i < b.N; i++ {
		RegexMatch("/topic/1", "^/topic/[0-9]+$")
	}
}
