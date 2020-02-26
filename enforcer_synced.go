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
	"log"
	"sync"
	"time"

	"github.com/Knetic/govaluate"

	"github.com/casbin/casbin/v2/effect"
	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/persist"
	"github.com/casbin/casbin/v2/rbac"
)

// SyncedEnforcer wraps Enforcer and provides synchronized access
type SyncedEnforcer struct {
	Enforcer *Enforcer
	base     BasicEnforcer
	api      APIEnforcer
	m        sync.RWMutex
	stopAutoLoad    chan struct{}
	autoLoadRunning bool
	watcher  persist.Watcher
}

// NewSyncedEnforcer creates a synchronized enforcer via file or DB.
func NewSyncedEnforcer(params ...interface{}) (*SyncedEnforcer, error) {
	e := &SyncedEnforcer{}
	if len(params) == 1 {
		if parent, ok := params[0].(FullEnforcer); ok {
			e.base = parent
			e.api = parent
		} else if parent, ok := params[0].(BasicEnforcer); ok {
			e.base = parent
			e.api = &DummyEnforcer{}
		}
	}

	if e.base == nil {
		ef, err := NewEnforcer(params...)
		if err != nil {
			return nil, err
		}
		e.base = ef
		e.api = ef
	}

	e.Enforcer = GetRootEnforcer(e.base)
	e.stopAutoLoad = make(chan struct{}, 1)
	return e, nil
}

// GetParentEnforcer returns the parent enforcer wrapped by this instance.
func (e *SyncedEnforcer) GetParentEnforcer() BasicEnforcer {
	return e.base
}

// InitWithFile initializes an enforcer with a model file and a policy file.
func (e *SyncedEnforcer) InitWithFile(modelPath string, policyPath string) error {
	e.m.Lock()
	defer e.m.Unlock()
	return e.base.InitWithFile(modelPath, policyPath)
}

// InitWithAdapter initializes an enforcer with a database adapter.
func (e *SyncedEnforcer) InitWithAdapter(modelPath string, adapter persist.Adapter) error {
	e.m.Lock()
	defer e.m.Unlock()
	return e.base.InitWithAdapter(modelPath, adapter)
}

// InitWithModelAndAdapter initializes an enforcer with a model and a database adapter.
func (e *SyncedEnforcer) InitWithModelAndAdapter(m model.Model, adapter persist.Adapter) error {
	e.m.Lock()
	defer e.m.Unlock()
	return e.base.InitWithModelAndAdapter(m, adapter)
}

// LoadModel reloads the model from the model CONF file.
// Because the policy is attached to a model, so the policy is invalidated and needs to be reloaded by calling LoadPolicy().
func (e *SyncedEnforcer) LoadModel() error {
	e.m.Lock()
	defer e.m.Unlock()
	return e.base.LoadModel()
}

// GetModel gets the current model.
func (e *SyncedEnforcer) GetModel() model.Model {
	return e.base.GetModel()
}

// SetModel sets the current model.
func (e *SyncedEnforcer) SetModel(m model.Model) {
	e.m.Lock()
	defer e.m.Unlock()
	e.base.SetModel(m)
}

// GetAdapter gets the current adapter.
func (e *SyncedEnforcer) GetAdapter() persist.Adapter {
	return e.base.GetAdapter()
}

// SetAdapter sets the current adapter.
func (e *SyncedEnforcer) SetAdapter(adapter persist.Adapter) {
	e.m.Lock()
	defer e.m.Unlock()
	e.base.SetAdapter(adapter)
}

// SetWatcher sets the current watcher.
func (e *SyncedEnforcer) SetWatcher(watcher persist.Watcher) error {
	e.watcher = watcher
	return watcher.SetUpdateCallback(func(string) { e.LoadPolicy() })
}

// GetRoleManager gets the current role manager.
func (e *SyncedEnforcer) GetRoleManager() rbac.RoleManager {
	return e.base.GetRoleManager()
}

// SetRoleManager sets the current role manager.
func (e *SyncedEnforcer) SetRoleManager(rm rbac.RoleManager) {
	e.m.Lock()
	defer e.m.Unlock()
	e.base.SetRoleManager(rm)
}

// SetEffector sets the current effector.
func (e *SyncedEnforcer) SetEffector(eft effect.Effector) {
	e.m.Lock()
	defer e.m.Unlock()
	e.base.SetEffector(eft)
}

// ClearPolicy clears all policy.
func (e *SyncedEnforcer) ClearPolicy() {
	e.m.Lock()
	defer e.m.Unlock()
	e.base.ClearPolicy()
}

// LoadPolicy reloads the policy from file/database.
func (e *SyncedEnforcer) LoadPolicy() error {
	e.m.Lock()
	defer e.m.Unlock()
	return e.base.LoadPolicy()
}

// LoadFilteredPolicy reloads a filtered policy from file/database.
func (e *SyncedEnforcer) LoadFilteredPolicy(filter interface{}) error {
	e.m.Lock()
	defer e.m.Unlock()
	return e.base.LoadFilteredPolicy(filter)
}

// IsFiltered returns true if the loaded policy has been filtered.
func (e *SyncedEnforcer) IsFiltered() bool {
	return e.base.IsFiltered()
}

// SavePolicy saves the current policy (usually after changed with Casbin API) back to file/database.
func (e *SyncedEnforcer) SavePolicy() error {
	e.m.RLock()
	defer e.m.RUnlock()
	if err := e.base.SavePolicy(); err != nil {
		return err
	}
	if e.watcher != nil {
		return e.watcher.Update()
	}
	return nil
}

// EnableEnforce changes the enforcing state of Casbin, when Casbin is disabled, all access will be allowed by the Enforce() function.
func (e *SyncedEnforcer) EnableEnforce(enable bool) {
	e.base.EnableEnforce(enable)
}

// EnableLog changes whether Casbin will log messages to the Logger.
func (e *SyncedEnforcer) EnableLog(enable bool) {
	e.base.EnableLog(enable)
}

// EnableAutoSave controls whether to save a policy rule automatically to the adapter when it is added or removed.
func (e *SyncedEnforcer) EnableAutoSave(autoSave bool) {
	e.base.EnableAutoSave(autoSave)
}

// EnableAutoBuildRoleLinks controls whether to rebuild the role inheritance relations when a role is added or deleted.
func (e *SyncedEnforcer) EnableAutoBuildRoleLinks(autoBuildRoleLinks bool) {
	e.base.EnableAutoBuildRoleLinks(autoBuildRoleLinks)
}

// BuildRoleLinks manually rebuild the role inheritance relations.
func (e *SyncedEnforcer) BuildRoleLinks() error {
	e.m.RLock()
	defer e.m.RUnlock()
	return e.base.BuildRoleLinks()
}

// Enforce decides whether a "subject" can access a "object" with the operation "action", input parameters are usually: (sub, obj, act).
func (e *SyncedEnforcer) Enforce(rvals ...interface{}) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.base.Enforce(rvals...)
}

// EnforceWithMatcher use a custom matcher to decides whether a "subject" can access a "object" with the operation "action", input parameters are usually: (matcher, sub, obj, act), use model matcher by default when matcher is "".
func (e *SyncedEnforcer) EnforceWithMatcher(matcher string, rvals ...interface{}) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.base.EnforceWithMatcher(matcher, rvals)
}

// StartAutoLoadPolicy starts a go routine that will every specified duration call LoadPolicy
func (e *SyncedEnforcer) StartAutoLoadPolicy(d time.Duration) {
	// Don't start another goroutine if there is already one running
	if e.autoLoadRunning {
		return
	}
	e.autoLoadRunning = true
	ticker := time.NewTicker(d)
	go func() {
		defer func() {
			ticker.Stop()
			e.autoLoadRunning = false
		}()
		n := 1
		log.Print("Start automatically load policy")
		for {
			select {
			case <-ticker.C:
				// error intentionally ignored
				_ = e.LoadPolicy()
				// Uncomment this line to see when the policy is loaded.
				// log.Print("Load policy for time: ", n)
				n++
			case <-e.stopAutoLoad:
				log.Print("Stop automatically load policy")
				return
			}
		}
	}()
}

// StopAutoLoadPolicy causes the go routine to exit.
func (e *SyncedEnforcer) StopAutoLoadPolicy() {
	if e.autoLoadRunning {
		e.stopAutoLoad <- struct{}{}
	}
}

// GetAllSubjects gets the list of subjects that show up in the current policy.
func (e *SyncedEnforcer) GetAllSubjects() []string {
	e.m.RLock()
	defer e.m.RUnlock()
	return e.api.GetAllSubjects()
}

// GetAllNamedSubjects gets the list of subjects that show up in the current named policy.
func (e *SyncedEnforcer) GetAllNamedSubjects(ptype string) []string {
	e.m.RLock()
	defer e.m.RUnlock()
	return e.api.GetAllNamedSubjects(ptype)
}

// GetAllObjects gets the list of objects that show up in the current policy.
func (e *SyncedEnforcer) GetAllObjects() []string {
	e.m.RLock()
	defer e.m.RUnlock()
	return e.api.GetAllObjects()
}

// GetAllNamedObjects gets the list of objects that show up in the current named policy.
func (e *SyncedEnforcer) GetAllNamedObjects(ptype string) []string {
	e.m.RLock()
	defer e.m.RUnlock()
	return e.api.GetAllNamedObjects(ptype)
}

// GetAllActions gets the list of actions that show up in the current policy.
func (e *SyncedEnforcer) GetAllActions() []string {
	e.m.RLock()
	defer e.m.RUnlock()
	return e.api.GetAllActions()
}

// GetAllNamedActions gets the list of actions that show up in the current named policy.
func (e *SyncedEnforcer) GetAllNamedActions(ptype string) []string {
	e.m.RLock()
	defer e.m.RUnlock()
	return e.api.GetAllNamedActions(ptype)
}

// GetAllRoles gets the list of roles that show up in the current policy.
func (e *SyncedEnforcer) GetAllRoles() []string {
	e.m.RLock()
	defer e.m.RUnlock()
	return e.api.GetAllRoles()
}

// GetAllNamedRoles gets the list of roles that show up in the current named policy.
func (e *SyncedEnforcer) GetAllNamedRoles(ptype string) []string {
	e.m.RLock()
	defer e.m.RUnlock()
	return e.api.GetAllNamedRoles(ptype)
}

// GetPolicy gets all the authorization rules in the policy.
func (e *SyncedEnforcer) GetPolicy() [][]string {
	e.m.RLock()
	defer e.m.RUnlock()
	return e.api.GetPolicy()
}

// GetFilteredPolicy gets all the authorization rules in the policy, field filters can be specified.
func (e *SyncedEnforcer) GetFilteredPolicy(fieldIndex int, fieldValues ...string) [][]string {
	e.m.RLock()
	defer e.m.RUnlock()
	return e.api.GetFilteredPolicy(fieldIndex, fieldValues...)
}

// GetNamedPolicy gets all the authorization rules in the named policy.
func (e *SyncedEnforcer) GetNamedPolicy(ptype string) [][]string {
	e.m.RLock()
	defer e.m.RUnlock()
	return e.api.GetNamedPolicy(ptype)
}

// GetFilteredNamedPolicy gets all the authorization rules in the named policy, field filters can be specified.
func (e *SyncedEnforcer) GetFilteredNamedPolicy(ptype string, fieldIndex int, fieldValues ...string) [][]string {
	e.m.RLock()
	defer e.m.RUnlock()
	return e.api.GetFilteredNamedPolicy(ptype, fieldIndex, fieldValues...)
}

// GetGroupingPolicy gets all the role inheritance rules in the policy.
func (e *SyncedEnforcer) GetGroupingPolicy() [][]string {
	e.m.RLock()
	defer e.m.RUnlock()
	return e.api.GetGroupingPolicy()
}

// GetFilteredGroupingPolicy gets all the role inheritance rules in the policy, field filters can be specified.
func (e *SyncedEnforcer) GetFilteredGroupingPolicy(fieldIndex int, fieldValues ...string) [][]string {
	e.m.RLock()
	defer e.m.RUnlock()
	return e.api.GetFilteredGroupingPolicy(fieldIndex, fieldValues...)
}

// GetNamedGroupingPolicy gets all the role inheritance rules in the policy.
func (e *SyncedEnforcer) GetNamedGroupingPolicy(ptype string) [][]string {
	e.m.RLock()
	defer e.m.RUnlock()
	return e.api.GetNamedGroupingPolicy(ptype)
}

// GetFilteredNamedGroupingPolicy gets all the role inheritance rules in the policy, field filters can be specified.
func (e *SyncedEnforcer) GetFilteredNamedGroupingPolicy(ptype string, fieldIndex int, fieldValues ...string) [][]string {
	e.m.RLock()
	defer e.m.RUnlock()
	return e.api.GetFilteredNamedGroupingPolicy(ptype, fieldIndex, fieldValues...)
}

// HasPolicy determines whether an authorization rule exists.
func (e *SyncedEnforcer) HasPolicy(params ...interface{}) bool {
	e.m.RLock()
	defer e.m.RUnlock()
	return e.api.HasPolicy(params...)
}

// HasNamedPolicy determines whether a named authorization rule exists.
func (e *SyncedEnforcer) HasNamedPolicy(ptype string, params ...interface{}) bool {
	e.m.RLock()
	defer e.m.RUnlock()
	return e.api.HasNamedPolicy(ptype, params...)
}

// AddPolicy adds an authorization rule to the current policy.
// If the rule already exists, the function returns false and the rule will not be added.
// Otherwise the function returns true by adding the new rule.
func (e *SyncedEnforcer) AddPolicy(params ...interface{}) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.api.AddPolicy(params...)
}

// AddNamedPolicy adds an authorization rule to the current named policy.
// If the rule already exists, the function returns false and the rule will not be added.
// Otherwise the function returns true by adding the new rule.
func (e *SyncedEnforcer) AddNamedPolicy(ptype string, params ...interface{}) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.api.AddNamedPolicy(ptype, params...)
}

// RemovePolicy removes an authorization rule from the current policy.
func (e *SyncedEnforcer) RemovePolicy(params ...interface{}) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.api.RemovePolicy(params...)
}

// RemoveFilteredPolicy removes an authorization rule from the current policy, field filters can be specified.
func (e *SyncedEnforcer) RemoveFilteredPolicy(fieldIndex int, fieldValues ...string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.api.RemoveFilteredPolicy(fieldIndex, fieldValues...)
}

// RemoveNamedPolicy removes an authorization rule from the current named policy.
func (e *SyncedEnforcer) RemoveNamedPolicy(ptype string, params ...interface{}) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.api.RemoveNamedPolicy(ptype, params...)
}

// RemoveFilteredNamedPolicy removes an authorization rule from the current named policy, field filters can be specified.
func (e *SyncedEnforcer) RemoveFilteredNamedPolicy(ptype string, fieldIndex int, fieldValues ...string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.api.RemoveFilteredNamedPolicy(ptype, fieldIndex, fieldValues...)
}

// HasGroupingPolicy determines whether a role inheritance rule exists.
func (e *SyncedEnforcer) HasGroupingPolicy(params ...interface{}) bool {
	e.m.RLock()
	defer e.m.RUnlock()
	return e.api.HasGroupingPolicy(params...)
}

// HasNamedGroupingPolicy determines whether a named role inheritance rule exists.
func (e *SyncedEnforcer) HasNamedGroupingPolicy(ptype string, params ...interface{}) bool {
	e.m.RLock()
	defer e.m.RUnlock()
	return e.api.HasNamedGroupingPolicy(ptype, params...)
}

// AddGroupingPolicy adds a role inheritance rule to the current policy.
// If the rule already exists, the function returns false and the rule will not be added.
// Otherwise the function returns true by adding the new rule.
func (e *SyncedEnforcer) AddGroupingPolicy(params ...interface{}) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.api.AddGroupingPolicy(params...)
}

// AddNamedGroupingPolicy adds a named role inheritance rule to the current policy.
// If the rule already exists, the function returns false and the rule will not be added.
// Otherwise the function returns true by adding the new rule.
func (e *SyncedEnforcer) AddNamedGroupingPolicy(ptype string, params ...interface{}) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.api.AddNamedGroupingPolicy(ptype, params...)
}

// RemoveGroupingPolicy removes a role inheritance rule from the current policy.
func (e *SyncedEnforcer) RemoveGroupingPolicy(params ...interface{}) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.api.RemoveGroupingPolicy(params...)
}

// RemoveFilteredGroupingPolicy removes a role inheritance rule from the current policy, field filters can be specified.
func (e *SyncedEnforcer) RemoveFilteredGroupingPolicy(fieldIndex int, fieldValues ...string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.api.RemoveFilteredGroupingPolicy(fieldIndex, fieldValues...)
}

// RemoveNamedGroupingPolicy removes a role inheritance rule from the current named policy.
func (e *SyncedEnforcer) RemoveNamedGroupingPolicy(ptype string, params ...interface{}) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.api.RemoveNamedGroupingPolicy(ptype, params...)
}

// RemoveFilteredNamedGroupingPolicy removes a role inheritance rule from the current named policy, field filters can be specified.
func (e *SyncedEnforcer) RemoveFilteredNamedGroupingPolicy(ptype string, fieldIndex int, fieldValues ...string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.api.RemoveFilteredNamedGroupingPolicy(ptype, fieldIndex, fieldValues...)
}

// AddFunction adds a customized function.
func (e *SyncedEnforcer) AddFunction(name string, function govaluate.ExpressionFunction) {
	e.m.Lock()
	defer e.m.Unlock()
	e.api.AddFunction(name, function)
}

// GetImplicitPermissionsForUser gets implicit permissions for a user or role.
// Compared to GetPermissionsForUser(), this function retrieves permissions for inherited roles.
// For example:
// p, admin, data1, read
// p, alice, data2, read
// g, alice, admin
//
// GetPermissionsForUser("alice") can only get: [["alice", "data2", "read"]].
// But GetImplicitPermissionsForUser("alice") will get: [["admin", "data1", "read"], ["alice", "data2", "read"]].
func (e *SyncedEnforcer) GetImplicitPermissionsForUser(user string, domain ...string) ([][]string, error) {
	e.m.RLock()
	defer e.m.RUnlock()
	return e.api.GetImplicitPermissionsForUser(user, domain...)
}

// GetImplicitRolesForUser gets implicit roles that a user has.
// Compared to GetRolesForUser(), this function retrieves indirect roles besides direct roles.
// For example:
// g, alice, role:admin
// g, role:admin, role:user
//
// GetRolesForUser("alice") can only get: ["role:admin"].
// But GetImplicitRolesForUser("alice") will get: ["role:admin", "role:user"].
func (e *SyncedEnforcer) GetImplicitRolesForUser(user string, domain ...string) ([]string, error) {
	e.m.RLock()
	defer e.m.RUnlock()
	return e.api.GetImplicitRolesForUser(user, domain...)
}

// GetImplicitUsersForPermission gets implicit users for a permission.
// For example:
// p, admin, data1, read
// p, bob, data1, read
// g, alice, admin
//
// GetImplicitUsersForPermission("data1", "read") will get: ["alice", "bob"].
// Note: only users will be returned, roles (2nd arg in "g") will be excluded.
func (e *SyncedEnforcer) GetImplicitUsersForPermission(permission ...string) ([]string, error) {
	e.m.RLock()
	defer e.m.RUnlock()
	return e.api.GetImplicitUsersForPermission(permission...)
}
