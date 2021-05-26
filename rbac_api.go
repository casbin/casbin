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
	"github.com/casbin/casbin/v2/errors"
	"github.com/casbin/casbin/v2/util"
)

// GetRolesForUser gets the roles that a user has.
func (e *Enforcer) GetRolesForUser(name string, domain ...string) ([]string, error) {
	res, err := e.model["g"]["g"].RM.GetRoles(name, domain...)
	return res, err
}

// GetUsersForRole gets the users that has a role.
func (e *Enforcer) GetUsersForRole(name string, domain ...string) ([]string, error) {
	res, err := e.model["g"]["g"].RM.GetUsers(name, domain...)
	return res, err
}

// HasRoleForUser determines whether a user has a role.
func (e *Enforcer) HasRoleForUser(name string, role string, domain ...string) (bool, error) {
	roles, err := e.GetRolesForUser(name, domain...)
	if err != nil {
		return false, err
	}
	hasRole := false
	for _, r := range roles {
		if r == role {
			hasRole = true
			break
		}
	}

	return hasRole, nil
}

// AddRoleForUser adds a role for a user.
// Returns false if the user already has the role (aka not affected).
func (e *Enforcer) AddRoleForUser(user string, role string, domain ...string) (bool, error) {
	args := []string{user, role}
	args = append(args, domain...)
	return e.AddGroupingPolicy(args)
}

// AddRolesForUser adds roles for a user.
// Returns false if the user already has the roles (aka not affected).
func (e *Enforcer) AddRolesForUser(user string, roles []string, domain ...string) (bool, error) {
	var rules [][]string
	for _, role := range roles {
		rule := []string{user, role}
		rule = append(rule, domain...)
		rules = append(rules, rule)
	}
	return e.AddGroupingPolicies(rules)
}

// DeleteRoleForUser deletes a role for a user.
// Returns false if the user does not have the role (aka not affected).
func (e *Enforcer) DeleteRoleForUser(user string, role string, domain ...string) (bool, error) {
	args := []string{user, role}
	args = append(args, domain...)
	return e.RemoveGroupingPolicy(args)
}

// DeleteRolesForUser deletes all roles for a user.
// Returns false if the user does not have any roles (aka not affected).
func (e *Enforcer) DeleteRolesForUser(user string, domain ...string) (bool, error) {
	var args []string
	if len(domain) == 0 {
		args = []string{user}
	} else if len(domain) > 1 {
		return false, errors.ERR_DOMAIN_PARAMETER
	} else {
		args = []string{user, "", domain[0]}
	}
	return e.RemoveFilteredGroupingPolicy(0, args...)
}

// DeleteUser deletes a user.
// Returns false if the user does not exist (aka not affected).
func (e *Enforcer) DeleteUser(user string) (bool, error) {
	var err error
	res1, err := e.RemoveFilteredGroupingPolicy(0, user)
	if err != nil {
		return res1, err
	}

	res2, err := e.RemoveFilteredPolicy(0, user)
	return res1 || res2, err
}

// DeleteRole deletes a role.
// Returns false if the role does not exist (aka not affected).
func (e *Enforcer) DeleteRole(role string) (bool, error) {
	var err error
	res1, err := e.RemoveFilteredGroupingPolicy(1, role)
	if err != nil {
		return res1, err
	}

	res2, err := e.RemoveFilteredPolicy(0, role)
	return res1 || res2, err
}

// DeletePermission deletes a permission.
// Returns false if the permission does not exist (aka not affected).
func (e *Enforcer) DeletePermission(permission ...string) (bool, error) {
	return e.RemoveFilteredPolicy(1, permission...)
}

// AddPermissionForUser adds a permission for a user or role.
// Returns false if the user or role already has the permission (aka not affected).
func (e *Enforcer) AddPermissionForUser(user string, permission ...string) (bool, error) {
	return e.AddPolicy(util.JoinSlice(user, permission...))
}

// DeletePermissionForUser deletes a permission for a user or role.
// Returns false if the user or role does not have the permission (aka not affected).
func (e *Enforcer) DeletePermissionForUser(user string, permission ...string) (bool, error) {
	return e.RemovePolicy(util.JoinSlice(user, permission...))
}

// DeletePermissionsForUser deletes permissions for a user or role.
// Returns false if the user or role does not have any permissions (aka not affected).
func (e *Enforcer) DeletePermissionsForUser(user string) (bool, error) {
	return e.RemoveFilteredPolicy(0, user)
}

// GetPermissionsForUser gets permissions for a user or role.
func (e *Enforcer) GetPermissionsForUser(user string, domain ...string) [][]string {
	permission := make([][]string, 0)
	for ptype, assertion := range e.model["p"] {
		args := make([]string, len(assertion.Tokens))
		args[0] = user

		if len(domain) > 0 {
			index := e.getDomainIndex(ptype)
			if index < len(assertion.Tokens) {
				args[index] = domain[0]
			}
		}
		perm := e.GetFilteredPolicy(0, args...)
		permission = append(permission, perm...)
	}
	return permission
}

// HasPermissionForUser determines whether a user has a permission.
func (e *Enforcer) HasPermissionForUser(user string, permission ...string) bool {
	return e.HasPolicy(util.JoinSlice(user, permission...))
}

// GetImplicitRolesForUser gets implicit roles that a user has.
// Compared to GetRolesForUser(), this function retrieves indirect roles besides direct roles.
// For example:
// g, alice, role:admin
// g, role:admin, role:user
//
// GetRolesForUser("alice") can only get: ["role:admin"].
// But GetImplicitRolesForUser("alice") will get: ["role:admin", "role:user"].
func (e *Enforcer) GetImplicitRolesForUser(name string, domain ...string) ([]string, error) {
	res := []string{}
	roleSet := make(map[string]bool)
	roleSet[name] = true

	q := make([]string, 0)
	q = append(q, name)

	for len(q) > 0 {
		name := q[0]
		q = q[1:]

		for _, rm := range e.rmMap {
			roles, err := rm.GetRoles(name, domain...)
			if err != nil {
				return nil, err
			}
			for _, r := range roles {
				if _, ok := roleSet[r]; !ok {
					res = append(res, r)
					q = append(q, r)
					roleSet[r] = true
				}
			}
		}
	}

	return res, nil
}

// GetImplicitUsersForRole gets implicit users for a role.
func (e *Enforcer) GetImplicitUsersForRole(name string, domain ...string) ([]string, error) {
	res := []string{}
	roleSet := make(map[string]bool)
	roleSet[name] = true

	q := make([]string, 0)
	q = append(q, name)

	for len(q) > 0 {
		name := q[0]
		q = q[1:]

		for _, rm := range e.rmMap {
			roles, err := rm.GetUsers(name, domain...)
			if err != nil && err.Error() != "error: name does not exist" {
				return nil, err
			}
			for _, r := range roles {
				if _, ok := roleSet[r]; !ok {
					res = append(res, r)
					q = append(q, r)
					roleSet[r] = true
				}
			}
		}
	}

	return res, nil
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
func (e *Enforcer) GetImplicitPermissionsForUser(user string, domain ...string) ([][]string, error) {
	roles, err := e.GetImplicitRolesForUser(user, domain...)
	if err != nil {
		return nil, err
	}

	roles = append([]string{user}, roles...)

	var res [][]string
	var permissions [][]string
	for _, role := range roles {
		permissions = e.GetPermissionsForUser(role, domain...)

		res = append(res, permissions...)
	}

	return res, nil
}

// GetImplicitUsersForPermission gets implicit users for a permission.
// For example:
// p, admin, data1, read
// p, bob, data1, read
// g, alice, admin
//
// GetImplicitUsersForPermission("data1", "read") will get: ["alice", "bob"].
// Note: only users will be returned, roles (2nd arg in "g") will be excluded.
func (e *Enforcer) GetImplicitUsersForPermission(permission ...string) ([]string, error) {
	pSubjects := e.GetAllSubjects()
	gInherit := e.model.GetValuesForFieldInPolicyAllTypes("g", 1)
	gSubjects := e.model.GetValuesForFieldInPolicyAllTypes("g", 0)

	subjects := append(pSubjects, gSubjects...)
	util.ArrayRemoveDuplicates(&subjects)

	subjects = util.SetSubtract(subjects, gInherit)

	res := []string{}
	for _, user := range subjects {
		req := util.JoinSliceAny(user, permission...)
		allowed, err := e.Enforce(req...)
		if err != nil {
			return nil, err
		}

		if allowed {
			res = append(res, user)
		}
	}

	return res, nil
}

// GetDomainsForUser gets all domains
func (e *Enforcer) GetDomainsForUser(user string) ([]string, error) {
	var domains []string
	for _, rm := range e.rmMap {
		domain, err := rm.GetDomains(user)
		if err != nil {
			return nil, err
		}
		domains = append(domains, domain...)
	}
	return domains, nil
}

// GetImplicitResourcesForUser returns all policies that user obtaining in domain
func (e *Enforcer) GetImplicitResourcesForUser(user string, domain ...string) ([][]string, error) {
	permissions, err := e.GetImplicitPermissionsForUser(user, domain...)
	if err != nil {
		return nil, err
	}
	res := make([][]string, 0)
	for _, permission := range permissions {
		if permission[0] == user {
			res = append(res, permission)
			continue
		}
		resLocal := [][]string{{user}}
		tokensLength := len(permission)
		t := make([][]string, 1, tokensLength)
		for _, token := range permission[1:] {
			tokens, err := e.GetImplicitUsersForRole(token, domain...)
			if err != nil {
				return nil, err
			}
			tokens = append(tokens, token)
			t = append(t, tokens)
		}
		for i := 1; i < tokensLength; i++ {
			n := make([][]string, 0)
			for _, tokens := range t[i] {
				for _, policy := range resLocal {
					t := append([]string(nil), policy...)
					t = append(t, tokens)
					n = append(n, t)
				}
			}
			resLocal = n
		}
		res = append(res, resLocal...)
	}
	return res, nil
}
