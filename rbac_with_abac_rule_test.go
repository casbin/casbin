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

import "testing"

// makeRBACABACCtx builds a request context with every ctx_rule field populated.
// Go's govaluate errors when a policy expression accesses a missing map key
// (e.g. r.ctx.age on an empty map), so unlike the jcasbin test we always
// supply values for age, type, network and RiskStatus. The default neutral
// values (age >= 18, type/network/RiskStatus not matching any deny rule) act
// as the jcasbin equivalent of an empty context: no deny rule fires.
func makeRBACABACCtx(age int, typ, network, risk string) map[string]interface{} {
	return map[string]interface{}{
		"age":        age,
		"type":       typ,
		"network":    network,
		"RiskStatus": risk,
	}
}

func testRBACWithABACRuleEnforce(t *testing.T, e *Enforcer, sub, obj, act string, ctx map[string]interface{}, res bool) {
	t.Helper()
	if myRes, err := e.Enforce(sub, obj, act, ctx); err != nil {
		t.Errorf("Enforce Error: %s", err)
	} else if myRes != res {
		t.Errorf("%s, %s, %s, %v: %t, supposed to be %t", sub, obj, act, ctx, myRes, res)
	}
}

// TestRBACWithABACRule verifies the combined RBAC + ABAC context-rule model:
//   - r = sub, obj, act, ctx
//   - p = sub, obj, act, ctx_rule, eft
//   - m = g(r.sub, p.sub) && r.obj == p.obj && (r.act == p.act || p.act == "*")
//     && (p.ctx_rule == "noRule" || eval(p.ctx_rule))
//
// The matcher evaluates ctx_rule as a per-request allow/deny filter on the request ctx.
func TestRBACWithABACRule(t *testing.T) {
	e, err := NewEnforcer("examples/rbac_with_abac_rule_model.conf", "examples/rbac_with_abac_rule_policy.csv")
	if err != nil {
		t.Fatalf("NewEnforcer failed: %v", err)
	}

	// neutralCtx is the Go equivalent of jcasbin's empty HashMap: all ctx_rule
	// fields are present, but none of the deny rules match.
	neutralCtx := makeRBACABACCtx(100, "adult", "https", "low")

	// For data1, ctx_rule is `r.ctx.age < 18 || r.ctx.type == 'minor'`.
	// age=18 is on the boundary (18 < 18 is false), and type="minor" triggers deny.
	minorCtx := makeRBACABACCtx(18, "minor", "https", "low")

	// For data2, ctx_rule is `r.ctx.network == 'http'`.
	httpCtx := makeRBACABACCtx(100, "adult", "http", "low")

	// For data4, ctx_rule is `r.ctx.RiskStatus == 'high'`.
	highRiskCtx := makeRBACABACCtx(100, "adult", "https", "high")

	// alice has roles {admin, user}; bob has role {admin}.

	// admin/data1/read: allow under noRule, deny when context matches r.ctx.age < 18 || r.ctx.type == "minor".
	testRBACWithABACRuleEnforce(t, e, "alice", "data1", "read", neutralCtx, true)
	testRBACWithABACRuleEnforce(t, e, "alice", "data1", "read", minorCtx, false)

	// admin/data2: no policy for "read" so it is denied; "write" is allowed under noRule
	// and denied when r.ctx.network == "http".
	testRBACWithABACRuleEnforce(t, e, "alice", "data2", "read", neutralCtx, false)
	testRBACWithABACRuleEnforce(t, e, "alice", "data2", "write", neutralCtx, true)
	testRBACWithABACRuleEnforce(t, e, "alice", "data2", "write", httpCtx, false)

	// admin/data3/* : wildcard action matches any act, allowed under noRule.
	testRBACWithABACRuleEnforce(t, e, "alice", "data3", "read", neutralCtx, true)
	testRBACWithABACRuleEnforce(t, e, "alice", "data3", "write", neutralCtx, true)

	// user/data4/read: allowed under noRule, denied when r.ctx.RiskStatus == "high".
	testRBACWithABACRuleEnforce(t, e, "alice", "data4", "read", neutralCtx, true)
	testRBACWithABACRuleEnforce(t, e, "alice", "data4", "read", highRiskCtx, false)

	// bob is admin only, so he can use admin policies but not user policies.
	testRBACWithABACRuleEnforce(t, e, "bob", "data1", "read", neutralCtx, true)
	testRBACWithABACRuleEnforce(t, e, "bob", "data4", "read", neutralCtx, false)

	// Unknown resource has no matching policy -> denied.
	testRBACWithABACRuleEnforce(t, e, "alice", "data5", "read", neutralCtx, false)
}
