// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package casbin

import "context"

var _ IEnforcerContext = &SyncedContextEnforcer{}

// SyncedContextEnforcer wraps ContextEnforcer and provides synchronized access.
type SyncedContextEnforcer struct {
	*SyncedEnforcer
	ctxEnforcer *ContextEnforcer
}

// NewSyncedContextEnforcer creates a synchronized context-aware enforcer via file or DB.
func NewSyncedContextEnforcer(params ...interface{}) (*SyncedContextEnforcer, error) {
	e, err := NewContextEnforcer(params...)
	if err != nil {
		return nil, err
	}

	ctxEnforcer := e.(*ContextEnforcer)
	return &SyncedContextEnforcer{
		SyncedEnforcer: &SyncedEnforcer{
			Enforcer:     ctxEnforcer.Enforcer,
			stopAutoLoad: make(chan struct{}, 1),
		},
		ctxEnforcer: ctxEnforcer,
	}, nil
}

// LoadPolicyCtx loads all policy rules from the storage with context and synchronization.
func (e *SyncedContextEnforcer) LoadPolicyCtx(ctx context.Context) error {
	e.m.RLock()
	newModel, err := e.ctxEnforcer.loadPolicyFromAdapterCtx(ctx, e.model)
	e.m.RUnlock()
	if err != nil {
		return err
	}
	e.m.Lock()
	defer e.m.Unlock()
	return e.applyModifiedModel(newModel)
}

// LoadFilteredPolicyCtx loads all policy rules from the storage with context and filter with synchronization.
func (e *SyncedContextEnforcer) LoadFilteredPolicyCtx(ctx context.Context, filter interface{}) error {
	e.m.Lock()
	defer e.m.Unlock()
	return e.ctxEnforcer.LoadFilteredPolicyCtx(ctx, filter)
}

// LoadIncrementalFilteredPolicyCtx append a filtered policy from file/database with context and synchronization.
func (e *SyncedContextEnforcer) LoadIncrementalFilteredPolicyCtx(ctx context.Context, filter interface{}) error {
	e.m.Lock()
	defer e.m.Unlock()
	return e.ctxEnforcer.LoadIncrementalFilteredPolicyCtx(ctx, filter)
}

// IsFilteredCtx returns true if the loaded policy has been filtered with context and synchronization.
func (e *SyncedContextEnforcer) IsFilteredCtx(ctx context.Context) bool {
	e.m.RLock()
	defer e.m.RUnlock()
	return e.ctxEnforcer.IsFilteredCtx(ctx)
}

// SavePolicyCtx calls ContextEnforcer.SavePolicyCtx with synchronization.
func (e *SyncedContextEnforcer) SavePolicyCtx(ctx context.Context) error {
	e.m.Lock()
	defer e.m.Unlock()
	return e.ctxEnforcer.SavePolicyCtx(ctx)
}

// AddPolicyCtx adds a policy rule to the storage with context and synchronization.
func (e *SyncedContextEnforcer) AddPolicyCtx(ctx context.Context, params ...interface{}) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.ctxEnforcer.AddPolicyCtx(ctx, params...)
}

// AddPoliciesCtx adds policy rules to the storage with context and synchronization.
func (e *SyncedContextEnforcer) AddPoliciesCtx(ctx context.Context, rules [][]string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.ctxEnforcer.AddPoliciesCtx(ctx, rules)
}

// AddNamedPolicyCtx adds a named policy rule to the storage with context and synchronization.
func (e *SyncedContextEnforcer) AddNamedPolicyCtx(ctx context.Context, ptype string, params ...interface{}) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.ctxEnforcer.AddNamedPolicyCtx(ctx, ptype, params...)
}

// AddNamedPoliciesCtx adds named policy rules to the storage with context and synchronization.
func (e *SyncedContextEnforcer) AddNamedPoliciesCtx(ctx context.Context, ptype string, rules [][]string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.ctxEnforcer.AddNamedPoliciesCtx(ctx, ptype, rules)
}

// AddPoliciesExCtx calls ContextEnforcer.AddPoliciesExCtx with synchronization.
func (e *SyncedContextEnforcer) AddPoliciesExCtx(ctx context.Context, rules [][]string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.ctxEnforcer.AddPoliciesExCtx(ctx, rules)
}

// AddNamedPoliciesExCtx calls ContextEnforcer.AddNamedPoliciesExCtx with synchronization.
func (e *SyncedContextEnforcer) AddNamedPoliciesExCtx(ctx context.Context, ptype string, rules [][]string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.ctxEnforcer.AddNamedPoliciesExCtx(ctx, ptype, rules)
}

// RemovePolicyCtx removes a policy rule from the storage with context and synchronization.
func (e *SyncedContextEnforcer) RemovePolicyCtx(ctx context.Context, params ...interface{}) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.ctxEnforcer.RemovePolicyCtx(ctx, params...)
}

// RemoveNamedPolicyCtx removes a named policy rule from the storage with context and synchronization.
func (e *SyncedContextEnforcer) RemoveNamedPolicyCtx(ctx context.Context, ptype string, params ...interface{}) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.ctxEnforcer.RemoveNamedPolicyCtx(ctx, ptype, params...)
}

// RemovePoliciesCtx removes policy rules from the storage with context and synchronization.
func (e *SyncedContextEnforcer) RemovePoliciesCtx(ctx context.Context, rules [][]string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.ctxEnforcer.RemovePoliciesCtx(ctx, rules)
}

// RemoveNamedPoliciesCtx removes named policy rules from the storage with context and synchronization.
func (e *SyncedContextEnforcer) RemoveNamedPoliciesCtx(ctx context.Context, ptype string, rules [][]string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.ctxEnforcer.RemoveNamedPoliciesCtx(ctx, ptype, rules)
}

// RemoveFilteredPolicyCtx removes policy rules that match the filter from the storage with context and synchronization.
func (e *SyncedContextEnforcer) RemoveFilteredPolicyCtx(ctx context.Context, fieldIndex int, fieldValues ...string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.ctxEnforcer.RemoveFilteredPolicyCtx(ctx, fieldIndex, fieldValues...)
}

// RemoveFilteredNamedPolicyCtx removes named policy rules that match the filter from the storage with context and synchronization.
func (e *SyncedContextEnforcer) RemoveFilteredNamedPolicyCtx(ctx context.Context, ptype string, fieldIndex int, fieldValues ...string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.ctxEnforcer.RemoveFilteredNamedPolicyCtx(ctx, ptype, fieldIndex, fieldValues...)
}

// UpdatePolicyCtx updates a policy rule in the storage with context and synchronization.
func (e *SyncedContextEnforcer) UpdatePolicyCtx(ctx context.Context, oldPolicy []string, newPolicy []string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.ctxEnforcer.UpdatePolicyCtx(ctx, oldPolicy, newPolicy)
}

// UpdateNamedPolicyCtx updates a named policy rule in the storage with context and synchronization.
func (e *SyncedContextEnforcer) UpdateNamedPolicyCtx(ctx context.Context, ptype string, p1 []string, p2 []string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.ctxEnforcer.UpdateNamedPolicyCtx(ctx, ptype, p1, p2)
}

// UpdatePoliciesCtx updates policy rules in the storage with context and synchronization.
func (e *SyncedContextEnforcer) UpdatePoliciesCtx(ctx context.Context, oldPolicies [][]string, newPolicies [][]string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.ctxEnforcer.UpdatePoliciesCtx(ctx, oldPolicies, newPolicies)
}

// UpdateNamedPoliciesCtx updates named policy rules in the storage with context and synchronization.
func (e *SyncedContextEnforcer) UpdateNamedPoliciesCtx(ctx context.Context, ptype string, p1 [][]string, p2 [][]string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.ctxEnforcer.UpdateNamedPoliciesCtx(ctx, ptype, p1, p2)
}

// UpdateFilteredPoliciesCtx updates policy rules that match the filter in the storage with context and synchronization.
func (e *SyncedContextEnforcer) UpdateFilteredPoliciesCtx(ctx context.Context, newPolicies [][]string, fieldIndex int, fieldValues ...string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.ctxEnforcer.UpdateFilteredPoliciesCtx(ctx, newPolicies, fieldIndex, fieldValues...)
}

// UpdateFilteredNamedPoliciesCtx updates named policy rules that match the filter in the storage with context and synchronization.
func (e *SyncedContextEnforcer) UpdateFilteredNamedPoliciesCtx(ctx context.Context, ptype string, newPolicies [][]string, fieldIndex int, fieldValues ...string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.ctxEnforcer.UpdateFilteredNamedPoliciesCtx(ctx, ptype, newPolicies, fieldIndex, fieldValues...)
}

// AddGroupingPolicyCtx adds a grouping policy rule to the storage with context and synchronization.
func (e *SyncedContextEnforcer) AddGroupingPolicyCtx(ctx context.Context, params ...interface{}) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.ctxEnforcer.AddGroupingPolicyCtx(ctx, params...)
}

// AddGroupingPoliciesCtx adds grouping policy rules to the storage with context and synchronization.
func (e *SyncedContextEnforcer) AddGroupingPoliciesCtx(ctx context.Context, rules [][]string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.ctxEnforcer.AddGroupingPoliciesCtx(ctx, rules)
}

// AddGroupingPoliciesExCtx calls ContextEnforcer.AddGroupingPoliciesExCtx with synchronization.
func (e *SyncedContextEnforcer) AddGroupingPoliciesExCtx(ctx context.Context, rules [][]string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.ctxEnforcer.AddGroupingPoliciesExCtx(ctx, rules)
}

// AddNamedGroupingPolicyCtx adds a named grouping policy rule to the storage with context and synchronization.
func (e *SyncedContextEnforcer) AddNamedGroupingPolicyCtx(ctx context.Context, ptype string, params ...interface{}) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.ctxEnforcer.AddNamedGroupingPolicyCtx(ctx, ptype, params...)
}

// AddNamedGroupingPoliciesCtx adds named grouping policy rules to the storage with context and synchronization.
func (e *SyncedContextEnforcer) AddNamedGroupingPoliciesCtx(ctx context.Context, ptype string, rules [][]string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.ctxEnforcer.AddNamedGroupingPoliciesCtx(ctx, ptype, rules)
}

// AddNamedGroupingPoliciesExCtx calls ContextEnforcer.AddNamedGroupingPoliciesExCtx with synchronization.
func (e *SyncedContextEnforcer) AddNamedGroupingPoliciesExCtx(ctx context.Context, ptype string, rules [][]string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.ctxEnforcer.AddNamedGroupingPoliciesExCtx(ctx, ptype, rules)
}

// RemoveGroupingPolicyCtx removes a grouping policy rule from the storage with context and synchronization.
func (e *SyncedContextEnforcer) RemoveGroupingPolicyCtx(ctx context.Context, params ...interface{}) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.ctxEnforcer.RemoveGroupingPolicyCtx(ctx, params...)
}

// RemoveNamedGroupingPolicyCtx removes a named grouping policy rule from the storage with context and synchronization.
func (e *SyncedContextEnforcer) RemoveNamedGroupingPolicyCtx(ctx context.Context, ptype string, params ...interface{}) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.ctxEnforcer.RemoveNamedGroupingPolicyCtx(ctx, ptype, params...)
}

// RemoveGroupingPoliciesCtx removes grouping policy rules from the storage with context and synchronization.
func (e *SyncedContextEnforcer) RemoveGroupingPoliciesCtx(ctx context.Context, rules [][]string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.ctxEnforcer.RemoveGroupingPoliciesCtx(ctx, rules)
}

// RemoveNamedGroupingPoliciesCtx removes named grouping policy rules from the storage with context and synchronization.
func (e *SyncedContextEnforcer) RemoveNamedGroupingPoliciesCtx(ctx context.Context, ptype string, rules [][]string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.ctxEnforcer.RemoveNamedGroupingPoliciesCtx(ctx, ptype, rules)
}

// RemoveFilteredGroupingPolicyCtx removes grouping policy rules that match the filter from the storage with context and synchronization.
func (e *SyncedContextEnforcer) RemoveFilteredGroupingPolicyCtx(ctx context.Context, fieldIndex int, fieldValues ...string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.ctxEnforcer.RemoveFilteredGroupingPolicyCtx(ctx, fieldIndex, fieldValues...)
}

// RemoveFilteredNamedGroupingPolicyCtx removes named grouping policy rules that match the filter from the storage with context and synchronization.
func (e *SyncedContextEnforcer) RemoveFilteredNamedGroupingPolicyCtx(ctx context.Context, ptype string, fieldIndex int, fieldValues ...string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.ctxEnforcer.RemoveFilteredNamedGroupingPolicyCtx(ctx, ptype, fieldIndex, fieldValues...)
}

// UpdateGroupingPolicyCtx updates a grouping policy rule in the storage with context and synchronization.
func (e *SyncedContextEnforcer) UpdateGroupingPolicyCtx(ctx context.Context, oldRule []string, newRule []string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.ctxEnforcer.UpdateGroupingPolicyCtx(ctx, oldRule, newRule)
}

// UpdateNamedGroupingPolicyCtx updates a named grouping policy rule in the storage with context and synchronization.
func (e *SyncedContextEnforcer) UpdateNamedGroupingPolicyCtx(ctx context.Context, ptype string, oldRule []string, newRule []string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.ctxEnforcer.UpdateNamedGroupingPolicyCtx(ctx, ptype, oldRule, newRule)
}

// UpdateGroupingPoliciesCtx updates grouping policy rules in the storage with context and synchronization.
func (e *SyncedContextEnforcer) UpdateGroupingPoliciesCtx(ctx context.Context, oldRules [][]string, newRules [][]string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.ctxEnforcer.UpdateGroupingPoliciesCtx(ctx, oldRules, newRules)
}

// UpdateNamedGroupingPoliciesCtx updates named grouping policy rules in the storage with context and synchronization.
func (e *SyncedContextEnforcer) UpdateNamedGroupingPoliciesCtx(ctx context.Context, ptype string, oldRules [][]string, newRules [][]string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.ctxEnforcer.UpdateNamedGroupingPoliciesCtx(ctx, ptype, oldRules, newRules)
}

// SelfAddPolicyCtx adds a policy rule to the current policy with context and synchronization.
func (e *SyncedContextEnforcer) SelfAddPolicyCtx(ctx context.Context, sec string, ptype string, rule []string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.ctxEnforcer.SelfAddPolicyCtx(ctx, sec, ptype, rule)
}

// SelfAddPoliciesCtx adds policy rules to the current policy with context and synchronization.
func (e *SyncedContextEnforcer) SelfAddPoliciesCtx(ctx context.Context, sec string, ptype string, rules [][]string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.ctxEnforcer.SelfAddPoliciesCtx(ctx, sec, ptype, rules)
}

// SelfAddPoliciesExCtx calls ContextEnforcer.SelfAddPoliciesExCtx with synchronization.
func (e *SyncedContextEnforcer) SelfAddPoliciesExCtx(ctx context.Context, sec string, ptype string, rules [][]string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.ctxEnforcer.SelfAddPoliciesExCtx(ctx, sec, ptype, rules)
}

// SelfRemovePolicyCtx removes a policy rule from the current policy with context and synchronization.
func (e *SyncedContextEnforcer) SelfRemovePolicyCtx(ctx context.Context, sec string, ptype string, rule []string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.ctxEnforcer.SelfRemovePolicyCtx(ctx, sec, ptype, rule)
}

// SelfRemovePoliciesCtx removes policy rules from the current policy with context and synchronization.
func (e *SyncedContextEnforcer) SelfRemovePoliciesCtx(ctx context.Context, sec string, ptype string, rules [][]string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.ctxEnforcer.SelfRemovePoliciesCtx(ctx, sec, ptype, rules)
}

// SelfRemoveFilteredPolicyCtx removes policy rules that match the filter from the current policy with context and synchronization.
func (e *SyncedContextEnforcer) SelfRemoveFilteredPolicyCtx(ctx context.Context, sec string, ptype string, fieldIndex int, fieldValues ...string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.ctxEnforcer.SelfRemoveFilteredPolicyCtx(ctx, sec, ptype, fieldIndex, fieldValues...)
}

// SelfUpdatePolicyCtx updates a policy rule in the current policy with context and synchronization.
func (e *SyncedContextEnforcer) SelfUpdatePolicyCtx(ctx context.Context, sec string, ptype string, oldRule, newRule []string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.ctxEnforcer.SelfUpdatePolicyCtx(ctx, sec, ptype, oldRule, newRule)
}

// SelfUpdatePoliciesCtx updates policy rules in the current policy with context and synchronization.
func (e *SyncedContextEnforcer) SelfUpdatePoliciesCtx(ctx context.Context, sec string, ptype string, oldRules, newRules [][]string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.ctxEnforcer.SelfUpdatePoliciesCtx(ctx, sec, ptype, oldRules, newRules)
}

// AddRoleForUserCtx adds a role for a user with context and synchronization.
func (e *SyncedContextEnforcer) AddRoleForUserCtx(ctx context.Context, user string, role string, domain ...string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.ctxEnforcer.AddRoleForUserCtx(ctx, user, role, domain...)
}

// DeleteRoleForUserCtx deletes a role for a user with context and synchronization.
func (e *SyncedContextEnforcer) DeleteRoleForUserCtx(ctx context.Context, user string, role string, domain ...string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.ctxEnforcer.DeleteRoleForUserCtx(ctx, user, role, domain...)
}

// DeleteRolesForUserCtx deletes all roles for a user with context and synchronization.
func (e *SyncedContextEnforcer) DeleteRolesForUserCtx(ctx context.Context, user string, domain ...string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.ctxEnforcer.DeleteRolesForUserCtx(ctx, user, domain...)
}

// DeleteUserCtx deletes a user with context and synchronization.
func (e *SyncedContextEnforcer) DeleteUserCtx(ctx context.Context, user string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.ctxEnforcer.DeleteUserCtx(ctx, user)
}

// DeleteRoleCtx deletes a role with context and synchronization.
func (e *SyncedContextEnforcer) DeleteRoleCtx(ctx context.Context, role string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.ctxEnforcer.DeleteRoleCtx(ctx, role)
}

// DeletePermissionCtx deletes a permission with context and synchronization.
func (e *SyncedContextEnforcer) DeletePermissionCtx(ctx context.Context, permission ...string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.ctxEnforcer.DeletePermissionCtx(ctx, permission...)
}

// AddPermissionForUserCtx adds a permission for a user or role with context and synchronization.
func (e *SyncedContextEnforcer) AddPermissionForUserCtx(ctx context.Context, user string, permission ...string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.ctxEnforcer.AddPermissionForUserCtx(ctx, user, permission...)
}

// AddPermissionsForUserCtx adds multiple permissions for a user or role with context and synchronization.
func (e *SyncedContextEnforcer) AddPermissionsForUserCtx(ctx context.Context, user string, permissions ...[]string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.ctxEnforcer.AddPermissionsForUserCtx(ctx, user, permissions...)
}

// DeletePermissionForUserCtx deletes a permission for a user or role with context and synchronization.
func (e *SyncedContextEnforcer) DeletePermissionForUserCtx(ctx context.Context, user string, permission ...string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.ctxEnforcer.DeletePermissionForUserCtx(ctx, user, permission...)
}

// DeletePermissionsForUserCtx deletes permissions for a user or role with context and synchronization.
func (e *SyncedContextEnforcer) DeletePermissionsForUserCtx(ctx context.Context, user string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.ctxEnforcer.DeletePermissionsForUserCtx(ctx, user)
}

// AddRoleForUserInDomainCtx adds a role for a user inside a domain with context and synchronization.
func (e *SyncedContextEnforcer) AddRoleForUserInDomainCtx(ctx context.Context, user string, role string, domain string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.ctxEnforcer.AddRoleForUserInDomainCtx(ctx, user, role, domain)
}

// DeleteRoleForUserInDomainCtx deletes a role for a user inside a domain with context and synchronization.
func (e *SyncedContextEnforcer) DeleteRoleForUserInDomainCtx(ctx context.Context, user string, role string, domain string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.ctxEnforcer.DeleteRoleForUserInDomainCtx(ctx, user, role, domain)
}

// DeleteRolesForUserInDomainCtx deletes all roles for a user inside a domain with context and synchronization.
func (e *SyncedContextEnforcer) DeleteRolesForUserInDomainCtx(ctx context.Context, user string, domain string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.ctxEnforcer.DeleteRolesForUserInDomainCtx(ctx, user, domain)
}

// DeleteAllUsersByDomainCtx deletes all users associated with the domain with context and synchronization.
func (e *SyncedContextEnforcer) DeleteAllUsersByDomainCtx(ctx context.Context, domain string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.ctxEnforcer.DeleteAllUsersByDomainCtx(ctx, domain)
}

// DeleteDomainsCtx deletes all associated policies for domains with context and synchronization.
func (e *SyncedContextEnforcer) DeleteDomainsCtx(ctx context.Context, domains ...string) (bool, error) {
	e.m.Lock()
	defer e.m.Unlock()
	return e.ctxEnforcer.DeleteDomainsCtx(ctx, domains...)
}
