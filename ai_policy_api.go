// Copyright 2026 The casbin Authors. All Rights Reserved.
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

// GetAIPolicy gets all the AI policy rules in the policy.
func (e *Enforcer) GetAIPolicy() ([][]string, error) {
	return e.GetNamedAIPolicy("a")
}

// GetFilteredAIPolicy gets all the AI policy rules in the policy, field filters can be specified.
func (e *Enforcer) GetFilteredAIPolicy(fieldIndex int, fieldValues ...string) ([][]string, error) {
	return e.GetFilteredNamedAIPolicy("a", fieldIndex, fieldValues...)
}

// GetNamedAIPolicy gets all the AI policy rules in the named policy.
func (e *Enforcer) GetNamedAIPolicy(ptype string) ([][]string, error) {
	return e.model.GetPolicy("a", ptype)
}

// GetFilteredNamedAIPolicy gets all the AI policy rules in the named policy, field filters can be specified.
func (e *Enforcer) GetFilteredNamedAIPolicy(ptype string, fieldIndex int, fieldValues ...string) ([][]string, error) {
	return e.model.GetFilteredPolicy("a", ptype, fieldIndex, fieldValues...)
}

// HasAIPolicy determines whether an AI policy rule exists.
func (e *Enforcer) HasAIPolicy(params ...string) (bool, error) {
	return e.HasNamedAIPolicy("a", params...)
}

// HasNamedAIPolicy determines whether a named AI policy rule exists.
func (e *Enforcer) HasNamedAIPolicy(ptype string, params ...string) (bool, error) {
	return e.model.HasPolicy("a", ptype, params)
}

// AddAIPolicy adds an AI policy rule to the current policy.
// If the rule already exists, the function returns false and the rule will not be added.
// Otherwise the function returns true by adding the new rule.
func (e *Enforcer) AddAIPolicy(params ...string) (bool, error) {
	return e.AddNamedAIPolicy("a", params...)
}

// AddAIPolicies adds AI policy rules to the current policy.
// If the rule already exists, the function returns false for the corresponding rule and the rule will not be added.
// Otherwise the function returns true for the corresponding rule by adding the new rule.
func (e *Enforcer) AddAIPolicies(rules [][]string) (bool, error) {
	return e.AddNamedAIPolicies("a", rules)
}

// AddNamedAIPolicy adds an AI policy rule to the current named policy.
// If the rule already exists, the function returns false and the rule will not be added.
// Otherwise the function returns true by adding the new rule.
func (e *Enforcer) AddNamedAIPolicy(ptype string, params ...string) (bool, error) {
	return e.addPolicyInternal("a", ptype, params)
}

// AddNamedAIPolicies adds AI policy rules to the current named policy.
// If the rule already exists, the function returns false for the corresponding policy rule and the rule will not be added.
// Otherwise the function returns true for the corresponding policy rule by adding the new rule.
func (e *Enforcer) AddNamedAIPolicies(ptype string, rules [][]string) (bool, error) {
	return e.addPoliciesInternal("a", ptype, rules)
}

// RemoveAIPolicy removes an AI policy rule from the current policy.
func (e *Enforcer) RemoveAIPolicy(params ...string) (bool, error) {
	return e.RemoveNamedAIPolicy("a", params...)
}

// RemoveAIPolicies removes AI policy rules from the current policy.
func (e *Enforcer) RemoveAIPolicies(rules [][]string) (bool, error) {
	return e.RemoveNamedAIPolicies("a", rules)
}

// RemoveFilteredAIPolicy removes an AI policy rule from the current policy, field filters can be specified.
func (e *Enforcer) RemoveFilteredAIPolicy(fieldIndex int, fieldValues ...string) (bool, error) {
	return e.RemoveFilteredNamedAIPolicy("a", fieldIndex, fieldValues...)
}

// RemoveNamedAIPolicy removes an AI policy rule from the current named policy.
func (e *Enforcer) RemoveNamedAIPolicy(ptype string, params ...string) (bool, error) {
	return e.removePolicyInternal("a", ptype, params)
}

// RemoveNamedAIPolicies removes AI policy rules from the current named policy.
func (e *Enforcer) RemoveNamedAIPolicies(ptype string, rules [][]string) (bool, error) {
	return e.removePoliciesInternal("a", ptype, rules)
}

// RemoveFilteredNamedAIPolicy removes an AI policy rule from the current named policy, field filters can be specified.
func (e *Enforcer) RemoveFilteredNamedAIPolicy(ptype string, fieldIndex int, fieldValues ...string) (bool, error) {
	return e.removeFilteredPolicyInternal("a", ptype, fieldIndex, fieldValues...)
}

// UpdateAIPolicy updates an AI policy rule from the current policy.
func (e *Enforcer) UpdateAIPolicy(oldPolicy []string, newPolicy []string) (bool, error) {
	return e.UpdateNamedAIPolicy("a", oldPolicy, newPolicy)
}

// UpdateAIPolicies updates AI policy rules from the current policy.
func (e *Enforcer) UpdateAIPolicies(oldPolicies [][]string, newPolicies [][]string) (bool, error) {
	return e.UpdateNamedAIPolicies("a", oldPolicies, newPolicies)
}

// UpdateNamedAIPolicy updates an AI policy rule from the current named policy.
func (e *Enforcer) UpdateNamedAIPolicy(ptype string, oldPolicy []string, newPolicy []string) (bool, error) {
	return e.updatePolicyInternal("a", ptype, oldPolicy, newPolicy)
}

// UpdateNamedAIPolicies updates AI policy rules from the current named policy.
func (e *Enforcer) UpdateNamedAIPolicies(ptype string, oldPolicies [][]string, newPolicies [][]string) (bool, error) {
	return e.updatePoliciesInternal("a", ptype, oldPolicies, newPolicies)
}
