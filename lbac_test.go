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

func testEnforceLBAC(t *testing.T, e *Enforcer, sub string, subConf, subInteg float64, obj string, objConf, objInteg float64, act string, res bool) {
	t.Helper()
	if myRes, err := e.Enforce(sub, subConf, subInteg, obj, objConf, objInteg, act); err != nil {
		t.Errorf("Enforce Error: %s", err)
	} else if myRes != res {
		t.Errorf("%s, conf=%v, integ=%v, %s, conf=%v, integ=%v, %s: %t, supposed to be %t", sub, subConf, subInteg, obj, objConf, objInteg, act, myRes, res)
	}
}

func TestLBACModel(t *testing.T) {
	e, _ := NewEnforcer("examples/lbac_model.conf")

	t.Log("Testing normal read operation scenarios")
	testEnforceLBAC(t, e, "admin", 5, 5, "file_topsecret", 3, 3, "read", true) // both high
	testEnforceLBAC(t, e, "manager", 4, 4, "file_secret", 4, 2, "read", true)  // confidentiality equal, integrity higher
	testEnforceLBAC(t, e, "staff", 3, 3, "file_internal", 2, 3, "read", true)  // confidentiality higher, integrity equal
	testEnforceLBAC(t, e, "guest", 2, 2, "file_public", 2, 2, "read", true)    // both dimensions equal

	t.Log("Testing read operation violation scenarios")
	testEnforceLBAC(t, e, "staff", 3, 3, "file_secret", 4, 2, "read", false)      // insufficient confidentiality level
	testEnforceLBAC(t, e, "manager", 4, 4, "file_sensitive", 3, 5, "read", false) // insufficient integrity level
	testEnforceLBAC(t, e, "guest", 2, 2, "file_internal", 3, 1, "read", false)    // insufficient confidentiality level
	testEnforceLBAC(t, e, "staff", 3, 3, "file_protected", 1, 4, "read", false)   // insufficient integrity level

	t.Log("Testing normal write operation scenarios")
	testEnforceLBAC(t, e, "guest", 2, 2, "file_public", 2, 2, "write", true)   // both dimensions equal
	testEnforceLBAC(t, e, "staff", 3, 3, "file_internal", 5, 4, "write", true) // both low
	testEnforceLBAC(t, e, "manager", 4, 4, "file_secret", 4, 5, "write", true) // confidentiality equal, integrity low
	testEnforceLBAC(t, e, "admin", 5, 5, "file_archive", 5, 5, "write", true)  // both dimensions equal

	t.Log("Testing write operation violation scenarios")
	testEnforceLBAC(t, e, "manager", 4, 4, "file_internal", 3, 5, "write", false) // confidentiality level too high
	testEnforceLBAC(t, e, "staff", 3, 3, "file_public", 2, 2, "write", false)     // both dimensions too high
	testEnforceLBAC(t, e, "admin", 5, 5, "file_secret", 5, 4, "write", false)     // integrity level too high
	testEnforceLBAC(t, e, "guest", 2, 2, "file_private", 1, 3, "write", false)    // confidentiality level too high
}
