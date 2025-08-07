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
	"testing"
)

func TestBLPModel(t *testing.T) {
	e, err := NewEnforcer("examples/blp_model.conf", "examples/blp_policy.csv")
	if err != nil {
		t.Fatal(err)
	}

	t.Run("BLP Core Rules", func(t *testing.T) {
		result, err := e.Enforce("alice", 3, "top_secret_doc", 3, "read")
		if err != nil {
			t.Fatalf("Enforce failed: %v", err)
		}
		if !result {
			t.Error("Alice should be able to read top_secret_doc")
		}

		result, err = e.Enforce("alice", 3, "secret_doc", 2, "read")
		if err != nil {
			t.Fatalf("Enforce failed: %v", err)
		}
		if !result {
			t.Error("Alice should be able to read secret_doc")
		}

		result, err = e.Enforce("bob", 2, "secret_doc", 2, "read")
		if err != nil {
			t.Fatalf("Enforce failed: %v", err)
		}
		if !result {
			t.Error("Bob should be able to read secret_doc")
		}

		result, err = e.Enforce("bob", 2, "top_secret_doc", 3, "write")
		if err != nil {
			t.Fatalf("Enforce failed: %v", err)
		}
		if !result {
			t.Error("Bob should be able to write to top_secret_doc")
		}

		result, err = e.Enforce("charlie", 1, "public_doc", 1, "read")
		if err != nil {
			t.Fatalf("Enforce failed: %v", err)
		}
		if !result {
			t.Error("Charlie should be able to read public_doc")
		}
	})

	t.Run("BLP Rule Violations", func(t *testing.T) {
		result, err := e.Enforce("bob", 2, "top_secret_doc", 3, "read")
		if err != nil {
			t.Fatalf("Enforce failed: %v", err)
		}
		if result {
			t.Error("Bob should not be able to read top_secret_doc")
		}

		result, err = e.Enforce("charlie", 1, "secret_doc", 2, "read")
		if err != nil {
			t.Fatalf("Enforce failed: %v", err)
		}
		if result {
			t.Error("Charlie should not be able to read secret_doc")
		}

		result, err = e.Enforce("alice", 3, "secret_doc", 2, "write")
		if err != nil {
			t.Fatalf("Enforce failed: %v", err)
		}
		if result {
			t.Error("Alice should not be able to write to secret_doc")
		}

		result, err = e.Enforce("bob", 2, "public_doc", 1, "write")
		if err != nil {
			t.Fatalf("Enforce failed: %v", err)
		}
		if result {
			t.Error("Bob should not be able to write to public_doc")
		}
	})
}
