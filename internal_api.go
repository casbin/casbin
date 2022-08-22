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

// addPolicy adds a rule.
// It returns false if the rule already exists.
func (e *Enforcer) addPolicy(sec string, ptype string, rule []string) (bool, error) {
	if e.model.HasPolicy(sec, ptype, rule) {
		return false, nil
	}

	if e.dispatcher != nil && e.autoNotifyDispatcher {
		return true, e.dispatcher.AddPolicies(sec, ptype, [][]string{rule})
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

	if e.watcher != nil && e.autoNotifyWatcher {
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

// addPolicies adds rules and ignores duplicates.
// If all policies exist, it returns false.
func (e *Enforcer) addPolicies(sec string, ptype string, rules [][]string) (bool, error) {
	if e.model.HasAllPolicies(sec, ptype, rules) {
		return false, nil
	}

	if e.dispatcher != nil && e.autoNotifyDispatcher {
		return true, e.dispatcher.AddPolicies(sec, ptype, rules)
	}

	if e.shouldPersist() {
		if err := e.adapter.(persist.BatchAdapter).AddPolicies(sec, ptype, rules); err != nil {
			if err.Error() != notImplemented {
				return false, err
			}
		}
	}

	affectedPolicies := e.model.AddPoliciesWithAffected(sec, ptype, rules)

	if sec == "g" {
		err := e.BuildIncrementalRoleLinks(model.PolicyAdd, ptype, affectedPolicies)
		if err != nil {
			return true, err
		}
	}

	if e.watcher != nil && e.autoNotifyWatcher {
		var err error
		if watcher, ok := e.watcher.(persist.WatcherEx); ok {
			err = watcher.UpdateForAddPolicies(sec, ptype, affectedPolicies...)
		} else {
			err = e.watcher.Update()
		}
		return true, err
	}

	return true, nil
}

// removePolicy removes the specified rule.
// It returns false if the rule does not exist.
func (e *Enforcer) removePolicy(sec string, ptype string, rule []string) (bool, error) {
	if !e.model.HasPolicy(sec, ptype, rule) {
		return false, nil
	}

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

	_ = e.model.RemovePolicy(sec, ptype, rule)
	// the returned value will always be true because HasPolicy() returns true.

	if sec == "g" {
		err := e.BuildIncrementalRoleLinks(model.PolicyRemove, ptype, [][]string{rule})
		if err != nil {
			return true, err
		}
	}

	if e.watcher != nil && e.autoNotifyWatcher {
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
	if !e.model.HasPolicy(sec, ptype, oldRule) || e.model.HasPolicy(sec, ptype, newRule) {
		return false, nil
	}

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
	_ = e.model.UpdatePolicy(sec, ptype, oldRule, newRule)
	// we already have a check in the beginning and UpdatePolicy() will definitely return true.

	if sec == "g" {
		err := e.BuildIncrementalRoleLinks(model.PolicyRemove, ptype, [][]string{oldRule}) // remove the old rule
		if err != nil {
			return true, err
		}
		err = e.BuildIncrementalRoleLinks(model.PolicyAdd, ptype, [][]string{newRule}) // add the new rule
		if err != nil {
			return true, err
		}
	}

	if e.watcher != nil && e.autoNotifyWatcher {
		var err error
		if watcher, ok := e.watcher.(persist.WatcherUpdatable); ok {
			err = watcher.UpdateForUpdatePolicy(oldRule, newRule)
		} else {
			err = e.watcher.Update()
		}
		return true, err
	}

	return true, nil
}

func (e *Enforcer) updatePolicies(sec string, ptype string, oldRules [][]string, newRules [][]string) (bool, error) {
	if len(newRules) != len(oldRules) {
		return false, fmt.Errorf("the length of oldRules should be equal to the length of newRules, but got the length of oldRules is %d, the length of newRules is %d", len(oldRules), len(newRules))
	}

	if !e.model.HasAnyPolicies(sec, ptype, oldRules) ||
		e.model.HasAllPolicies(sec, ptype, newRules) ||
		model.CheckDuplicatePolicy(newRules) {
		return false, nil
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

	_ = e.model.UpdatePoliciesWithAffected(sec, ptype, oldRules, newRules)
	// the returned value will always be true because we have a check in the beginning.

	if sec == "g" {
		err := e.BuildIncrementalRoleLinks(model.PolicyRemove, ptype, oldRules) // remove the old rules
		if err != nil {
			return true, err
		}
		err = e.BuildIncrementalRoleLinks(model.PolicyAdd, ptype, newRules) // add the new rules
		if err != nil {
			return true, err
		}
	}

	if e.watcher != nil && e.autoNotifyWatcher {
		var err error
		if watcher, ok := e.watcher.(persist.WatcherUpdatable); ok {
			err = watcher.UpdateForUpdatePolicies(oldRules, newRules)
		} else {
			err = e.watcher.Update()
		}
		return true, err
	}

	return true, nil
}

// removePolicies removes the specified rules.
//If all policies do not exist, it returns false.
func (e *Enforcer) removePolicies(sec string, ptype string, rules [][]string) (bool, error) {
	if !e.model.HasAnyPolicies(sec, ptype, rules) {
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

	affectedPolicies := e.model.RemovePoliciesWithAffected(sec, ptype, rules)

	if sec == "g" {
		err := e.BuildIncrementalRoleLinks(model.PolicyRemove, ptype, affectedPolicies)
		if err != nil {
			return true, err
		}
	}

	if e.watcher != nil && e.autoNotifyWatcher {
		var err error
		if watcher, ok := e.watcher.(persist.WatcherEx); ok {
			err = watcher.UpdateForRemovePolicies(sec, ptype, affectedPolicies...)
		} else {
			err = e.watcher.Update()
		}
		return true, err
	}

	return true, nil
}

// removeFilteredPolicy removes rules based on field filters from the current policy.
// If no rules are filtered out, it returns false.
func (e *Enforcer) removeFilteredPolicy(sec string, ptype string, fieldIndex int, fieldValues ...string) (bool, error) {
	if len(fieldValues) == 0 {
		return false, Err.INVALID_FIELDVAULES_PARAMETER
	}

	if !e.model.HasFilteredPolicies(sec, ptype, fieldIndex, fieldValues...) {
		return false, nil
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

	_, affectedPolicies := e.model.RemoveFilteredPolicy(sec, ptype, fieldIndex, fieldValues...)
	// the first returned value is always true because HasFilteredPolicies() returned true

	if sec == "g" {
		err := e.BuildIncrementalRoleLinks(model.PolicyRemove, ptype, affectedPolicies)
		if err != nil {
			return true, err
		}
	}
	if e.watcher != nil && e.autoNotifyWatcher {
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
	var (
		oldRules [][]string
		err      error
	)

	if !e.model.HasFilteredPolicies(sec, ptype, fieldIndex, fieldValues...) ||
		e.model.HasAllPolicies(sec, ptype, newRules) ||
		model.CheckDuplicatePolicy(newRules) {
		return false, nil
	}

	if e.dispatcher != nil && e.autoNotifyDispatcher {
		return true, e.dispatcher.UpdateFilteredPolicies(sec, ptype, oldRules, newRules)
	}

	if e.shouldPersist() {
		if oldRules, err = e.adapter.(persist.UpdatableAdapter).UpdateFilteredPolicies(sec, ptype, newRules, fieldIndex, fieldValues...); err != nil {
			if err.Error() != notImplemented {
				return false, err
			}
		}
		// For compatibility, because some adapters return oldRules containing ptype, see https://github.com/casbin/xorm-adapter/issues/49
		for i, oldRule := range oldRules {
			if len(oldRules[i]) == len(newRules[i])+1 {
				oldRules[i] = oldRule[1:]
			}
		}
	}

	_ = e.model.RemovePolicies(sec, ptype, oldRules)
	// the returned value will always be true because we have a check in the beginning.
	e.model.AddPolicies(sec, ptype, newRules)

	if sec == "g" {
		err := e.BuildIncrementalRoleLinks(model.PolicyRemove, ptype, oldRules) // remove the old rules
		if err != nil {
			return true, err
		}
		err = e.BuildIncrementalRoleLinks(model.PolicyAdd, ptype, newRules) // add the new rules
		if err != nil {
			return true, err
		}
	}

	if e.watcher != nil && e.autoNotifyWatcher {
		var err error
		if watcher, ok := e.watcher.(persist.WatcherUpdatable); ok {
			err = watcher.UpdateForUpdatePolicies(oldRules, newRules)
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
