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
	return e.GetFilteredPolicy(0, user, domain)
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
	if len(g.Tokens) != 3 {
		return []string{}
	}
	users := make([]string, 0)
	for _, policy := range g.Policy {
		if _, ok := m[policy[2]]; policy[2] == domain && ok {
			users = append(users, policy[0])
		}
	}
	return users
}

// DeleteAllUsersByDomain would delete all users associated with the domain.
func (e *Enforcer) DeleteAllUsersByDomain(domain string) (bool, error) {
	g := e.model["g"]["g"]
	if len(g.Tokens) != 3 {
		return false, nil
	}
	policies := make([][]string, 0)
	for _, policy := range g.Policy {
		if policy[3] == domain {
			policies = append(policies, policy)
		}
	}
	return e.RemoveGroupingPolicies(policies)
}

// DeleteDomains would delete all associated users and roles.
// It would delete all domains if parameter is not provided.
func (e *Enforcer) DeleteDomains(domains ...string) (bool, error) {
	g := e.model["g"]["g"]
	if len(g.Tokens) != 3 {
		return false, nil
	}
	if len(domains) == 0 {
		return e.RemoveGroupingPolicies(g.Policy)
	}
	m := make(map[string]struct{})
	for _, domain := range domains {
		m[domain] = struct{}{}
	}
	policies := make([][]string, 0)
	for _, policy := range g.Policy {
		if _, ok := m[policy[2]]; ok {
			policies = append(policies, policy)
		}
	}
	return e.RemoveGroupingPolicies(policies)
}
