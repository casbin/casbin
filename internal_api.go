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
	"fmt"

	Err "github.com/casbin/casbin/v2/errors"
	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/persist"
)

const (
	notImplemented = "not implemented"
)

func (e *Enforcer) shouldPersist() bool {
	return e.adapter != nil && e.autoSave
}

func (e *Enforcer) shouldNotify() bool {
	return e.watcher != nil && e.autoNotifyWatcher
}

// addPolicy adds a rule to the current policy.
func (e *Enforcer) addPolicyWithoutNotify(sec string, ptype string, rule []string) (bool, error) {
	if e.dispatcher != nil && e.autoNotifyDispatcher {
		return true, e.dispatcher.AddPolicies(sec, ptype, [][]string{rule})
	}

	if e.model.HasPolicy(sec, ptype, rule) {
		return false, nil
	}

	if e.shouldPersist() {
		if err := e.adapter.AddPolicy(sec, ptype, rule); err != nil {
			if err.Error() != notImplemented {
				return false, err
			}
		}
	}

	e.model.AddPolicy(sec, ptype, rule)

	if sec == "g" {
		err := e.BuildIncrementalRoleLinks(model.PolicyAdd, ptype, [][]string{rule})
		if err != nil {
			return true, err
		}
	}

	return true, nil
}

// addPoliciesWithoutNotify adds rules to the current policy without notify
// If autoRemoveRepeat == true, existing rules are automatically filtered
// Otherwise, false is returned directly
func (e *Enforcer) addPoliciesWithoutNotify(sec string, ptype string, rules [][]string, autoRemoveRepeat bool) (bool, error) {
	if e.dispatcher != nil && e.autoNotifyDispatcher {
		return true, e.dispatcher.AddPolicies(sec, ptype, rules)
	}

	if !autoRemoveRepeat && e.model.HasPolicies(sec, ptype, rules) {
		return false, nil
	}

	if e.shouldPersist() {
		if err := e.adapter.(persist.BatchAdapter).AddPolicies(sec, ptype, rules); err != nil {
			if err.Error() != notImplemented {
				return false, err
			}
		}
	}

	e.model.AddPolicies(sec, ptype, rules)

	if sec == "g" {
		err := e.BuildIncrementalRoleLinks(model.PolicyAdd, ptype, rules)
		if err != nil {
			return true, err
		}
	}

	return true, nil
}

// removePolicy removes a rule from the current policy.
func (e *Enforcer) removePolicyWithoutNotify(sec string, ptype string, rule []string) (bool, error) {
	if e.dispatcher != nil && e.autoNotifyDispatcher {
		return true, e.dispatcher.RemovePolicies(sec, ptype, [][]string{rule})
	}

	if e.shouldPersist() {
		if err := e.adapter.RemovePolicy(sec, ptype, rule); err != nil {
			if err.Error() != notImplemented {
				return false, err
			}
		}
	}

	ruleRemoved := e.model.RemovePolicy(sec, ptype, rule)
	if !ruleRemoved {
		return ruleRemoved, nil
	}

	if sec == "g" {
		err := e.BuildIncrementalRoleLinks(model.PolicyRemove, ptype, [][]string{rule})
		if err != nil {
			return ruleRemoved, err
		}
	}

	return ruleRemoved, nil
}

func (e *Enforcer) updatePolicyWithoutNotify(sec string, ptype string, oldRule []string, newRule []string) (bool, error) {
	if e.dispatcher != nil && e.autoNotifyDispatcher {
		return true, e.dispatcher.UpdatePolicy(sec, ptype, oldRule, newRule)
	}

	if e.shouldPersist() {
		if err := e.adapter.(persist.UpdatableAdapter).UpdatePolicy(sec, ptype, oldRule, newRule); err != nil {
			if err.Error() != notImplemented {
				return false, err
			}
		}
	}
	ruleUpdated := e.model.UpdatePolicy(sec, ptype, oldRule, newRule)
	if !ruleUpdated {
		return ruleUpdated, nil
	}

	if sec == "g" {
		err := e.BuildIncrementalRoleLinks(model.PolicyRemove, ptype, [][]string{oldRule}) // remove the old rule
		if err != nil {
			return ruleUpdated, err
		}
		err = e.BuildIncrementalRoleLinks(model.PolicyAdd, ptype, [][]string{newRule}) // add the new rule
		if err != nil {
			return ruleUpdated, err
		}
	}

	return ruleUpdated, nil
}

func (e *Enforcer) updatePoliciesWithoutNotify(sec string, ptype string, oldRules [][]string, newRules [][]string) (bool, error) {
	if len(newRules) != len(oldRules) {
		return false, fmt.Errorf("the length of oldRules should be equal to the length of newRules, but got the length of oldRules is %d, the length of newRules is %d", len(oldRules), len(newRules))
	}

	if e.dispatcher != nil && e.autoNotifyDispatcher {
		return true, e.dispatcher.UpdatePolicies(sec, ptype, oldRules, newRules)
	}

	if e.shouldPersist() {
		if err := e.adapter.(persist.UpdatableAdapter).UpdatePolicies(sec, ptype, oldRules, newRules); err != nil {
			if err.Error() != notImplemented {
				return false, err
			}
		}
	}

	ruleUpdated := e.model.UpdatePolicies(sec, ptype, oldRules, newRules)
	if !ruleUpdated {
		return ruleUpdated, nil
	}

	if sec == "g" {
		err := e.BuildIncrementalRoleLinks(model.PolicyRemove, ptype, oldRules) // remove the old rules
		if err != nil {
			return ruleUpdated, err
		}
		err = e.BuildIncrementalRoleLinks(model.PolicyAdd, ptype, newRules) // add the new rules
		if err != nil {
			return ruleUpdated, err
		}
	}

	return ruleUpdated, nil
}

// removePolicies removes rules from the current policy.
func (e *Enforcer) removePoliciesWithoutNotify(sec string, ptype string, rules [][]string) (bool, error) {
	if !e.model.HasPolicies(sec, ptype, rules) {
		return false, nil
	}

	if e.dispatcher != nil && e.autoNotifyDispatcher {
		return true, e.dispatcher.RemovePolicies(sec, ptype, rules)
	}

	if e.shouldPersist() {
		if err := e.adapter.(persist.BatchAdapter).RemovePolicies(sec, ptype, rules); err != nil {
			if err.Error() != notImplemented {
				return false, err
			}
		}
	}

	rulesRemoved := e.model.RemovePolicies(sec, ptype, rules)
	if !rulesRemoved {
		return rulesRemoved, nil
	}

	if sec == "g" {
		err := e.BuildIncrementalRoleLinks(model.PolicyRemove, ptype, rules)
		if err != nil {
			return rulesRemoved, err
		}
	}
	return rulesRemoved, nil
}

// removeFilteredPolicy removes rules based on field filters from the current policy.
func (e *Enforcer) removeFilteredPolicyWithoutNotify(sec string, ptype string, fieldIndex int, fieldValues []string) (bool, error) {
	if len(fieldValues) == 0 {
		return false, Err.ErrInvalidFieldValuesParameter
	}

	if e.dispatcher != nil && e.autoNotifyDispatcher {
		return true, e.dispatcher.RemoveFilteredPolicy(sec, ptype, fieldIndex, fieldValues...)
	}

	if e.shouldPersist() {
		if err := e.adapter.RemoveFilteredPolicy(sec, ptype, fieldIndex, fieldValues...); err != nil {
			if err.Error() != notImplemented {
				return false, err
			}
		}
	}

	ruleRemoved, effects := e.model.RemoveFilteredPolicy(sec, ptype, fieldIndex, fieldValues...)
	if !ruleRemoved {
		return ruleRemoved, nil
	}

	if sec == "g" {
		err := e.BuildIncrementalRoleLinks(model.PolicyRemove, ptype, effects)
		if err != nil {
			return ruleRemoved, err
		}
	}

	return ruleRemoved, nil
}

func (e *Enforcer) updateFilteredPoliciesWithoutNotify(sec string, ptype string, newRules [][]string, fieldIndex int, fieldValues ...string) ([][]string, error) {
	var (
		oldRules [][]string
		err      error
	)

	if e.shouldPersist() {
		if oldRules, err = e.adapter.(persist.UpdatableAdapter).UpdateFilteredPolicies(sec, ptype, newRules, fieldIndex, fieldValues...); err != nil {
			if err.Error() != notImplemented {
				return nil, err
			}
		}
		// For compatibility, because some adapters return oldRules containing ptype, see https://github.com/casbin/xorm-adapter/issues/49
		for i, oldRule := range oldRules {
			if len(oldRules[i]) == len(e.model[sec][ptype].Tokens)+1 {
				oldRules[i] = oldRule[1:]
			}
		}
	}

	if e.dispatcher != nil && e.autoNotifyDispatcher {
		return oldRules, e.dispatcher.UpdateFilteredPolicies(sec, ptype, oldRules, newRules)
	}

	ruleChanged := e.model.RemovePolicies(sec, ptype, oldRules)
	e.model.AddPolicies(sec, ptype, newRules)
	ruleChanged = ruleChanged && len(newRules) != 0
	if !ruleChanged {
		return make([][]string, 0), nil
	}

	if sec == "g" {
		err := e.BuildIncrementalRoleLinks(model.PolicyRemove, ptype, oldRules) // remove the old rules
		if err != nil {
			return oldRules, err
		}
		err = e.BuildIncrementalRoleLinks(model.PolicyAdd, ptype, newRules) // add the new rules
		if err != nil {
			return oldRules, err
		}
	}

	return oldRules, nil
}

// addPolicy adds a rule to the current policy.
func (e *Enforcer) addPolicy(sec string, ptype string, rule []string) (bool, error) {
	ok, err := e.addPolicyWithoutNotify(sec, ptype, rule)
	if !ok || err != nil {
		return ok, err
	}

	if e.shouldNotify() {
		var err error
		if watcher, ok := e.watcher.(persist.WatcherEx); ok {
			err = watcher.UpdateForAddPolicy(sec, ptype, rule...)
		} else {
			err = e.watcher.Update()
		}
		return true, err
	}

	return true, nil
}

// addPolicies adds rules to the current policy.
// If autoRemoveRepeat == true, existing rules are automatically filtered
// Otherwise, false is returned directly
func (e *Enforcer) addPolicies(sec string, ptype string, rules [][]string, autoRemoveRepeat bool) (bool, error) {
	ok, err := e.addPoliciesWithoutNotify(sec, ptype, rules, autoRemoveRepeat)
	if !ok || err != nil {
		return ok, err
	}

	if e.shouldNotify() {
		var err error
		if watcher, ok := e.watcher.(persist.WatcherEx); ok {
			err = watcher.UpdateForAddPolicies(sec, ptype, rules...)
		} else {
			err = e.watcher.Update()
		}
		return true, err
	}

	return true, nil
}

// removePolicy removes a rule from the current policy.
func (e *Enforcer) removePolicy(sec string, ptype string, rule []string) (bool, error) {
	ok, err := e.removePolicyWithoutNotify(sec, ptype, rule)
	if !ok || err != nil {
		return ok, err
	}

	if e.shouldNotify() {
		var err error
		if watcher, ok := e.watcher.(persist.WatcherEx); ok {
			err = watcher.UpdateForRemovePolicy(sec, ptype, rule...)
		} else {
			err = e.watcher.Update()
		}
		return true, err

	}

	return true, nil
}

func (e *Enforcer) updatePolicy(sec string, ptype string, oldRule []string, newRule []string) (bool, error) {
	ok, err := e.updatePolicyWithoutNotify(sec, ptype, oldRule, newRule)
	if !ok || err != nil {
		return ok, err
	}

	if e.shouldNotify() {
		var err error
		if watcher, ok := e.watcher.(persist.UpdatableWatcher); ok {
			err = watcher.UpdateForUpdatePolicy(sec, ptype, oldRule, newRule)
		} else {
			err = e.watcher.Update()
		}
		return true, err
	}

	return true, nil
}

func (e *Enforcer) updatePolicies(sec string, ptype string, oldRules [][]string, newRules [][]string) (bool, error) {
	ok, err := e.updatePoliciesWithoutNotify(sec, ptype, oldRules, newRules)
	if !ok || err != nil {
		return ok, err
	}

	if e.shouldNotify() {
		var err error
		if watcher, ok := e.watcher.(persist.UpdatableWatcher); ok {
			err = watcher.UpdateForUpdatePolicies(sec, ptype, oldRules, newRules)
		} else {
			err = e.watcher.Update()
		}
		return true, err
	}

	return true, nil
}

// removePolicies removes rules from the current policy.
func (e *Enforcer) removePolicies(sec string, ptype string, rules [][]string) (bool, error) {
	ok, err := e.removePoliciesWithoutNotify(sec, ptype, rules)
	if !ok || err != nil {
		return ok, err
	}

	if e.shouldNotify() {
		var err error
		if watcher, ok := e.watcher.(persist.WatcherEx); ok {
			err = watcher.UpdateForRemovePolicies(sec, ptype, rules...)
		} else {
			err = e.watcher.Update()
		}
		return true, err
	}

	return true, nil
}

// removeFilteredPolicy removes rules based on field filters from the current policy.
func (e *Enforcer) removeFilteredPolicy(sec string, ptype string, fieldIndex int, fieldValues []string) (bool, error) {
	ok, err := e.removeFilteredPolicyWithoutNotify(sec, ptype, fieldIndex, fieldValues)
	if !ok || err != nil {
		return ok, err
	}

	if e.shouldNotify() {
		var err error
		if watcher, ok := e.watcher.(persist.WatcherEx); ok {
			err = watcher.UpdateForRemoveFilteredPolicy(sec, ptype, fieldIndex, fieldValues...)
		} else {
			err = e.watcher.Update()
		}
		return true, err
	}

	return true, nil
}

func (e *Enforcer) updateFilteredPolicies(sec string, ptype string, newRules [][]string, fieldIndex int, fieldValues ...string) (bool, error) {
	oldRules, err := e.updateFilteredPoliciesWithoutNotify(sec, ptype, newRules, fieldIndex, fieldValues...)
	ok := len(oldRules) != 0
	if !ok || err != nil {
		return ok, err
	}

	if e.shouldNotify() {
		var err error
		if watcher, ok := e.watcher.(persist.UpdatableWatcher); ok {
			err = watcher.UpdateForUpdatePolicies(sec, ptype, oldRules, newRules)
		} else {
			err = e.watcher.Update()
		}
		return true, err
	}

	return true, nil
}

func (e *Enforcer) GetFieldIndex(ptype string, field string) (int, error) {
	return e.model.GetFieldIndex(ptype, field)
}

func (e *Enforcer) SetFieldIndex(ptype string, field string, index int) {
	assertion := e.model["p"][ptype]
	assertion.FieldIndexMap[field] = index
}
