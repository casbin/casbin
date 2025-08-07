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

func testEnforceBLP(t *testing.T, e *Enforcer, sub string, subLevel float64, obj string, objLevel float64, act string, res bool) {
	t.Helper()
	if myRes, err := e.Enforce(sub, subLevel, obj, objLevel, act); err != nil {
		t.Errorf("Enforce Error: %s", err)
	} else if myRes != res {
		t.Errorf("%s, %v, %s, %v, %s: %t, supposed to be %t", sub, subLevel, obj, objLevel, act, myRes, res)
	}
}

func TestBLPModel(t *testing.T) {
	e, _ := NewEnforcer("examples/blp_model.conf")

	testEnforceBLP(t, e, "alice", 3, "top_secret_doc", 3, "read", true)
	testEnforceBLP(t, e, "alice", 3, "secret_doc", 2, "read", true)
	testEnforceBLP(t, e, "bob", 2, "secret_doc", 2, "read", true)
	testEnforceBLP(t, e, "bob", 2, "top_secret_doc", 3, "write", true)
	testEnforceBLP(t, e, "charlie", 1, "public_doc", 1, "read", true)

	testEnforceBLP(t, e, "bob", 2, "top_secret_doc", 3, "read", false)
	testEnforceBLP(t, e, "charlie", 1, "secret_doc", 2, "read", false)
	testEnforceBLP(t, e, "alice", 3, "secret_doc", 2, "write", false)
	testEnforceBLP(t, e, "bob", 2, "public_doc", 1, "write", false)
}
