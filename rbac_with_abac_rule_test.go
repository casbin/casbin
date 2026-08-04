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

package casbin

import (
	"testing"
)

// testCtx holds the request-time attributes that the cond column of
// examples/rbac_with_abac_rule_policy.csv reads. A struct is the safe shape for
// a context: every field always exists, so a policy can never reference an
// attribute the request forgot to supply. See TestRBACWithABACRuleMissingAttr
// for what happens when it does.
type testCtx struct {
	Hour   int
	Secure bool
	Risk   string
}

// Helper function for RBAC-with-ABAC-condition enforcement testing.
func testEnforceRBACWithABACRule(t *testing.T, e *Enforcer, sub string, obj string, act string, ctx interface{}, res bool) {
	t.Helper()
	myRes, err := e.Enforce(sub, obj, act, ctx)
	if err != nil {
		t.Errorf("Enforce Error: %s", err)
		return
	}
	if myRes != res {
		t.Errorf("%s, %s, %s, %v: %t, supposed to be %t", sub, obj, act, ctx, myRes, res)
	}
}

// TestRBACWithABACRule covers a model where roles decide *what* a subject may
// touch and a per-policy condition decides *under which circumstances*:
//
//	m = g(r.sub, p.sub) && keyMatch(r.obj, p.obj) && keyMatch(r.act, p.act) && eval(p.cond)
//
// The deny-override effect lets a single policy line veto every allow above it.
func TestRBACWithABACRule(t *testing.T) {
	e, err := NewEnforcer("examples/rbac_with_abac_rule_model.conf", "examples/rbac_with_abac_rule_policy.csv")
	if err != nil {
		t.Fatalf("NewEnforcer failed: %v", err)
	}

	// alice is an admin, and admin inherits employee; bob is a plain employee.
	office := testCtx{Hour: 10, Secure: true, Risk: "low"}
	afterHours := testCtx{Hour: 20, Secure: true, Risk: "low"}
	insecure := testCtx{Hour: 10, Secure: false, Risk: "low"}
	risky := testCtx{Hour: 10, Secure: true, Risk: "high"}

	// data1/read is allowed to employees during business hours only.
	testEnforceRBACWithABACRule(t, e, "alice", "data1", "read", office, true)
	testEnforceRBACWithABACRule(t, e, "bob", "data1", "read", office, true)
	testEnforceRBACWithABACRule(t, e, "alice", "data1", "read", afterHours, false)
	testEnforceRBACWithABACRule(t, e, "bob", "data1", "read", afterHours, false)

	// No policy grants data1/write, so the condition never gets a say.
	testEnforceRBACWithABACRule(t, e, "alice", "data1", "write", office, false)

	// data2/write requires a secure connection; data2/read is not granted.
	testEnforceRBACWithABACRule(t, e, "bob", "data2", "write", office, true)
	testEnforceRBACWithABACRule(t, e, "bob", "data2", "write", insecure, false)
	testEnforceRBACWithABACRule(t, e, "bob", "data2", "read", office, false)

	// data3/read is unconditional ("true") but restricted to the admin role.
	testEnforceRBACWithABACRule(t, e, "alice", "data3", "read", office, true)
	testEnforceRBACWithABACRule(t, e, "alice", "data3", "read", afterHours, true)
	testEnforceRBACWithABACRule(t, e, "bob", "data3", "read", office, false)

	// The wildcard deny line outranks every allow, including unconditional ones.
	testEnforceRBACWithABACRule(t, e, "alice", "data1", "read", risky, false)
	testEnforceRBACWithABACRule(t, e, "alice", "data3", "read", risky, false)
	testEnforceRBACWithABACRule(t, e, "bob", "data2", "write", risky, false)

	// carol holds no role, and data4 is covered by no policy.
	testEnforceRBACWithABACRule(t, e, "carol", "data1", "read", office, false)
	testEnforceRBACWithABACRule(t, e, "alice", "data4", "read", office, false)
}

// TestRBACWithABACRuleMapCtx shows that a map works as the context too, as long
// as it carries every attribute the matching policies read.
func TestRBACWithABACRuleMapCtx(t *testing.T) {
	e, err := NewEnforcer("examples/rbac_with_abac_rule_model.conf", "examples/rbac_with_abac_rule_policy.csv")
	if err != nil {
		t.Fatalf("NewEnforcer failed: %v", err)
	}

	office := map[string]interface{}{"Hour": 10, "Secure": true, "Risk": "low"}
	risky := map[string]interface{}{"Hour": 10, "Secure": true, "Risk": "high"}

	testEnforceRBACWithABACRule(t, e, "alice", "data1", "read", office, true)
	testEnforceRBACWithABACRule(t, e, "alice", "data1", "read", risky, false)
}

// TestRBACWithABACRuleMissingAttr pins down the one sharp edge of putting
// expressions in policies: a cond that reads an attribute the request context
// does not carry is an evaluation error, not a denial. Callers must not treat
// the returned false as "denied" without checking err.
func TestRBACWithABACRuleMissingAttr(t *testing.T) {
	e, err := NewEnforcer("examples/rbac_with_abac_rule_model.conf", "examples/rbac_with_abac_rule_policy.csv")
	if err != nil {
		t.Fatalf("NewEnforcer failed: %v", err)
	}

	// The data1/read policy reads r.ctx.Hour, which this map omits.
	if _, err := e.Enforce("bob", "data1", "read", map[string]interface{}{"Secure": true, "Risk": "low"}); err == nil {
		t.Error("Enforce: expected an error when the context omits an attribute used by a policy condition, got nil")
	}

	// Same for a nil context: there is nothing to read attributes from.
	if _, err := e.Enforce("bob", "data1", "read", nil); err == nil {
		t.Error("Enforce: expected an error for a nil context, got nil")
	}
}
