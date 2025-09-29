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

import (
	"context"
	"fmt"

	"github.com/casbin/casbin/v2/constant"
)

// AddRoleForUserInDomainCtx adds a role for a user inside a domain with context support.
// Returns false if the user already has the role (aka not affected).
func (e *ContextEnforcer) AddRoleForUserInDomainCtx(ctx context.Context, user string, role string, domain string) (bool, error) {
	return e.AddGroupingPolicyCtx(ctx, user, role, domain)
}

// DeleteRoleForUserInDomainCtx deletes a role for a user inside a domain with context support.
// Returns false if the user does not have the role (aka not affected).
func (e *ContextEnforcer) DeleteRoleForUserInDomainCtx(ctx context.Context, user string, role string, domain string) (bool, error) {
	return e.RemoveGroupingPolicyCtx(ctx, user, role, domain)
}

// DeleteRolesForUserInDomainCtx deletes all roles for a user inside a domain with context support.
// Returns false if the user does not have any roles (aka not affected).
func (e *ContextEnforcer) DeleteRolesForUserInDomainCtx(ctx context.Context, user string, domain string) (bool, error) {
	if e.GetRoleManager() == nil {
		return false, fmt.Errorf("role manager is not initialized")
	}
	roles, err := e.GetRoleManager().GetRoles(user, domain)
	if err != nil {
		return false, err
	}

	var rules [][]string
	for _, role := range roles {
		rules = append(rules, []string{user, role, domain})
	}

	return e.RemoveGroupingPoliciesCtx(ctx, rules)
}

// DeleteAllUsersByDomainCtx deletes all users associated with the domain with context support.
func (e *ContextEnforcer) DeleteAllUsersByDomainCtx(ctx context.Context, domain string) (bool, error) {
	g, err := e.model.GetAssertion("g", "g")
	if err != nil {
		return false, err
	}
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
	if _, err = e.RemoveGroupingPoliciesCtx(ctx, users); err != nil {
		return false, err
	}
	users = getUser(index, p.Policy, domain)
	if _, err = e.RemovePoliciesCtx(ctx, users); err != nil {
		return false, err
	}
	return true, nil
}

// DeleteDomainsCtx deletes all associated policies for domains with context support.
// It would delete all domains if parameter is not provided.
func (e *ContextEnforcer) DeleteDomainsCtx(ctx context.Context, domains ...string) (bool, error) {
	if len(domains) == 0 {
		var err error
		domains, err = e.GetAllDomains()
		if err != nil {
			return false, err
		}
	}
	for _, domain := range domains {
		if _, err := e.DeleteAllUsersByDomainCtx(ctx, domain); err != nil {
			return false, err
		}
		// remove the domain from the RoleManager.
		if e.GetRoleManager() != nil {
			if err := e.GetRoleManager().DeleteDomain(domain); err != nil {
				return false, err
			}
		}
	}
	return true, nil
}
