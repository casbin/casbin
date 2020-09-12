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
	"github.com/casbin/casbin/v3/log"
	"github.com/casbin/casbin/v3/persist"
	"github.com/casbin/casbin/v3/util"
)

func (e *Enforcer) shouldPersist() bool {
	return e.adapter != nil && e.autoSave
}

// addPolicy adds a rule to the current policy.
func (e *Enforcer) addPolicy(sec string, ptype string, rule []string) (bool, error) {
	if e.dispatcher != nil && e.autoNotifyDispatcher {
		return true, e.dispatcher.AddPolicies(sec, ptype, [][]string{rule})
	}

	if result, err := e.internal.AddPolicy(sec, ptype, rule, e.shouldPersist()); err != nil || !result {
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

// addPolicies adds rules to the current policy.
func (e *Enforcer) addPolicies(sec string, ptype string, rules [][]string) (bool, error) {
	if e.dispatcher != nil && e.autoNotifyDispatcher {
		return true, e.dispatcher.AddPolicies(sec, ptype, rules)
	}

	result, effects, err := e.internal.AddPolicies(sec, ptype, rules, e.shouldPersist())
	if err != nil || !result {
		return result, err
	}

	if e.watcher != nil && e.autoNotifyWatcher {
		err := e.watcher.Update()
		if err != nil {
			return true, err
		}
	}

	if log.GetLogger().IsEnabled() {
		log.LogPrintf("Policy Management, Type: AddPolicies Assertion: %s::%s\nrules: %s", sec, ptype, util.Array2DToString(effects))
	}

	return true, nil
}

// removePolicy removes a rule from the current policy.
func (e *Enforcer) removePolicy(sec string, ptype string, rule []string) (bool, error) {
	if e.dispatcher != nil && e.autoNotifyDispatcher {
		return true, e.dispatcher.RemovePolicies(sec, ptype, [][]string{rule})
	}

	ruleRemoved, err := e.internal.RemovePolicy(sec, ptype, rule, e.shouldPersist())
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

// removePolicies removes rules from the current policy.
func (e *Enforcer) removePolicies(sec string, ptype string, rules [][]string) (bool, error) {
	if e.dispatcher != nil && e.autoNotifyDispatcher {
		return true, e.dispatcher.RemovePolicies(sec, ptype, rules)
	}

	rulesRemoved, effects, err := e.internal.RemovePolicies(sec, ptype, rules, e.shouldPersist())
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
		log.LogPrintf("Policy Management, Type: RemovePolicies Assertion %s::%s\nrules: %s", sec, ptype, util.Array2DToString(effects))
	}

	return rulesRemoved, nil
}

// removeFilteredPolicy removes rules based on field filters from the current policy.
func (e *Enforcer) removeFilteredPolicy(sec string, ptype string, fieldIndex int, fieldValues ...string) (bool, error) {
	if e.dispatcher != nil && e.autoNotifyDispatcher {
		return true, e.dispatcher.RemoveFilteredPolicy(sec, ptype, fieldIndex, fieldValues...)
	}

	ruleRemoved, effects, err := e.internal.RemoveFilteredPolicy(sec, ptype, e.shouldPersist(), fieldIndex, fieldValues...)
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
