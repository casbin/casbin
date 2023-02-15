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
	"sync"
	"sync/atomic"
	"time"

	"github.com/Knetic/govaluate"

	"github.com/casbin/casbin/v2/persist"
	"github.com/casbin/casbin/v2/rbac"
	defaultrolemanager "github.com/casbin/casbin/v2/rbac/default-role-manager"
)

// SyncedEnforcer wraps Enforcer and provides synchronized access
type SyncedEnforcer struct {
	*Enforcer
	m               sync.RWMutex
	stopAutoLoad    chan struct{}
	autoLoadRunning int32
}

// NewSyncedEnforcer creates a synchronized enforcer via file or DB.
func NewSyncedEnforcer(params ...interface{}) (*SyncedEnforcer, error) {
	e := &SyncedEnforcer{}
	var err error
	e.Enforcer, err = NewEnforcer(params...)
	if err != nil {
		return nil, err
	}

	e.stopAutoLoad = make(chan struct{}, 1)
	e.autoLoadRunning = 0
	return e, nil
}

// GetLock return the private RWMutex lock
func (e *SyncedEnforcer) GetLock() *sync.RWMutex {
	return &e.m
}

// IsAutoLoadingRunning check if SyncedEnforcer is auto loading policies
func (e *SyncedEnforcer) IsAutoLoadingRunning() bool {
	return atomic.LoadInt32(&(e.autoLoadRunning)) != 0
}

// StartAutoLoadPolicy starts a go routine that will every specified duration call LoadPolicy
func (e *SyncedEnforcer) StartAutoLoadPolicy(d time.Duration) {
	// Don't start another goroutine if there is already one running
	if !atomic.CompareAndSwapInt32(&e.autoLoadRunning, 0, 1) {
		return
	}

	ticker := time.NewTicker(d)
	go func() {
		defer func() {
			ticker.Stop()
			atomic.StoreInt32(&(e.autoLoadRunning), int32(0))
		}()
		n := 1
		for {
			select {
			case <-ticker.C:
				// error intentionally ignored
				_ = e.LoadPolicy()
				// Uncomment this line to see when the policy is loaded.
				// log.Print("Load policy for time: ", n)
				n++
			case <-e.stopAutoLoad:
				return
			}
		}
	}()
}

// StopAutoLoadPolicy causes the go routine to exit.
func (e *SyncedEnforcer) StopAutoLoadPolicy() {
	if e.IsAutoLoadingRunning() {
		e.stopAutoLoad <- struct{}{}
	}
}

// SetWatcher sets the current watcher.
func (e *SyncedEnforcer) SetWatcher(watcher persist.Watcher) error {
	e.m.Lock()
	defer e.m.Unlock()
	return e.Enforcer.SetWatcher(watcher)
}

// LoadModel reloads the model from the model CONF file.
func (e *SyncedEnforcer) LoadModel() error {
	e.m.Lock()
	defer e.m.Unlock()
	return e.Enforcer.LoadModel()
}

// ClearPolicy clears all policy.
func (e *SyncedEnforcer) ClearPolicy() {
	e.m.Lock()
	defer e.m.Unlock()
	e.Enforcer.ClearPolicy()
}

// LoadPolicy reloads the policy from file/database.
func (e *SyncedEnforcer) LoadPolicy() error {
	e.m.Lock()
	defer e.m.Unlock()
	return e.Enforcer.LoadPolicy()
}

// LoadPolicyFast is not blocked when adapter calls LoadPolicy.
func (e *SyncedEnforcer) LoadPolicyFast() error {
	e.m.RLock()
	newModel := e.model.Copy()
	e.m.RUnlock()

	newModel.ClearPolicy()
	newRmMap := map[string]rbac.RoleManager{}
	var err error

	if err = e.adapter.LoadPolicy(newModel); err != nil && err.Error() != "invalid file path, file path cannot be empty" {
		return err
	}

	if err = newModel.SortPoliciesBySubjectHierarchy(); err != nil {
		return err
	}

	if err = newModel.SortPoliciesByPriority(); err != nil {
		return err
	}

	if e.autoBuildRoleLinks {
		for ptype := range newModel["g"] {
			newRmMap[ptype] = defaultrolemanager.NewRoleManager(10)
		}
		err = newModel.BuildRoleLinks(newRmMap)
		if err != nil {
			return err
		}
	}

	// reduce the lock range
	e.m.Lock()
	defer e.m.Unlock()
	e.model = newModel
	e.rmMap = newRmMap
	return nil
}

// LoadFilteredPolicy reloads a filtered policy from file/database.
func (e *SyncedEnforcer) LoadFilteredPolicy(filter interface{}) error {
	e.m.Lock()
	defer e.m.Unlock()
	return e.Enforcer.LoadFilteredPolicy(filter)
}

// LoadIncrementalFilteredPolicy reloads a filtered policy from file/database.
func (e *SyncedEnforcer) LoadIncrementalFilteredPolicy(filter interface{}) error {
	e.m.Lock()
	defer e.m.Unlock()
	return e.Enforcer.LoadIncrementalFilteredPolicy(filter)
}

// SavePolicy saves the current policy (usually after changed with Casbin API) back to file/database.
func (e *SyncedEnforcer) SavePolicy() error {
	e.m.Lock()
	defer e.m.Unlock()
	return e.Enforcer.SavePolicy()
}

// BuildRoleLinks manually rebuild the role inheritance relations.
func (e *SyncedEnforcer) BuildRoleLinks() error {
	e.m.Lock()
	defer e.m.Unlock()
	return e.Enforcer.BuildRoleLinks()
}

// Enforce decides whether a "subject" can access a "object" with the operation "action", input parameters are usually: (sub, obj, act).
func (e *SyncedEnforcer) Enforce(rvals ...interface{}) (bool, error) {
	e.m.RLock()
	defer e.m.RUnlock()
	return e.Enforcer.Enforce(rvals...)
}

// EnforceWithMatcher use a custom matcher to decides whether a "subject" can access a "object" with the operation "action", input parameters are usually: (matcher, sub, obj, act), use model matcher by default when matcher is "".
func (e *SyncedEnforcer) EnforceWithMatcher(matcher string, rvals ...interface{}) (bool, error) {
	e.m.RLock()
	defer e.m.RUnlock()
	return e.Enforcer.EnforceWithMatcher(matcher, rvals...)
}

// EnforceEx explain enforcement by informing matched rules
func (e *SyncedEnforcer) EnforceEx(rvals ...interface{}) (bool, []string, error) {
	e.m.RLock()
	defer e.m.RUnlock()
	return e.Enforcer.EnforceEx(rvals...)
}

// EnforceExWithMatcher use a custom matcher and explain enforcement by informing matched rules
func (e *SyncedEnforcer) EnforceExWithMatcher(matcher string, rvals ...interface{}) (bool, []string, error) {
	e.m.RLock()
	defer e.m.RUnlock()
	return e.Enforcer.EnforceExWithMatcher(matcher, rvals...)
}

// BatchEnforce enforce in batches
func (e *SyncedEnforcer) BatchEnforce(requests [][]interface{}) ([]bool, error) {
	e.m.RLock()
	defer e.m.RUnlock()
	return e.Enforcer.BatchEnforce(requests)
}

// BatchEnforceWithMatcher enforce with matcher in batches
func (e *SyncedEnforcer) BatchEnforceWithMatcher(matcher string, requests [][]interface{}) ([]bool, error) {
	e.m.RLock()
	defer e.m.RUnlock()
	return e.Enforcer.BatchEnforceWithMatcher(matcher, requests)
}

// GetAllSubjects gets the list of subjects that show up in the current policy.
func (e *SyncedEnforcer) GetAllSubjects() []string {
	e.m.RLock()
	defer e.m.RUnlock()
	return e.Enforcer.GetAllSubjects()
}

// GetAllNamedSubjects gets the list of subjects that show up in the current named policy.
func (e *SyncedEnforcer) GetAllNamedSubjects(ptype string) []string {
	e.m.RLock()
	defer e.m.RUnlock()
	return e.Enforcer.GetAllNamedSubjects(ptype)
}

// GetAllObjects gets the list of objects that show up in the current policy.
func (e *SyncedEnforcer) GetAllObjects() []string {
	e.m.RLock()
	defer e.m.RUnlock()
	return e.Enforcer.GetAllObjects()
}

// GetAllNamedObjects gets the list of objects that show up in the current named policy.
func (e *SyncedEnforcer) GetAllNamedObjects(ptype string) []string {
	e.m.RLock()
	defer e.m.RUnlock()
	return e.Enforcer.GetAllNamedObjects(ptype)
}

// GetAllActions gets the list of actions that show up in the current policy.
func (e *SyncedEnforcer) GetAllActions() []string {
	e.m.RLock()
	defer e.m.RUnlock()
	return e.Enforcer.GetAllActions()
}

// GetAllNamedActions gets the list of actions that show up in the current named policy.
func (e *SyncedEnforcer) GetAllNamedActions(ptype string) []string {
	e.m.RLock()
	defer e.m.RUnlock()
	return e.Enforcer.GetAllNamedActions(ptype)
}

// GetAllRoles gets the list of roles that show up in the current policy.
func (e *SyncedEnforcer) GetAllRoles() []string {
	e.m.RLock()
	defer e.m.RUnlock()
	return e.Enforcer.GetAllRoles()
}

// GetAllNamedRoles gets the list of roles that show up in the current named policy.
func (e *SyncedEnforcer) GetAllNamedRoles(ptype string) []string {
	e.m.RLock()
	defer e.m.RUnlock()
	return e.Enforcer.GetAllNamedRoles(ptype)
}

// GetPolicy gets all the authorization rules in the policy.
func (e *SyncedEnforcer) GetPolicy() [][]string {
	e.m.RLock()
	defer e.m.RUnlock()
	return e.Enforcer.GetPolicy()
}

// GetFilteredPolicy gets all the authorization rules in the policy, field filters can be specified.
func (e *SyncedEnforcer) GetFilteredPolicy(fieldIndex int, fieldValues ...string) [][]string {
	e.m.RLock()
	defer e.m.RUnlock()
	return e.Enforcer.GetFilteredPolicy(fieldIndex, fieldValues...)
}

// GetNamedPolicy gets all the authorization rules in the named policy.
func (e *SyncedEnforcer) GetNamedPolicy(ptype string) [][]string {
	e.m.RLock()
	defer e.m.RUnlock()
	return e.Enforcer.GetNamedPolicy(ptype)
}

// GetFilteredNamedPolicy gets all the authorization rules in the named policy, field filters can be specified.
func (e *SyncedEnforcer) GetFilteredNamedPolicy(ptype string, fieldIndex int, fieldValues ...string) [][]string {
	e.m.RLock()
	defer e.m.RUnlock()
	return e.Enforcer.GetFilteredNamedPolicy(ptype, fieldIndex, fieldValues...)
}

// GetGroupingPolicy gets all the role inheritance rules in the policy.
func (e *SyncedEnforcer) GetGroupingPolicy() [][]string {
	e.m.RLock()
	defer e.m.RUnlock()
	return e.Enforcer.GetGroupingPolicy()
}

// GetFilteredGroupingPolicy gets all the role inheritance rules in the policy, field filters can be specified.
func (e *SyncedEnforcer) GetFilteredGroupingPolicy(fieldIndex int, fieldValues ...string) [][]string {
	e.m.RLock()
	defer e.m.RUnlock()
	return e.Enforcer.GetFilteredGroupingPolicy(fieldIndex, fieldValues...)
}

// GetNamedGroupingPolicy gets all the role inheritance rules in the policy.
func (e *SyncedEnforcer) GetNamedGroupingPolicy(ptype string) [][]string {
	e.m.RLock()
	defer e.m.RUnlock()
	return e.Enforcer.GetNamedGroupingPolicy(ptype)
}

// GetFilteredNamedGroupingPolicy gets all the role inheritance rules in the policy, field filters can be specified.
func (e *SyncedEnforcer) GetFilteredNamedGroupingPolicy(ptype string, fieldIndex int, fieldValues ...string) [][]string {
	e.m.RLock()
	defer e.m.RUnlock()
	return e.Enforcer.GetFilteredNamedGroupingPolicy(ptype, fieldIndex, fieldValues...)
}

// HasPolicy determines whether an authorization rule exists.
func (e *SyncedEnforcer) HasPolicy(params ...interface{}) bool {
	e.m.RLock()
	defer e.m.RUnlock()
	return e.Enforcer.HasPolicy(params...)
}

// HasNamedPolicy determines whether a named authorization rule exists.
func (e *SyncedEnforcer) HasNamedPolicy(ptype string, params ...interface{}) bool {
	e.m.RLock()
	defer e.m.RUnlock()
	return e.Enforcer.HasNamedPolicy(ptype, params...)
}

// AddPolicy adds an authorization rule to the current policy.
// If the rule already exists, the function returns false and the rule will not be added.
// Otherwise the function returns true by adding the new rule.
func (e *SyncedEnforcer) AddPolicy(params ...interface{}) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.Enforcer.AddPolicy(params...)
}

// AddPolicies adds authorization rules to the current policy.
// If the rule already exists, the function returns false for the corresponding rule and the rule will not be added.
// Otherwise the function returns true for the corresponding rule by adding the new rule.
func (e *SyncedEnforcer) AddPolicies(rules [][]string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.Enforcer.AddPolicies(rules)
}

// AddPoliciesEx adds authorization rules to the current policy.
// If the rule already exists, the rule will not be added.
// But unlike AddPolicies, other non-existent rules are added instead of returning false directly
func (e *SyncedEnforcer) AddPoliciesEx(rules [][]string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.Enforcer.AddPoliciesEx(rules)
}

// AddNamedPolicy adds an authorization rule to the current named policy.
// If the rule already exists, the function returns false and the rule will not be added.
// Otherwise the function returns true by adding the new rule.
func (e *SyncedEnforcer) AddNamedPolicy(ptype string, params ...interface{}) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.Enforcer.AddNamedPolicy(ptype, params...)
}

// AddNamedPolicies adds authorization rules to the current named policy.
// If the rule already exists, the function returns false for the corresponding rule and the rule will not be added.
// Otherwise the function returns true for the corresponding by adding the new rule.
func (e *SyncedEnforcer) AddNamedPolicies(ptype string, rules [][]string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.Enforcer.AddNamedPolicies(ptype, rules)
}

// AddNamedPoliciesEx adds authorization rules to the current named policy.
// If the rule already exists, the rule will not be added.
// But unlike AddNamedPolicies, other non-existent rules are added instead of returning false directly
func (e *SyncedEnforcer) AddNamedPoliciesEx(ptype string, rules [][]string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.Enforcer.AddNamedPoliciesEx(ptype, rules)
}

// RemovePolicy removes an authorization rule from the current policy.
func (e *SyncedEnforcer) RemovePolicy(params ...interface{}) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.Enforcer.RemovePolicy(params...)
}

// UpdatePolicy updates an authorization rule from the current policy.
func (e *SyncedEnforcer) UpdatePolicy(oldPolicy []string, newPolicy []string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.Enforcer.UpdatePolicy(oldPolicy, newPolicy)
}

func (e *SyncedEnforcer) UpdateNamedPolicy(ptype string, p1 []string, p2 []string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.Enforcer.UpdateNamedPolicy(ptype, p1, p2)
}

// UpdatePolicies updates authorization rules from the current policies.
func (e *SyncedEnforcer) UpdatePolicies(oldPolices [][]string, newPolicies [][]string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.Enforcer.UpdatePolicies(oldPolices, newPolicies)
}

func (e *SyncedEnforcer) UpdateNamedPolicies(ptype string, p1 [][]string, p2 [][]string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.Enforcer.UpdateNamedPolicies(ptype, p1, p2)
}

func (e *SyncedEnforcer) UpdateFilteredPolicies(newPolicies [][]string, fieldIndex int, fieldValues ...string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.Enforcer.UpdateFilteredPolicies(newPolicies, fieldIndex, fieldValues...)
}

func (e *SyncedEnforcer) UpdateFilteredNamedPolicies(ptype string, newPolicies [][]string, fieldIndex int, fieldValues ...string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.Enforcer.UpdateFilteredNamedPolicies(ptype, newPolicies, fieldIndex, fieldValues...)
}

// RemovePolicies removes authorization rules from the current policy.
func (e *SyncedEnforcer) RemovePolicies(rules [][]string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.Enforcer.RemovePolicies(rules)
}

// RemoveFilteredPolicy removes an authorization rule from the current policy, field filters can be specified.
func (e *SyncedEnforcer) RemoveFilteredPolicy(fieldIndex int, fieldValues ...string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.Enforcer.RemoveFilteredPolicy(fieldIndex, fieldValues...)
}

// RemoveNamedPolicy removes an authorization rule from the current named policy.
func (e *SyncedEnforcer) RemoveNamedPolicy(ptype string, params ...interface{}) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.Enforcer.RemoveNamedPolicy(ptype, params...)
}

// RemoveNamedPolicies removes authorization rules from the current named policy.
func (e *SyncedEnforcer) RemoveNamedPolicies(ptype string, rules [][]string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.Enforcer.RemoveNamedPolicies(ptype, rules)
}

// RemoveFilteredNamedPolicy removes an authorization rule from the current named policy, field filters can be specified.
func (e *SyncedEnforcer) RemoveFilteredNamedPolicy(ptype string, fieldIndex int, fieldValues ...string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.Enforcer.RemoveFilteredNamedPolicy(ptype, fieldIndex, fieldValues...)
}

// HasGroupingPolicy determines whether a role inheritance rule exists.
func (e *SyncedEnforcer) HasGroupingPolicy(params ...interface{}) bool {
	e.m.RLock()
	defer e.m.RUnlock()
	return e.Enforcer.HasGroupingPolicy(params...)
}

// HasNamedGroupingPolicy determines whether a named role inheritance rule exists.
func (e *SyncedEnforcer) HasNamedGroupingPolicy(ptype string, params ...interface{}) bool {
	e.m.RLock()
	defer e.m.RUnlock()
	return e.Enforcer.HasNamedGroupingPolicy(ptype, params...)
}

// AddGroupingPolicy adds a role inheritance rule to the current policy.
// If the rule already exists, the function returns false and the rule will not be added.
// Otherwise the function returns true by adding the new rule.
func (e *SyncedEnforcer) AddGroupingPolicy(params ...interface{}) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.Enforcer.AddGroupingPolicy(params...)
}

// AddGroupingPolicies adds role inheritance rulea to the current policy.
// If the rule already exists, the function returns false for the corresponding policy rule and the rule will not be added.
// Otherwise the function returns true for the corresponding policy rule by adding the new rule.
func (e *SyncedEnforcer) AddGroupingPolicies(rules [][]string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.Enforcer.AddGroupingPolicies(rules)
}

// AddGroupingPoliciesEx adds role inheritance rules to the current policy.
// If the rule already exists, the rule will not be added.
// But unlike AddGroupingPolicies, other non-existent rules are added instead of returning false directly
func (e *SyncedEnforcer) AddGroupingPoliciesEx(rules [][]string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.Enforcer.AddGroupingPoliciesEx(rules)
}

// AddNamedGroupingPolicy adds a named role inheritance rule to the current policy.
// If the rule already exists, the function returns false and the rule will not be added.
// Otherwise the function returns true by adding the new rule.
func (e *SyncedEnforcer) AddNamedGroupingPolicy(ptype string, params ...interface{}) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.Enforcer.AddNamedGroupingPolicy(ptype, params...)
}

// AddNamedGroupingPolicies adds named role inheritance rules to the current policy.
// If the rule already exists, the function returns false for the corresponding policy rule and the rule will not be added.
// Otherwise the function returns true for the corresponding policy rule by adding the new rule.
func (e *SyncedEnforcer) AddNamedGroupingPolicies(ptype string, rules [][]string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.Enforcer.AddNamedGroupingPolicies(ptype, rules)
}

// AddNamedGroupingPoliciesEx adds named role inheritance rules to the current policy.
// If the rule already exists, the rule will not be added.
// But unlike AddNamedGroupingPolicies, other non-existent rules are added instead of returning false directly
func (e *SyncedEnforcer) AddNamedGroupingPoliciesEx(ptype string, rules [][]string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.Enforcer.AddNamedGroupingPoliciesEx(ptype, rules)
}

// RemoveGroupingPolicy removes a role inheritance rule from the current policy.
func (e *SyncedEnforcer) RemoveGroupingPolicy(params ...interface{}) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.Enforcer.RemoveGroupingPolicy(params...)
}

// RemoveGroupingPolicies removes role inheritance rules from the current policy.
func (e *SyncedEnforcer) RemoveGroupingPolicies(rules [][]string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.Enforcer.RemoveGroupingPolicies(rules)
}

// RemoveFilteredGroupingPolicy removes a role inheritance rule from the current policy, field filters can be specified.
func (e *SyncedEnforcer) RemoveFilteredGroupingPolicy(fieldIndex int, fieldValues ...string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.Enforcer.RemoveFilteredGroupingPolicy(fieldIndex, fieldValues...)
}

// RemoveNamedGroupingPolicy removes a role inheritance rule from the current named policy.
func (e *SyncedEnforcer) RemoveNamedGroupingPolicy(ptype string, params ...interface{}) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.Enforcer.RemoveNamedGroupingPolicy(ptype, params...)
}

// RemoveNamedGroupingPolicies removes role inheritance rules from the current named policy.
func (e *SyncedEnforcer) RemoveNamedGroupingPolicies(ptype string, rules [][]string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.Enforcer.RemoveNamedGroupingPolicies(ptype, rules)
}

func (e *SyncedEnforcer) UpdateGroupingPolicy(oldRule []string, newRule []string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.Enforcer.UpdateGroupingPolicy(oldRule, newRule)
}

func (e *SyncedEnforcer) UpdateGroupingPolicies(oldRules [][]string, newRules [][]string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.Enforcer.UpdateGroupingPolicies(oldRules, newRules)
}

func (e *SyncedEnforcer) UpdateNamedGroupingPolicy(ptype string, oldRule []string, newRule []string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.Enforcer.UpdateNamedGroupingPolicy(ptype, oldRule, newRule)
}

func (e *SyncedEnforcer) UpdateNamedGroupingPolicies(ptype string, oldRules [][]string, newRules [][]string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.Enforcer.UpdateNamedGroupingPolicies(ptype, oldRules, newRules)
}

// RemoveFilteredNamedGroupingPolicy removes a role inheritance rule from the current named policy, field filters can be specified.
func (e *SyncedEnforcer) RemoveFilteredNamedGroupingPolicy(ptype string, fieldIndex int, fieldValues ...string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.Enforcer.RemoveFilteredNamedGroupingPolicy(ptype, fieldIndex, fieldValues...)
}

// AddFunction adds a customized function.
func (e *SyncedEnforcer) AddFunction(name string, function govaluate.ExpressionFunction) {
	e.m.Lock()
	defer e.m.Unlock()
	e.Enforcer.AddFunction(name, function)
}

func (e *SyncedEnforcer) SelfAddPolicy(sec string, ptype string, rule []string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.Enforcer.SelfAddPolicy(sec, ptype, rule)
}

func (e *SyncedEnforcer) SelfAddPolicies(sec string, ptype string, rules [][]string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.Enforcer.SelfAddPolicies(sec, ptype, rules)
}

func (e *SyncedEnforcer) SelfAddPoliciesEx(sec string, ptype string, rules [][]string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.Enforcer.SelfAddPoliciesEx(sec, ptype, rules)
}

func (e *SyncedEnforcer) SelfRemovePolicy(sec string, ptype string, rule []string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.Enforcer.SelfRemovePolicy(sec, ptype, rule)
}

func (e *SyncedEnforcer) SelfRemovePolicies(sec string, ptype string, rules [][]string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.Enforcer.SelfRemovePolicies(sec, ptype, rules)
}

func (e *SyncedEnforcer) SelfRemoveFilteredPolicy(sec string, ptype string, fieldIndex int, fieldValues ...string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.Enforcer.SelfRemoveFilteredPolicy(sec, ptype, fieldIndex, fieldValues...)
}

func (e *SyncedEnforcer) SelfUpdatePolicy(sec string, ptype string, oldRule, newRule []string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.Enforcer.SelfUpdatePolicy(sec, ptype, oldRule, newRule)
}

func (e *SyncedEnforcer) SelfUpdatePolicies(sec string, ptype string, oldRules, newRules [][]string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.Enforcer.SelfUpdatePolicies(sec, ptype, oldRules, newRules)
}
