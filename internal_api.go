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
	Err "github.com/casbin/casbin/v3/errors"
	"github.com/casbin/casbin/v3/log"
	"github.com/casbin/casbin/v3/model"
	"github.com/casbin/casbin/v3/persist"
	"github.com/casbin/casbin/v3/util"
)

const (
	notImplemented = "not implemented"
)

func (e *Enforcer) shouldPersist() bool {
	return e.adapter != nil && e.autoSave
}

// addPolicy adds a rule to the current policy.
func (e *Enforcer) addPolicy(sec string, ptype string, rule []string) (bool, error) {
	if e.dispatcher != nil && e.autoNotifyDispatcher {
		return true, e.dispatcher.AddPolicies(sec, ptype, [][]string{rule})
	}

	if result, err := e.AddPolicyInternal(sec, ptype, rule); err != nil || !result {
		return result, err
	}

	if e.watcher != nil && e.autoNotifyWatcher {
		var err error
		if watcher, ok := e.watcher.(persist.WatcherEx); ok {
			err = watcher.UpdateForAddPolicy(rule...)
		} else {
			err = e.watcher.Update()
		}
		return true, err
	}

	if log.GetLogger().IsEnabled() {
		log.LogPrintf("Policy Management, Type: AddPolicy Assertion: %s::%s\nrule: %s", sec, ptype, util.ArrayToString(rule))
	}

	return true, nil
}

// AddPolicyInternal adds a rule to model and adapter.
func (e *Enforcer) AddPolicyInternal(sec string, ptype string, rule []string) (bool, error) {
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

// addPolicies adds rules to the current policy.
func (e *Enforcer) addPolicies(sec string, ptype string, rules [][]string) (bool, error) {
	if e.dispatcher != nil && e.autoNotifyDispatcher {
		return true, e.dispatcher.AddPolicies(sec, ptype, rules)
	}

	if result, err := e.AddPoliciesInternal(sec, ptype, rules); err != nil || !result {
		return result, err
	}

	if e.watcher != nil && e.autoNotifyWatcher {
		err := e.watcher.Update()
		if err != nil {
			return true, err
		}
	}

	if log.GetLogger().IsEnabled() {
		log.LogPrintf("Policy Management, Type: AddPolicies Assertion: %s::%s\nrules: %s", sec, ptype, util.Array2DToString(rules))
	}

	return true, nil
}

// AddPoliciesInternal adds rules to model and adapter.
func (e *Enforcer) AddPoliciesInternal(sec string, ptype string, rules [][]string) (bool, error) {
	if e.model.HasPolicies(sec, ptype, rules) {
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
func (e *Enforcer) removePolicy(sec string, ptype string, rule []string) (bool, error) {
	if e.dispatcher != nil && e.autoNotifyDispatcher {
		return true, e.dispatcher.RemovePolicies(sec, ptype, [][]string{rule})
	}

	ruleRemoved, err := e.RemovePolicyInternal(sec, ptype, rule)
	if err != nil || !ruleRemoved {
		return ruleRemoved, err
	}

	if e.watcher != nil && e.autoNotifyWatcher {
		var err error
		if watcher, ok := e.watcher.(persist.WatcherEx); ok {
			err = watcher.UpdateForRemovePolicy(rule...)
		} else {
			err = e.watcher.Update()
		}
		return ruleRemoved, err

	}

	if log.GetLogger().IsEnabled() {
		log.LogPrintf("Policy Management, Type: RemovePolicy Assertion %s::%s\nrule: %s", sec, ptype, util.ArrayToString(rule))
	}

	return ruleRemoved, nil
}

// RemovePolicyInternal removes a rule from model and adapter.
func (e *Enforcer) RemovePolicyInternal(sec string, ptype string, rule []string) (bool, error) {
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

// removePolicies removes rules from the current policy.
func (e *Enforcer) removePolicies(sec string, ptype string, rules [][]string) (bool, error) {
	if e.dispatcher != nil && e.autoNotifyDispatcher {
		return true, e.dispatcher.RemovePolicies(sec, ptype, rules)
	}

	rulesRemoved, err := e.RemovePoliciesInternal(sec, ptype, rules)
	if err != nil || !rulesRemoved {
		return rulesRemoved, err
	}

	if e.watcher != nil && e.autoNotifyWatcher {
		err := e.watcher.Update()
		if err != nil {
			return rulesRemoved, err
		}
	}

	if log.GetLogger().IsEnabled() {
		log.LogPrintf("Policy Management, Type: RemovePolicies Assertion %s::%s\nrules: %s", sec, ptype, util.Array2DToString(rules))
	}

	return rulesRemoved, nil
}

// RemovePoliciesInternal removes rules from model and adapter.
func (e *Enforcer) RemovePoliciesInternal(sec string, ptype string, rules [][]string) (bool, error) {
	if !e.model.HasPolicies(sec, ptype, rules) {
		return false, nil
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
func (e *Enforcer) removeFilteredPolicy(sec string, ptype string, fieldIndex int, fieldValues ...string) (bool, error) {
	if e.dispatcher != nil && e.autoNotifyDispatcher {
		return true, e.dispatcher.RemoveFilteredPolicy(sec, ptype, fieldIndex, fieldValues...)
	}

	ruleRemoved, effects, err := e.RemoveFilteredPolicyInternal(sec, ptype, fieldIndex, fieldValues...)
	if err != nil || !ruleRemoved {
		return ruleRemoved, err
	}

	if e.watcher != nil && e.autoNotifyWatcher {
		var err error
		if watcher, ok := e.watcher.(persist.WatcherEx); ok {
			err = watcher.UpdateForRemoveFilteredPolicy(fieldIndex, fieldValues...)
		} else {
			err = e.watcher.Update()
		}
		return ruleRemoved, err
	}

	if log.GetLogger().IsEnabled() {
		log.LogPrintf("Policy Management, Type: RemoveFilteredPolicy Assertion: %s::%s\nrules: %s", sec, ptype, util.Array2DToString(effects))
	}

	return ruleRemoved, nil
}

// RemoveFilteredPolicyInternal removes rules based on field filters from model and adapter.
func (e *Enforcer) RemoveFilteredPolicyInternal(sec string, ptype string, fieldIndex int, fieldValues ...string) (bool, [][]string, error) {
	if len(fieldValues) == 0 {
		return false, nil, Err.INVALID_FIELDVAULES_PARAMETER
	}

	if e.shouldPersist() {
		if err := e.adapter.RemoveFilteredPolicy(sec, ptype, fieldIndex, fieldValues...); err != nil {
			if err.Error() != notImplemented {
				return false, nil, err
			}
		}
	}

	ruleRemoved, effects := e.model.RemoveFilteredPolicy(sec, ptype, fieldIndex, fieldValues...)
	if !ruleRemoved {
		return ruleRemoved, effects, nil
	}

	if sec == "g" {
		err := e.BuildIncrementalRoleLinks(model.PolicyRemove, ptype, effects)
		if err != nil {
			return ruleRemoved, effects, err
		}
	}

	return ruleRemoved, effects, nil
}
