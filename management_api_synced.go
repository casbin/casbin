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

func (e *SyncedEnforcer) SelfAddPolicy(sec string, ptype string, rule []string) (bool, error) {
	e.m.RLock()
	defer e.m.RUnlock()
	return e.addPolicyWithoutNotify(sec, ptype, rule)
}

func (e *SyncedEnforcer) SelfAddPolicies(sec string, ptype string, rules [][]string) (bool, error) {
	e.m.RLock()
	defer e.m.RUnlock()
	return e.addPoliciesWithoutNotify(sec, ptype, rules)
}

func (e *SyncedEnforcer) SelfRemovePolicy(sec string, ptype string, rule []string) (bool, error) {
	e.m.RLock()
	defer e.m.RUnlock()
	return e.removePolicyWithoutNotify(sec, ptype, rule)
}

func (e *SyncedEnforcer) SelfRemovePolicies(sec string, ptype string, rules [][]string) (bool, error) {
	e.m.RLock()
	defer e.m.RUnlock()
	return e.removePoliciesWithoutNotify(sec, ptype, rules)
}

func (e *SyncedEnforcer) SelfRemoveFilteredPolicy(sec string, ptype string, fieldIndex int, fieldValues ...string) (bool, error) {
	e.m.RLock()
	defer e.m.RUnlock()
	return e.removeFilteredPolicyWithoutNotify(sec, ptype, fieldIndex, fieldValues)
}

func (e *SyncedEnforcer) SelfUpdatePolicy(sec string, ptype string, oldRule, newRule []string) (bool, error) {
	e.m.RLock()
	defer e.m.RUnlock()
	return e.updatePolicyWithoutNotify(sec, ptype, oldRule, newRule)
}

func (e *SyncedEnforcer) SelfUpdatePolicies(sec string, ptype string, oldRules, newRules [][]string) (bool, error) {
	e.m.RLock()
	defer e.m.RUnlock()
	return e.updatePoliciesWithoutNotify(sec, ptype, oldRules, newRules)
}
