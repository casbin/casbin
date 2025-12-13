// Copyright 2024 The casbin Authors. All Rights Reserved.
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

package model

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/casbin/casbin/v3/errors"
)

// ConstraintType represents the type of constraint.
type ConstraintType int

const (
	ConstraintTypeSOD ConstraintType = iota
	ConstraintTypeSODMax
	ConstraintTypeRoleMax
	ConstraintTypeRolePre
)

// Constraint represents a policy constraint.
type Constraint struct {
	Key        string
	Type       ConstraintType
	Roles      []string
	Role       string
	MaxCount   int
	PreReqRole string
}

var (
	// Regex patterns for parsing constraints
	sodPattern     = regexp.MustCompile(`^sod\s*\(\s*"([^"]+)"\s*,\s*"([^"]+)"\s*\)$`)
	sodMaxPattern  = regexp.MustCompile(`^sodMax\s*\(\s*\[([^\]]+)\]\s*,\s*(\d+)\s*\)$`)
	roleMaxPattern = regexp.MustCompile(`^roleMax\s*\(\s*"([^"]+)"\s*,\s*(\d+)\s*\)$`)
	rolePrePattern = regexp.MustCompile(`^rolePre\s*\(\s*"([^"]+)"\s*,\s*"([^"]+)"\s*\)$`)
)

// parseConstraint parses a constraint definition string.
func parseConstraint(key, value string) (*Constraint, error) {
	value = strings.TrimSpace(value)

	// Try to match sod pattern
	if matches := sodPattern.FindStringSubmatch(value); matches != nil {
		return &Constraint{
			Key:   key,
			Type:  ConstraintTypeSOD,
			Roles: []string{matches[1], matches[2]},
		}, nil
	}

	// Try to match sodMax pattern
	if matches := sodMaxPattern.FindStringSubmatch(value); matches != nil {
		rolesStr := matches[1]
		maxCount, err := strconv.Atoi(matches[2])
		if err != nil {
			return nil, fmt.Errorf("invalid max count in sodMax: %w", err)
		}
		
		// Parse the roles array
		var roles []string
		for _, role := range strings.Split(rolesStr, ",") {
			role = strings.TrimSpace(role)
			role = strings.Trim(role, `"`)
			if role != "" {
				roles = append(roles, role)
			}
		}
		
		if len(roles) == 0 {
			return nil, fmt.Errorf("sodMax requires at least one role")
		}
		
		return &Constraint{
			Key:      key,
			Type:     ConstraintTypeSODMax,
			Roles:    roles,
			MaxCount: maxCount,
		}, nil
	}

	// Try to match roleMax pattern
	if matches := roleMaxPattern.FindStringSubmatch(value); matches != nil {
		maxCount, err := strconv.Atoi(matches[2])
		if err != nil {
			return nil, fmt.Errorf("invalid max count in roleMax: %w", err)
		}
		return &Constraint{
			Key:      key,
			Type:     ConstraintTypeRoleMax,
			Role:     matches[1],
			MaxCount: maxCount,
		}, nil
	}

	// Try to match rolePre pattern
	if matches := rolePrePattern.FindStringSubmatch(value); matches != nil {
		return &Constraint{
			Key:        key,
			Type:       ConstraintTypeRolePre,
			Role:       matches[1],
			PreReqRole: matches[2],
		}, nil
	}

	return nil, fmt.Errorf("unrecognized constraint format: %s", value)
}

// ValidateConstraints validates all constraints against the current policy.
func (model Model) ValidateConstraints() error {
	// Check if constraints exist
	if model["c"] == nil || len(model["c"]) == 0 {
		return nil // No constraints to validate
	}

	// Check if RBAC is enabled
	if model["g"] == nil || len(model["g"]) == 0 {
		return errors.ErrConstraintRequiresRBAC
	}

	// Get grouping policy
	gAssertion := model["g"]["g"]
	if gAssertion == nil {
		return errors.ErrConstraintRequiresRBAC
	}

	// Validate each constraint
	for _, assertion := range model["c"] {
		constraint, err := parseConstraint(assertion.Key, assertion.Value)
		if err != nil {
			return fmt.Errorf("%w: %s", errors.ErrConstraintParsingError, err.Error())
		}

		if err := model.validateConstraint(constraint, gAssertion.Policy); err != nil {
			return err
		}
	}

	return nil
}

// validateConstraint validates a single constraint against the policy.
func (model Model) validateConstraint(constraint *Constraint, groupingPolicy [][]string) error {
	switch constraint.Type {
	case ConstraintTypeSOD:
		return model.validateSOD(constraint, groupingPolicy)
	case ConstraintTypeSODMax:
		return model.validateSODMax(constraint, groupingPolicy)
	case ConstraintTypeRoleMax:
		return model.validateRoleMax(constraint, groupingPolicy)
	case ConstraintTypeRolePre:
		return model.validateRolePre(constraint, groupingPolicy)
	default:
		return fmt.Errorf("unknown constraint type")
	}
}

// validateSOD validates a Separation of Duties constraint.
func (model Model) validateSOD(constraint *Constraint, groupingPolicy [][]string) error {
	if len(constraint.Roles) != 2 {
		return errors.NewConstraintViolationError(constraint.Key, "sod requires exactly 2 roles")
	}

	role1, role2 := constraint.Roles[0], constraint.Roles[1]
	userRoles := make(map[string]map[string]bool)

	// Build a map of users to their roles
	for _, rule := range groupingPolicy {
		if len(rule) < 2 {
			continue
		}
		user := rule[0]
		role := rule[1]
		
		if userRoles[user] == nil {
			userRoles[user] = make(map[string]bool)
		}
		userRoles[user][role] = true
	}

	// Check if any user has both roles
	for user, roles := range userRoles {
		if roles[role1] && roles[role2] {
			return errors.NewConstraintViolationError(constraint.Key, 
				fmt.Sprintf("user '%s' cannot have both roles '%s' and '%s'", user, role1, role2))
		}
	}

	return nil
}

// validateSODMax validates a maximum role count constraint for a role set.
func (model Model) validateSODMax(constraint *Constraint, groupingPolicy [][]string) error {
	userRoles := make(map[string]map[string]bool)

	// Build a map of users to their roles
	for _, rule := range groupingPolicy {
		if len(rule) < 2 {
			continue
		}
		user := rule[0]
		role := rule[1]
		
		if userRoles[user] == nil {
			userRoles[user] = make(map[string]bool)
		}
		userRoles[user][role] = true
	}

	// Check if any user has more than maxCount roles from the role set
	for user, roles := range userRoles {
		count := 0
		for _, role := range constraint.Roles {
			if roles[role] {
				count++
			}
		}
		if count > constraint.MaxCount {
			return errors.NewConstraintViolationError(constraint.Key,
				fmt.Sprintf("user '%s' has %d roles from %v, exceeds maximum of %d", 
					user, count, constraint.Roles, constraint.MaxCount))
		}
	}

	return nil
}

// validateRoleMax validates a role cardinality constraint.
func (model Model) validateRoleMax(constraint *Constraint, groupingPolicy [][]string) error {
	roleCount := 0

	// Count how many users have this role
	for _, rule := range groupingPolicy {
		if len(rule) < 2 {
			continue
		}
		role := rule[1]
		
		if role == constraint.Role {
			roleCount++
		}
	}

	if roleCount > constraint.MaxCount {
		return errors.NewConstraintViolationError(constraint.Key,
			fmt.Sprintf("role '%s' assigned to %d users, exceeds maximum of %d", 
				constraint.Role, roleCount, constraint.MaxCount))
	}

	return nil
}

// validateRolePre validates a prerequisite role constraint.
func (model Model) validateRolePre(constraint *Constraint, groupingPolicy [][]string) error {
	userRoles := make(map[string]map[string]bool)

	// Build a map of users to their roles
	for _, rule := range groupingPolicy {
		if len(rule) < 2 {
			continue
		}
		user := rule[0]
		role := rule[1]
		
		if userRoles[user] == nil {
			userRoles[user] = make(map[string]bool)
		}
		userRoles[user][role] = true
	}

	// Check if any user has the main role without the prerequisite role
	for user, roles := range userRoles {
		if roles[constraint.Role] && !roles[constraint.PreReqRole] {
			return errors.NewConstraintViolationError(constraint.Key,
				fmt.Sprintf("user '%s' has role '%s' but lacks prerequisite role '%s'", 
					user, constraint.Role, constraint.PreReqRole))
		}
	}

	return nil
}
