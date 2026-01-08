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

package detector

import (
	"fmt"
	"strings"

	"github.com/casbin/casbin/v3/model"
	"github.com/casbin/casbin/v3/rbac"
)

// ModelDetector defines the interface for detectors that need access to both the model and role manager.
type ModelDetector interface {
	// CheckModel checks whether the current status contains logical errors.
	// param: m Model instance
	// param: rm RoleManager instance
	// return: If an error is found, return a descriptive error; otherwise return nil.
	CheckModel(m model.Model, rm rbac.RoleManager) error
}

// EffectConflictDetector detects conflicts between user policies and role policies.
// It identifies cases where a user is explicitly allowed/denied to do something,
// but their role has the opposite effect for the same action.
//
// Note: In Casbin, explicit user policies override role policies, so such conflicts
// are not errors but might indicate policy design issues that should be reviewed.
// This detector is opt-in and not enabled by default.
//
// Example conflict:
//   p, alice, data2, write, deny
//   p, admin, data2, write, allow
//   g, alice, admin
// Here alice is explicitly denied but her role allows it - this might be intentional
// (to override the role permission) or it might be a mistake.
type EffectConflictDetector struct{}

// NewEffectConflictDetector creates a new instance of EffectConflictDetector.
//
// Usage example:
//   e, _ := casbin.NewEnforcer("model.conf", "policy.csv")
//   e.SetDetectors([]detector.Detector{
//       detector.NewDefaultDetector(),
//       detector.NewEffectConflictDetector(),
//   })
//   err := e.RunDetections()
func NewEffectConflictDetector() *EffectConflictDetector {
	return &EffectConflictDetector{}
}

// CheckModel checks for effect conflicts between user and role policies.
func (d *EffectConflictDetector) CheckModel(m model.Model, rm rbac.RoleManager) error {
	if m == nil {
		return fmt.Errorf("model cannot be nil")
	}
	if rm == nil {
		return fmt.Errorf("role manager cannot be nil")
	}

	// Get all policies
	policies, err := m.GetPolicy("p", "p")
	if err != nil {
		return err
	}

	// Get all role assignments
	roles, err := m.GetPolicy("g", "g")
	if err != nil {
		// If no role assignments, no conflicts possible
		return nil
	}

	// Build a map of user -> roles
	userRoles := make(map[string][]string)
	for _, role := range roles {
		if len(role) < 2 {
			continue
		}
		user := role[0]
		roleName := role[1]
		userRoles[user] = append(userRoles[user], roleName)
	}

	// Build a map of (subject, object, action) -> effect
	policyEffects := make(map[string]string)
	for _, policy := range policies {
		if len(policy) < 3 {
			continue
		}
		subject := policy[0]
		object := policy[1]
		action := policy[2]
		effect := "allow" // Default effect if not specified
		if len(policy) >= 4 {
			effect = policy[3]
		}
		
		key := fmt.Sprintf("%s:%s:%s", subject, object, action)
		policyEffects[key] = effect
	}

	// Check for conflicts
	for user, roleList := range userRoles {
		for _, roleName := range roleList {
			// Check all policy combinations
			for policyKey, effect := range policyEffects {
				parts := strings.Split(policyKey, ":")
				if len(parts) != 3 {
					continue
				}
				subject := parts[0]
				object := parts[1]
				action := parts[2]

				// Check if this is a user policy
				if subject == user {
					// Check if any role has opposite effect
					roleKey := fmt.Sprintf("%s:%s:%s", roleName, object, action)
					if roleEffect, exists := policyEffects[roleKey]; exists {
						if (effect == "allow" && roleEffect == "deny") ||
							(effect == "deny" && roleEffect == "allow") {
							return fmt.Errorf(
								"effect conflict detected: user '%s' has '%s' effect for (%s, %s), "+
									"but role '%s' has '%s' effect for the same action",
								user, effect, object, action, roleName, roleEffect)
						}
					}
				} else if subject == roleName {
					// Check if user has opposite effect
					userKey := fmt.Sprintf("%s:%s:%s", user, object, action)
					if userEffect, exists := policyEffects[userKey]; exists {
						if (effect == "allow" && userEffect == "deny") ||
							(effect == "deny" && userEffect == "allow") {
							return fmt.Errorf(
								"effect conflict detected: user '%s' has '%s' effect for (%s, %s), "+
									"but role '%s' has '%s' effect for the same action",
								user, userEffect, object, action, roleName, effect)
						}
					}
				}
			}
		}
	}

	return nil
}

// Check implements the Detector interface by returning an error indicating this detector needs model access.
func (d *EffectConflictDetector) Check(rm rbac.RoleManager) error {
	return fmt.Errorf("EffectConflictDetector requires model access, use CheckModel instead")
}
