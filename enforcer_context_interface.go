// Copyright 2025 The casbin Authors. All Rights Reserved.
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

import "context"

type IEnforcerContext interface {
	IEnforcer

	/* Enforcer API */
	LoadPolicyCtx(ctx context.Context) error
	LoadFilteredPolicyCtx(ctx context.Context, filter interface{}) error
	LoadIncrementalFilteredPolicyCtx(ctx context.Context, filter interface{}) error
	IsFilteredCtx(ctx context.Context) bool
	SavePolicyCtx(ctx context.Context) error

	/* RBAC API */
	AddRoleForUserCtx(ctx context.Context, user string, role string, domain ...string) (bool, error)
	AddPermissionForUserCtx(ctx context.Context, user string, permission ...string) (bool, error)
	AddPermissionsForUserCtx(ctx context.Context, user string, permissions ...[]string) (bool, error)
	DeletePermissionForUserCtx(ctx context.Context, user string, permission ...string) (bool, error)
	DeletePermissionsForUserCtx(ctx context.Context, user string) (bool, error)

	DeleteRoleForUserCtx(ctx context.Context, user string, role string, domain ...string) (bool, error)
	DeleteRolesForUserCtx(ctx context.Context, user string, domain ...string) (bool, error)
	DeleteUserCtx(ctx context.Context, user string) (bool, error)
	DeleteRoleCtx(ctx context.Context, role string) (bool, error)
	DeletePermissionCtx(ctx context.Context, permission ...string) (bool, error)

	/* RBAC API with domains*/
	AddRoleForUserInDomainCtx(ctx context.Context, user string, role string, domain string) (bool, error)
	DeleteRoleForUserInDomainCtx(ctx context.Context, user string, role string, domain string) (bool, error)
	DeleteRolesForUserInDomainCtx(ctx context.Context, user string, domain string) (bool, error)
	DeleteAllUsersByDomainCtx(ctx context.Context, domain string) (bool, error)
	DeleteDomainsCtx(ctx context.Context, domains ...string) (bool, error)

	/* Management API */
	AddPolicyCtx(ctx context.Context, params ...interface{}) (bool, error)
	AddPoliciesCtx(ctx context.Context, rules [][]string) (bool, error)
	AddNamedPolicyCtx(ctx context.Context, ptype string, params ...interface{}) (bool, error)
	AddNamedPoliciesCtx(ctx context.Context, ptype string, rules [][]string) (bool, error)
	AddPoliciesExCtx(ctx context.Context, rules [][]string) (bool, error)
	AddNamedPoliciesExCtx(ctx context.Context, ptype string, rules [][]string) (bool, error)

	RemovePolicyCtx(ctx context.Context, params ...interface{}) (bool, error)
	RemovePoliciesCtx(ctx context.Context, rules [][]string) (bool, error)
	RemoveFilteredPolicyCtx(ctx context.Context, fieldIndex int, fieldValues ...string) (bool, error)
	RemoveNamedPolicyCtx(ctx context.Context, ptype string, params ...interface{}) (bool, error)
	RemoveNamedPoliciesCtx(ctx context.Context, ptype string, rules [][]string) (bool, error)
	RemoveFilteredNamedPolicyCtx(ctx context.Context, ptype string, fieldIndex int, fieldValues ...string) (bool, error)

	AddGroupingPolicyCtx(ctx context.Context, params ...interface{}) (bool, error)
	AddGroupingPoliciesCtx(ctx context.Context, rules [][]string) (bool, error)
	AddGroupingPoliciesExCtx(ctx context.Context, rules [][]string) (bool, error)
	AddNamedGroupingPolicyCtx(ctx context.Context, ptype string, params ...interface{}) (bool, error)
	AddNamedGroupingPoliciesCtx(ctx context.Context, ptype string, rules [][]string) (bool, error)
	AddNamedGroupingPoliciesExCtx(ctx context.Context, ptype string, rules [][]string) (bool, error)

	RemoveGroupingPolicyCtx(ctx context.Context, params ...interface{}) (bool, error)
	RemoveGroupingPoliciesCtx(ctx context.Context, rules [][]string) (bool, error)
	RemoveFilteredGroupingPolicyCtx(ctx context.Context, fieldIndex int, fieldValues ...string) (bool, error)
	RemoveNamedGroupingPolicyCtx(ctx context.Context, ptype string, params ...interface{}) (bool, error)
	RemoveNamedGroupingPoliciesCtx(ctx context.Context, ptype string, rules [][]string) (bool, error)
	RemoveFilteredNamedGroupingPolicyCtx(ctx context.Context, ptype string, fieldIndex int, fieldValues ...string) (bool, error)

	UpdatePolicyCtx(ctx context.Context, oldPolicy []string, newPolicy []string) (bool, error)
	UpdatePoliciesCtx(ctx context.Context, oldPolicies [][]string, newPolicies [][]string) (bool, error)
	UpdateFilteredPoliciesCtx(ctx context.Context, newPolicies [][]string, fieldIndex int, fieldValues ...string) (bool, error)

	UpdateGroupingPolicyCtx(ctx context.Context, oldRule []string, newRule []string) (bool, error)
	UpdateGroupingPoliciesCtx(ctx context.Context, oldRules [][]string, newRules [][]string) (bool, error)
	UpdateNamedGroupingPolicyCtx(ctx context.Context, ptype string, oldRule []string, newRule []string) (bool, error)
	UpdateNamedGroupingPoliciesCtx(ctx context.Context, ptype string, oldRules [][]string, newRules [][]string) (bool, error)

	/* Management API with autoNotifyWatcher disabled */
	SelfAddPolicyCtx(ctx context.Context, sec string, ptype string, rule []string) (bool, error)
	SelfAddPoliciesCtx(ctx context.Context, sec string, ptype string, rules [][]string) (bool, error)
	SelfAddPoliciesExCtx(ctx context.Context, sec string, ptype string, rules [][]string) (bool, error)
	SelfRemovePolicyCtx(ctx context.Context, sec string, ptype string, rule []string) (bool, error)
	SelfRemovePoliciesCtx(ctx context.Context, sec string, ptype string, rules [][]string) (bool, error)
	SelfRemoveFilteredPolicyCtx(ctx context.Context, sec string, ptype string, fieldIndex int, fieldValues ...string) (bool, error)
	SelfUpdatePolicyCtx(ctx context.Context, sec string, ptype string, oldRule, newRule []string) (bool, error)
	SelfUpdatePoliciesCtx(ctx context.Context, sec string, ptype string, oldRules, newRules [][]string) (bool, error)
}
