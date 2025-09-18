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

// Helper function for PBAC enforcement testing.
func testEnforcePBAC(t *testing.T, e *Enforcer, sub interface{}, obj interface{}, act string, res bool) {
	t.Helper()
	myRes, err := e.Enforce(sub, obj, act)
	if err != nil {
		t.Errorf("Enforce Error: %s", err)
		return
	}
	if myRes != res {
		t.Errorf("%v, %v, %s: %t, supposed to be %t", sub, obj, act, myRes, res)
	}
}

func TestPBACModel(t *testing.T) {
	e, _ := NewEnforcer("examples/pbac_model.conf", "examples/pbac_policy.csv")

	testEnforcePBAC(t, e, map[string]interface{}{"Role": "admin", "Age": 25}, map[string]interface{}{"Type": "doc"}, "read", true)

	testEnforcePBAC(t, e, map[string]interface{}{"Role": "user", "Age": 25}, map[string]interface{}{"Type": "doc"}, "read", false)
	testEnforcePBAC(t, e, map[string]interface{}{"Role": "manager", "Age": 30}, map[string]interface{}{"Type": "doc"}, "read", false)

	testEnforcePBAC(t, e, map[string]interface{}{"Role": "admin", "Age": 25}, map[string]interface{}{"Type": "doc"}, "write", false)
	testEnforcePBAC(t, e, map[string]interface{}{"Role": "admin", "Age": 25}, map[string]interface{}{"Type": "doc"}, "delete", false)

	testEnforcePBAC(t, e, map[string]interface{}{"Role": "admin", "Age": 25}, map[string]interface{}{"Type": "video"}, "read", false)
	testEnforcePBAC(t, e, map[string]interface{}{"Role": "admin", "Age": 25}, map[string]interface{}{"Type": "image"}, "read", false)

	testEnforcePBAC(t, e, map[string]interface{}{"Role": "user", "Age": 18}, map[string]interface{}{"Type": "video"}, "play", true)
	testEnforcePBAC(t, e, map[string]interface{}{"Role": "user", "Age": 25}, map[string]interface{}{"Type": "video"}, "play", true)
	testEnforcePBAC(t, e, map[string]interface{}{"Role": "admin", "Age": 30}, map[string]interface{}{"Type": "video"}, "play", true)

	testEnforcePBAC(t, e, map[string]interface{}{"Role": "user", "Age": 16}, map[string]interface{}{"Type": "video"}, "play", false)
	testEnforcePBAC(t, e, map[string]interface{}{"Role": "user", "Age": 17}, map[string]interface{}{"Type": "video"}, "play", false)

	testEnforcePBAC(t, e, map[string]interface{}{"Role": "user", "Age": 20}, map[string]interface{}{"Type": "video"}, "read", false)
	testEnforcePBAC(t, e, map[string]interface{}{"Role": "user", "Age": 20}, map[string]interface{}{"Type": "video"}, "write", false)

	testEnforcePBAC(t, e, map[string]interface{}{"Role": "user", "Age": 20}, map[string]interface{}{"Type": "doc"}, "play", false)
	testEnforcePBAC(t, e, map[string]interface{}{"Role": "user", "Age": 20}, map[string]interface{}{"Type": "image"}, "play", false)

	testEnforcePBAC(t, e, map[string]interface{}{"Role": "admin", "Age": 20}, map[string]interface{}{"Type": "doc"}, "read", true)
	testEnforcePBAC(t, e, map[string]interface{}{"Role": "admin", "Age": 20}, map[string]interface{}{"Type": "video"}, "play", true)

	testEnforcePBAC(t, e, map[string]interface{}{"Role": "guest", "Age": 15}, map[string]interface{}{"Type": "secret"}, "access", false)
	testEnforcePBAC(t, e, map[string]interface{}{"Role": "visitor", "Age": 25}, map[string]interface{}{"Type": "private"}, "view", false)
}
