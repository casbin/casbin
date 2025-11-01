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

func testEnforceOrBAC(t *testing.T, e *Enforcer, sub string, org string, obj string, act string, res bool) {
	t.Helper()
	if myRes, err := e.Enforce(sub, org, obj, act); err != nil {
		t.Errorf("Enforce Error: %s", err)
	} else if myRes != res {
		t.Errorf("%s, %s, %s, %s: %t, supposed to be %t", sub, org, obj, act, myRes, res)
	}
}

func TestOrBACModel(t *testing.T) {
	e, err := NewEnforcer("examples/orbac_model.conf", "examples/orbac_policy.csv")
	if err != nil {
		t.Fatalf("Failed to create enforcer: %v", err)
	}

	// Test alice as manager in org1
	testEnforceOrBAC(t, e, "alice", "org1", "data1", "read", true)
	testEnforceOrBAC(t, e, "alice", "org1", "data1", "write", true)

	// Test bob as employee in org1
	testEnforceOrBAC(t, e, "bob", "org1", "data1", "read", true)
	testEnforceOrBAC(t, e, "bob", "org1", "data1", "write", false)

	// Test charlie as manager in org2
	testEnforceOrBAC(t, e, "charlie", "org2", "data2", "read", true)
	testEnforceOrBAC(t, e, "charlie", "org2", "data2", "write", true)

	// Test david as employee in org2
	testEnforceOrBAC(t, e, "david", "org2", "data2", "read", true)
	testEnforceOrBAC(t, e, "david", "org2", "data2", "write", false)

	// Test cross-organization access (should be denied)
	testEnforceOrBAC(t, e, "alice", "org2", "data2", "read", false)
	testEnforceOrBAC(t, e, "alice", "org2", "data2", "write", false)
	testEnforceOrBAC(t, e, "charlie", "org1", "data1", "read", false)
	testEnforceOrBAC(t, e, "charlie", "org1", "data1", "write", false)

	// Test access to non-existent resources
	testEnforceOrBAC(t, e, "alice", "org1", "data2", "read", false)
	testEnforceOrBAC(t, e, "bob", "org1", "data2", "write", false)
}
