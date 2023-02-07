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

import "github.com/casbin/casbin/v2/constant"

// GetUsersForRoleInDomain gets the users that has a role inside a domain. Add by Gordon
func (e *Enforcer) GetUsersForRoleInDomain(name string, domain string) []string {
	res, _ := e.model["g"]["g"].RM.GetUsers(name, domain)
	return res
}

// GetRolesForUserInDomain gets the roles that a user has inside a domain.
func (e *Enforcer) GetRolesForUserInDomain(name string, domain string) []string {
	res, _ := e.model["g"]["g"].RM.GetRoles(name, domain)
	return res
}

// GetPermissionsForUserInDomain gets permissions for a user or role inside a domain.
func (e *Enforcer) GetPermissionsForUserInDomain(user string, domain string) [][]string {
	res, _ := e.GetImplicitPermissionsForUser(user, domain)
	return res
}

// AddRoleForUserInDomain adds a role for a user inside a domain.
// Returns false if the user already has the role (aka not affected).
func (e *Enforcer) AddRoleForUserInDomain(user string, role string, domain string) (bool, error) {
	return e.AddGroupingPolicy(user, role, domain)
}

// DeleteRoleForUserInDomain deletes a role for a user inside a domain.
// Returns false if the user does not have the role (aka not affected).
func (e *Enforcer) DeleteRoleForUserInDomain(user string, role string, domain string) (bool, error) {
	return e.RemoveGroupingPolicy(user, role, domain)
}

// DeleteRolesForUserInDomain deletes all roles for a user inside a domain.
// Returns false if the user does not have any roles (aka not affected).
func (e *Enforcer) DeleteRolesForUserInDomain(user string, domain string) (bool, error) {
	roles, err := e.model["g"]["g"].RM.GetRoles(user, domain)
	if err != nil {
		return false, err
	}

	var rules [][]string
	for _, role := range roles {
		rules = append(rules, []string{user, role, domain})
	}

	return e.RemoveGroupingPolicies(rules)
}

// GetAllUsersByDomain would get all users associated with the domain.
func (e *Enforcer) GetAllUsersByDomain(domain string) []string {
	m := make(map[string]struct{})
	g := e.model["g"]["g"]
	p := e.model["p"]["p"]
	users := make([]string, 0)
	index, err := e.GetFieldIndex("p", constant.DomainIndex)
	if err != nil {
		return []string{}
	}

	getUser := func(index int, policies [][]string, domain string, m map[string]struct{}) []string {
		if len(policies) == 0 || len(policies[0]) <= index {
			return []string{}
		}
		res := make([]string, 0)
		for _, policy := range policies {
			if _, ok := m[policy[0]]; policy[index] == domain && !ok {
				res = append(res, policy[0])
				m[policy[0]] = struct{}{}
			}
		}
		return res
	}

	users = append(users, getUser(2, g.Policy, domain, m)...)
	users = append(users, getUser(index, p.Policy, domain, m)...)
	return users
}

// DeleteAllUsersByDomain would delete all users associated with the domain.
func (e *Enforcer) DeleteAllUsersByDomain(domain string) (bool, error) {
	g := e.model["g"]["g"]
	p := e.model["p"]["p"]
	index, err := e.GetFieldIndex("p", constant.DomainIndex)
	if err != nil {
		return false, err
	}

	getUser := func(index int, policies [][]string, domain string) [][]string {
		if len(policies) == 0 || len(policies[0]) <= index {
			return [][]string{}
		}
		res := make([][]string, 0)
		for _, policy := range policies {
			if policy[index] == domain {
				res = append(res, policy)
			}
		}
		return res
	}

	users := getUser(2, g.Policy, domain)
	if _, err := e.RemoveGroupingPolicies(users); err != nil {
		return false, err
	}
	users = getUser(index, p.Policy, domain)
	if _, err := e.RemovePolicies(users); err != nil {
		return false, err
	}
	return true, nil
}

// DeleteDomains would delete all associated users and roles.
// It would delete all domains if parameter is not provided.
func (e *Enforcer) DeleteDomains(domains ...string) (bool, error) {
	if len(domains) == 0 {
		e.ClearPolicy()
		return true, nil
	}
	for _, domain := range domains {
		if _, err := e.DeleteAllUsersByDomain(domain); err != nil {
			return false, err
		}
	}
	return true, nil
}

// GetAllDomains would get all domains.
func (e *Enforcer) GetAllDomains() ([]string, error) {
	return e.model["g"]["g"].RM.GetAllDomains()
}

// GetAllRolesByDomain would get all roles associated with the domain.
// note: Not applicable to Domains with inheritance relationship  (implicit roles)
func (e *Enforcer) GetAllRolesByDomain(domain string) []string {
	g := e.model["g"]["g"]
	policies := g.Policy
	roles := make([]string, 0)
	existMap := make(map[string]bool) // remove duplicates

	for _, policy := range policies {
		if policy[len(policy)-1] == domain {
			role := policy[len(policy)-2]
			if _, ok := existMap[role]; !ok {
				roles = append(roles, role)
				existMap[role] = true
			}
		}
	}

	return roles
}
