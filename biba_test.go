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

func testEnforceBiba(t *testing.T, e *Enforcer, sub string, subLevel float64, obj string, objLevel float64, act string, res bool) {
	t.Helper()
	if myRes, err := e.Enforce(sub, subLevel, obj, objLevel, act); err != nil {
		t.Errorf("Enforce Error: %s", err)
	} else if myRes != res {
		t.Errorf("%s, %v, %s, %v, %s: %t, supposed to be %t", sub, subLevel, obj, objLevel, act, myRes, res)
	}
}

func TestBibaModel(t *testing.T) {
	e, _ := NewEnforcer("examples/biba_model.conf")

	testEnforceBiba(t, e, "alice", 3, "data1", 1, "read", false)
	testEnforceBiba(t, e, "bob", 2, "data2", 2, "read", true)
	testEnforceBiba(t, e, "charlie", 1, "data1", 1, "read", true)
	testEnforceBiba(t, e, "bob", 2, "data3", 3, "read", true)
	testEnforceBiba(t, e, "charlie", 1, "data2", 2, "read", true)

	testEnforceBiba(t, e, "alice", 3, "data3", 3, "write", true)
	testEnforceBiba(t, e, "bob", 2, "data3", 3, "write", false)
	testEnforceBiba(t, e, "charlie", 1, "data2", 2, "write", false)
	testEnforceBiba(t, e, "alice", 3, "data1", 1, "write", true)
	testEnforceBiba(t, e, "bob", 2, "data1", 1, "write", true)
}
