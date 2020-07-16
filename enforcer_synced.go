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
	"github.com/casbin/casbin/v3/log"
	"github.com/casbin/casbin/v3/persist"
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

// IsAudoLoadingRunning check if SyncedEnforcer is auto loading policies
func (e *SyncedEnforcer) IsAudoLoadingRunning() bool {
	return atomic.LoadInt32(&(e.autoLoadRunning)) != 0
}

// StartAutoLoadPolicy starts a go routine that will every specified duration call LoadPolicy
func (e *SyncedEnforcer) StartAutoLoadPolicy(d time.Duration) {
	// Don't start another goroutine if there is already one running
	if e.IsAudoLoadingRunning() {
		return
	}
	atomic.StoreInt32(&(e.autoLoadRunning), int32(1))
	ticker := time.NewTicker(d)
	go func() {
		defer func() {
			ticker.Stop()
			atomic.StoreInt32(&(e.autoLoadRunning), int32(0))
		}()
		n := 1
		log.LogPrintf("Start automatically load policy")
		for {
			select {
			case <-ticker.C:
				// error intentionally ignored
				_ = e.LoadPolicy()
				// Uncomment this line to see when the policy is loaded.
				// log.Print("Load policy for time: ", n)
				n++
			case <-e.stopAutoLoad:
				log.LogPrintf("Stop automatically load policy")
				return
			}
		}
	}()
}

// StopAutoLoadPolicy causes the go routine to exit.
func (e *SyncedEnforcer) StopAutoLoadPolicy() {
	if e.IsAudoLoadingRunning() {
		e.stopAutoLoad <- struct{}{}
	}
}

// SetWatcher sets the current watcher.
func (e *SyncedEnforcer) SetWatcher(watcher persist.Watcher) error {
	e.watcher = watcher
	return watcher.SetUpdateCallback(func(string) { _ = e.LoadPolicy() })
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

// LoadFilteredPolicy reloads a filtered policy from file/database.
func (e *SyncedEnforcer) LoadFilteredPolicy(filter interface{}) error {
	e.m.Lock()
	defer e.m.Unlock()
	return e.Enforcer.LoadFilteredPolicy(filter)
}

// SavePolicy saves the current policy (usually after changed with Casbin API) back to file/database.
func (e *SyncedEnforcer) SavePolicy() error {
	e.m.RLock()
	defer e.m.RUnlock()
	return e.Enforcer.SavePolicy()
}

// BuildRoleLinks manually rebuild the role inheritance relations.
func (e *SyncedEnforcer) BuildRoleLinks() error {
	e.m.RLock()
	defer e.m.RUnlock()
	return e.Enforcer.BuildRoleLinks()
}

// Enforce decides whether a "subject" can access a "object" with the operation "action", input parameters are usually: (sub, obj, act).
func (e *SyncedEnforcer) Enforce(rvals ...interface{}) (bool, error) {
	e.m.RLock()
	defer e.m.RUnlock()
	return e.Enforcer.Enforce(rvals...)
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

// AddNamedPolicy adds an authorization rule to the current named policy.
// If the rule already exists, the function returns false and the rule will not be added.
// Otherwise the function returns true by adding the new rule.
func (e *SyncedEnforcer) AddNamedPolicy(ptype string, params ...interface{}) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.Enforcer.AddNamedPolicy(ptype, params...)
}

// RemovePolicy removes an authorization rule from the current policy.
func (e *SyncedEnforcer) RemovePolicy(params ...interface{}) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.Enforcer.RemovePolicy(params...)
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

// AddNamedGroupingPolicy adds a named role inheritance rule to the current policy.
// If the rule already exists, the function returns false and the rule will not be added.
// Otherwise the function returns true by adding the new rule.
func (e *SyncedEnforcer) AddNamedGroupingPolicy(ptype string, params ...interface{}) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.Enforcer.AddNamedGroupingPolicy(ptype, params...)
}

// RemoveGroupingPolicy removes a role inheritance rule from the current policy.
func (e *SyncedEnforcer) RemoveGroupingPolicy(params ...interface{}) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.Enforcer.RemoveGroupingPolicy(params...)
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

// AddGroupingPolicies adds role inheritance rulea to the current policy.
// If the rule already exists, the function returns false for the corresponding policy rule and the rule will not be added.
// Otherwise the function returns true for the corresponding policy rule by adding the new rule.
func (e *SyncedEnforcer) AddGroupingPolicies(rules [][]string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.Enforcer.AddGroupingPolicies(rules)
}

// AddNamedGroupingPolicies adds named role inheritance rules to the current policy.
// If the rule already exists, the function returns false for the corresponding policy rule and the rule will not be added.
// Otherwise the function returns true for the corresponding policy rule by adding the new rule.
func (e *SyncedEnforcer) AddNamedGroupingPolicies(ptype string, rules [][]string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.Enforcer.AddNamedGroupingPolicies(ptype, rules)
}

// AddPolicies adds authorization rules to the current policy.
// If the rule already exists, the function returns false for the corresponding rule and the rule will not be added.
// Otherwise the function returns true for the corresponding rule by adding the new rule.
func (e *SyncedEnforcer) AddPolicies(rules [][]string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.Enforcer.AddPolicies(rules)
}

// AddNamedPolicies adds authorization rules to the current named policy.
// If the rule already exists, the function returns false for the corresponding rule and the rule will not be added.
// Otherwise the function returns true for the corresponding by adding the new rule.
func (e *SyncedEnforcer) AddNamedPolicies(ptype string, rules [][]string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.Enforcer.AddNamedPolicies(ptype, rules)
}
