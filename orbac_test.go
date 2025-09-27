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

func testEnforceOrBAC(t *testing.T, e *Enforcer, sub string, org string, role string, activity string, view string, obj string, res bool) {
	t.Helper()
	if myRes, err := e.Enforce(sub, org, role, activity, view, obj); err != nil {
		t.Errorf("Enforce Error: %s", err)
	} else if myRes != res {
		t.Errorf("%s, %s, %s, %s, %s, %s: %t, supposed to be %t", sub, org, role, activity, view, obj, myRes, res)
	}
}

func TestOrBACModel(t *testing.T) {
	e, err := NewEnforcer("examples/orbac_model.conf", "examples/orbac_policy.csv")
	if err != nil {
		t.Fatalf("Error creating enforcer: %v", err)
	}

	t.Log("Testing hospital organization scenarios")
	// Hospital: doctor can read and write patient records
	testEnforceOrBAC(t, e, "alice", "hospital", "doctor", "read", "patient_record", "patient1", true)
	testEnforceOrBAC(t, e, "alice", "hospital", "doctor", "write", "patient_record", "patient1", true)

	// Hospital: nurse can read but not write patient records
	testEnforceOrBAC(t, e, "bob", "hospital", "nurse", "read", "patient_record", "patient1", true)
	testEnforceOrBAC(t, e, "bob", "hospital", "nurse", "write", "patient_record", "patient1", false)

	// Hospital: admin can read and write patient records
	testEnforceOrBAC(t, e, "charlie", "hospital", "admin", "read", "patient_record", "patient1", true)
	testEnforceOrBAC(t, e, "charlie", "hospital", "admin", "write", "patient_record", "patient1", true)

	t.Log("Testing school organization scenarios")
	// School: teacher can read and write student records
	testEnforceOrBAC(t, e, "david", "school", "teacher", "read", "student_record", "student1", true)
	testEnforceOrBAC(t, e, "david", "school", "teacher", "write", "student_record", "student1", true)

	// School: student cannot read or write student records
	testEnforceOrBAC(t, e, "eve", "school", "student", "read", "student_record", "student1", false)
	testEnforceOrBAC(t, e, "eve", "school", "student", "write", "student_record", "student1", false)

	// School: admin can read and write student records
	testEnforceOrBAC(t, e, "frank", "school", "admin", "read", "student_record", "student1", true)
	testEnforceOrBAC(t, e, "frank", "school", "admin", "write", "student_record", "student1", true)

	t.Log("Testing cross-organization access (should be denied)")
	// Cross-organization access should be denied
	testEnforceOrBAC(t, e, "alice", "school", "doctor", "read", "student_record", "student1", false)
	testEnforceOrBAC(t, e, "david", "hospital", "teacher", "read", "patient_record", "patient1", false)
}
