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

// rbac_api_context.go
package casbin

import (
	"context"

	"github.com/casbin/casbin/v2/constant"
	"github.com/casbin/casbin/v2/errors"
	"github.com/casbin/casbin/v2/util"
)

// AddRoleForUserCtx adds a role for a user with context support.
// Returns false if the user already has the role (aka not affected).
func (e *ContextEnforcer) AddRoleForUserCtx(ctx context.Context, user string, role string, domain ...string) (bool, error) {
	args := []string{user, role}
	args = append(args, domain...)
	return e.AddGroupingPolicyCtx(ctx, args)
}

// DeleteRoleForUserCtx deletes a role for a user with context support.
// Returns false if the user does not have the role (aka not affected).
func (e *ContextEnforcer) DeleteRoleForUserCtx(ctx context.Context, user string, role string, domain ...string) (bool, error) {
	args := []string{user, role}
	args = append(args, domain...)
	return e.RemoveGroupingPolicyCtx(ctx, args)
}

// DeleteRolesForUserCtx deletes all roles for a user with context support.
// Returns false if the user does not have any roles (aka not affected).
func (e *ContextEnforcer) DeleteRolesForUserCtx(ctx context.Context, user string, domain ...string) (bool, error) {
	var args []string
	if len(domain) == 0 {
		args = []string{user}
	} else if len(domain) > 1 {
		return false, errors.ErrDomainParameter
	} else {
		args = []string{user, "", domain[0]}
	}
	return e.RemoveFilteredGroupingPolicyCtx(ctx, 0, args...)
}

// DeleteUserCtx deletes a user with context support.
// Returns false if the user does not exist (aka not affected).
func (e *ContextEnforcer) DeleteUserCtx(ctx context.Context, user string) (bool, error) {
	var err error
	res1, err := e.RemoveFilteredGroupingPolicyCtx(ctx, 0, user)
	if err != nil {
		return res1, err
	}

	subIndex, err := e.GetFieldIndex("p", constant.SubjectIndex)
	if err != nil {
		return false, err
	}
	res2, err := e.RemoveFilteredPolicyCtx(ctx, subIndex, user)
	return res1 || res2, err
}

// DeleteRoleCtx deletes a role with context support.
// Returns false if the role does not exist (aka not affected).
func (e *ContextEnforcer) DeleteRoleCtx(ctx context.Context, role string) (bool, error) {
	var err error
	res1, err := e.RemoveFilteredGroupingPolicyCtx(ctx, 0, role)
	if err != nil {
		return res1, err
	}

	res2, err := e.RemoveFilteredGroupingPolicyCtx(ctx, 1, role)
	if err != nil {
		return res1, err
	}

	subIndex, err := e.GetFieldIndex("p", constant.SubjectIndex)
	if err != nil {
		return false, err
	}
	res3, err := e.RemoveFilteredPolicyCtx(ctx, subIndex, role)
	return res1 || res2 || res3, err
}

// DeletePermissionCtx deletes a permission with context support.
// Returns false if the permission does not exist (aka not affected).
func (e *ContextEnforcer) DeletePermissionCtx(ctx context.Context, permission ...string) (bool, error) {
	return e.RemoveFilteredPolicyCtx(ctx, 1, permission...)
}

// AddPermissionForUserCtx adds a permission for a user or role with context support.
// Returns false if the user or role already has the permission (aka not affected).
func (e *ContextEnforcer) AddPermissionForUserCtx(ctx context.Context, user string, permission ...string) (bool, error) {
	return e.AddPolicyCtx(ctx, util.JoinSlice(user, permission...))
}

// AddPermissionsForUserCtx adds multiple permissions for a user or role with context support.
// Returns false if the user or role already has one of the permissions (aka not affected).
func (e *ContextEnforcer) AddPermissionsForUserCtx(ctx context.Context, user string, permissions ...[]string) (bool, error) {
	var rules [][]string
	for _, permission := range permissions {
		rules = append(rules, util.JoinSlice(user, permission...))
	}
	return e.AddPoliciesCtx(ctx, rules)
}

// DeletePermissionForUserCtx deletes a permission for a user or role with context support.
// Returns false if the user or role does not have the permission (aka not affected).
func (e *ContextEnforcer) DeletePermissionForUserCtx(ctx context.Context, user string, permission ...string) (bool, error) {
	return e.RemovePolicyCtx(ctx, util.JoinSlice(user, permission...))
}

// DeletePermissionsForUserCtx deletes permissions for a user or role with context support.
// Returns false if the user or role does not have any permissions (aka not affected).
func (e *ContextEnforcer) DeletePermissionsForUserCtx(ctx context.Context, user string) (bool, error) {
	subIndex, err := e.GetFieldIndex("p", constant.SubjectIndex)
	if err != nil {
		return false, err
	}
	return e.RemoveFilteredPolicyCtx(ctx, subIndex, user)
}
