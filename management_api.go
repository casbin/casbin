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

import "github.com/Knetic/govaluate"

// GetAllSubjects gets the list of subjects that show up in the current policy.
func (e *Enforcer) GetAllSubjects() []string {
	return e.model.GetValuesForFieldInPolicyAllTypes("p", 0)
}

// GetAllNamedSubjects gets the list of subjects that show up in the current named policy.
func (e *Enforcer) GetAllNamedSubjects(ptype string) []string {
	return e.model.GetValuesForFieldInPolicy("p", ptype, 0)
}

// GetAllObjects gets the list of objects that show up in the current policy.
func (e *Enforcer) GetAllObjects() []string {
	return e.model.GetValuesForFieldInPolicyAllTypes("p", 1)
}

// GetAllNamedObjects gets the list of objects that show up in the current named policy.
func (e *Enforcer) GetAllNamedObjects(ptype string) []string {
	return e.model.GetValuesForFieldInPolicy("p", ptype, 1)
}

// GetAllActions gets the list of actions that show up in the current policy.
func (e *Enforcer) GetAllActions() []string {
	return e.model.GetValuesForFieldInPolicyAllTypes("p", 2)
}

// GetAllNamedActions gets the list of actions that show up in the current named policy.
func (e *Enforcer) GetAllNamedActions(ptype string) []string {
	return e.model.GetValuesForFieldInPolicy("p", ptype, 2)
}

// GetAllRoles gets the list of roles that show up in the current policy.
func (e *Enforcer) GetAllRoles() []string {
	return e.model.GetValuesForFieldInPolicyAllTypes("g", 1)
}

// GetAllNamedRoles gets the list of roles that show up in the current named policy.
func (e *Enforcer) GetAllNamedRoles(ptype string) []string {
	return e.model.GetValuesForFieldInPolicy("g", ptype, 1)
}

// GetPolicy gets all the authorization rules in the policy.
func (e *Enforcer) GetPolicy() [][]string {
	return e.GetNamedPolicy("p")
}

// GetFilteredPolicy gets all the authorization rules in the policy, field filters can be specified.
func (e *Enforcer) GetFilteredPolicy(fieldIndex int, fieldValues ...string) [][]string {
	return e.GetFilteredNamedPolicy("p", fieldIndex, fieldValues...)
}

// GetNamedPolicy gets all the authorization rules in the named policy.
func (e *Enforcer) GetNamedPolicy(ptype string) [][]string {
	return e.model.GetPolicy("p", ptype)
}

// GetFilteredNamedPolicy gets all the authorization rules in the named policy, field filters can be specified.
func (e *Enforcer) GetFilteredNamedPolicy(ptype string, fieldIndex int, fieldValues ...string) [][]string {
	return e.model.GetFilteredPolicy("p", ptype, fieldIndex, fieldValues...)
}

// GetGroupingPolicy gets all the role inheritance rules in the policy.
func (e *Enforcer) GetGroupingPolicy() [][]string {
	return e.GetNamedGroupingPolicy("g")
}

// GetFilteredGroupingPolicy gets all the role inheritance rules in the policy, field filters can be specified.
func (e *Enforcer) GetFilteredGroupingPolicy(fieldIndex int, fieldValues ...string) [][]string {
	return e.GetFilteredNamedGroupingPolicy("g", fieldIndex, fieldValues...)
}

// GetNamedGroupingPolicy gets all the role inheritance rules in the policy.
func (e *Enforcer) GetNamedGroupingPolicy(ptype string) [][]string {
	return e.model.GetPolicy("g", ptype)
}

// GetFilteredNamedGroupingPolicy gets all the role inheritance rules in the policy, field filters can be specified.
func (e *Enforcer) GetFilteredNamedGroupingPolicy(ptype string, fieldIndex int, fieldValues ...string) [][]string {
	return e.model.GetFilteredPolicy("g", ptype, fieldIndex, fieldValues...)
}

// HasPolicy determines whether an authorization rule exists.
func (e *Enforcer) HasPolicy(params ...interface{}) bool {
	return e.HasNamedPolicy("p", params...)
}

// HasNamedPolicy determines whether a named authorization rule exists.
func (e *Enforcer) HasNamedPolicy(ptype string, params ...interface{}) bool {
	if strSlice, ok := params[0].([]string); len(params) == 1 && ok {
		return e.model.HasPolicy("p", ptype, strSlice)
	}

	policy := make([]string, 0)
	for _, param := range params {
		policy = append(policy, param.(string))
	}

	return e.model.HasPolicy("p", ptype, policy)
}

// AddPolicy adds an authorization rule to the current policy.
// If the rule already exists, the function returns false and the rule will not be added.
// Otherwise the function returns true by adding the new rule.
func (e *Enforcer) AddPolicy(params ...interface{}) (bool, error) {
	return e.AddNamedPolicy("p", params...)
}

// AddPolicies adds authorization rules to the current policy.
// If the rule already exists, the function returns false for the corresponding rule and the rule will not be added.
// Otherwise the function returns true for the corresponding rule by adding the new rule.
func (e *Enforcer) AddPolicies(rules [][]string) (bool, error) {
	return e.AddNamedPolicies("p", rules)
}

// AddNamedPolicy adds an authorization rule to the current named policy.
// If the rule already exists, the function returns false and the rule will not be added.
// Otherwise the function returns true by adding the new rule.
func (e *Enforcer) AddNamedPolicy(ptype string, params ...interface{}) (bool, error) {
	if strSlice, ok := params[0].([]string); len(params) == 1 && ok {
		return e.addPolicy("p", ptype, strSlice)
	}
	policy := make([]string, 0)
	for _, param := range params {
		policy = append(policy, param.(string))
	}

	return e.addPolicy("p", ptype, policy)
}

// AddNamedPolicies adds authorization rules to the current named policy.
// If the rule already exists, the function returns false for the corresponding rule and the rule will not be added.
// Otherwise the function returns true for the corresponding by adding the new rule.
func (e *Enforcer) AddNamedPolicies(ptype string, rules [][]string) (bool, error) {
	return e.addPolicies("p", ptype, rules)
}

// RemovePolicy removes an authorization rule from the current policy.
func (e *Enforcer) RemovePolicy(params ...interface{}) (bool, error) {
	return e.RemoveNamedPolicy("p", params...)
}

// RemovePolicies removes authorization rules from the current policy.
func (e *Enforcer) RemovePolicies(rules [][]string) (bool, error) {
	return e.RemoveNamedPolicies("p", rules)
}

// RemoveFilteredPolicy removes an authorization rule from the current policy, field filters can be specified.
func (e *Enforcer) RemoveFilteredPolicy(fieldIndex int, fieldValues ...string) (bool, error) {
	return e.RemoveFilteredNamedPolicy("p", fieldIndex, fieldValues...)
}

// RemoveNamedPolicy removes an authorization rule from the current named policy.
func (e *Enforcer) RemoveNamedPolicy(ptype string, params ...interface{}) (bool, error) {
	if strSlice, ok := params[0].([]string); len(params) == 1 && ok {
		return e.removePolicy("p", ptype, strSlice)
	}
	policy := make([]string, 0)
	for _, param := range params {
		policy = append(policy, param.(string))
	}

	return e.removePolicy("p", ptype, policy)
}

// RemoveNamedPolicies removes authorization rules from the current named policy.
func (e *Enforcer) RemoveNamedPolicies(ptype string, rules [][]string) (bool, error) {
	return e.removePolicies("p", ptype, rules)
}

// RemoveFilteredNamedPolicy removes an authorization rule from the current named policy, field filters can be specified.
func (e *Enforcer) RemoveFilteredNamedPolicy(ptype string, fieldIndex int, fieldValues ...string) (bool, error) {
	return e.removeFilteredPolicy("p", ptype, fieldIndex, fieldValues...)
}

// HasGroupingPolicy determines whether a role inheritance rule exists.
func (e *Enforcer) HasGroupingPolicy(params ...interface{}) bool {
	return e.HasNamedGroupingPolicy("g", params...)
}

// HasNamedGroupingPolicy determines whether a named role inheritance rule exists.
func (e *Enforcer) HasNamedGroupingPolicy(ptype string, params ...interface{}) bool {
	if strSlice, ok := params[0].([]string); len(params) == 1 && ok {
		return e.model.HasPolicy("g", ptype, strSlice)
	}

	policy := make([]string, 0)
	for _, param := range params {
		policy = append(policy, param.(string))
	}

	return e.model.HasPolicy("g", ptype, policy)
}

// AddGroupingPolicy adds a role inheritance rule to the current policy.
// If the rule already exists, the function returns false and the rule will not be added.
// Otherwise the function returns true by adding the new rule.
func (e *Enforcer) AddGroupingPolicy(params ...interface{}) (bool, error) {
	return e.AddNamedGroupingPolicy("g", params...)
}

// AddGroupingPolicies adds role inheritance rules to the current policy.
// If the rule already exists, the function returns false for the corresponding policy rule and the rule will not be added.
// Otherwise the function returns true for the corresponding policy rule by adding the new rule.
func (e *Enforcer) AddGroupingPolicies(rules [][]string) (bool, error) {
	return e.AddNamedGroupingPolicies("g", rules)
}

// AddNamedGroupingPolicy adds a named role inheritance rule to the current policy.
// If the rule already exists, the function returns false and the rule will not be added.
// Otherwise the function returns true by adding the new rule.
func (e *Enforcer) AddNamedGroupingPolicy(ptype string, params ...interface{}) (bool, error) {
	var ruleAdded bool
	var err error
	if strSlice, ok := params[0].([]string); len(params) == 1 && ok {
		ruleAdded, err = e.addPolicy("g", ptype, strSlice)
	} else {
		policy := make([]string, 0)
		for _, param := range params {
			policy = append(policy, param.(string))
		}

		ruleAdded, err = e.addPolicy("g", ptype, policy)
	}

	return ruleAdded, err
}

// AddNamedGroupingPolicies adds named role inheritance rules to the current policy.
// If the rule already exists, the function returns false for the corresponding policy rule and the rule will not be added.
// Otherwise the function returns true for the corresponding policy rule by adding the new rule.
func (e *Enforcer) AddNamedGroupingPolicies(ptype string, rules [][]string) (bool, error) {
	return e.addPolicies("g", ptype, rules)
}

// RemoveGroupingPolicy removes a role inheritance rule from the current policy.
func (e *Enforcer) RemoveGroupingPolicy(params ...interface{}) (bool, error) {
	return e.RemoveNamedGroupingPolicy("g", params...)
}

// RemoveGroupingPolicies removes role inheritance rules from the current policy.
func (e *Enforcer) RemoveGroupingPolicies(rules [][]string) (bool, error) {
	return e.RemoveNamedGroupingPolicies("g", rules)
}

// RemoveFilteredGroupingPolicy removes a role inheritance rule from the current policy, field filters can be specified.
func (e *Enforcer) RemoveFilteredGroupingPolicy(fieldIndex int, fieldValues ...string) (bool, error) {
	return e.RemoveFilteredNamedGroupingPolicy("g", fieldIndex, fieldValues...)
}

// RemoveNamedGroupingPolicy removes a role inheritance rule from the current named policy.
func (e *Enforcer) RemoveNamedGroupingPolicy(ptype string, params ...interface{}) (bool, error) {
	var ruleRemoved bool
	var err error
	if strSlice, ok := params[0].([]string); len(params) == 1 && ok {
		ruleRemoved, err = e.removePolicy("g", ptype, strSlice)
	} else {
		policy := make([]string, 0)
		for _, param := range params {
			policy = append(policy, param.(string))
		}

		ruleRemoved, err = e.removePolicy("g", ptype, policy)
	}

	return ruleRemoved, err
}

// RemoveNamedGroupingPolicies removes role inheritance rules from the current named policy.
func (e *Enforcer) RemoveNamedGroupingPolicies(ptype string, rules [][]string) (bool, error) {
	return e.removePolicies("g", ptype, rules)
}

// RemoveFilteredNamedGroupingPolicy removes a role inheritance rule from the current named policy, field filters can be specified.
func (e *Enforcer) RemoveFilteredNamedGroupingPolicy(ptype string, fieldIndex int, fieldValues ...string) (bool, error) {
	return e.removeFilteredPolicy("g", ptype, fieldIndex, fieldValues...)
}

// AddFunction adds a customized function.
func (e *Enforcer) AddFunction(name string, function govaluate.ExpressionFunction) {
	e.fm.AddFunction(name, function)
}
