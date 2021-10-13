// Copyright 2018 The casbin Authors. All Rights Reserved.
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

package effector

import "errors"

// DefaultEffector is default effector for Casbin.
type DefaultEffector struct {
}

const (
	AllowOverride   = "some(where (p_eft == allow))"
	DenyOverride    = "!some(where (p_eft == deny))"
	AllowAndDeny    = "some(where (p_eft == allow)) && !some(where (p_eft == deny))"
	Priority        = "priority(p_eft) || deny"
	SubjectPriority = "subjectPriority(p_eft) || deny"
)

// NewDefaultEffector is the constructor for DefaultEffector.
func NewDefaultEffector() *DefaultEffector {
	e := DefaultEffector{}
	return &e
}

// MergeEffects merges all matching results collected by the enforcer into a single decision.
func (e *DefaultEffector) MergeEffects(expr string, effects []Effect, matches []float64) (Effect, int, error) {
	result := Indeterminate
	hitPolicyIndex := -1

	switch expr {
	case AllowOverride:
		for i, eft := range effects {
			if matches[i] == 0 {
				continue
			}

			if eft == Allow {
				result = Allow
				hitPolicyIndex = i
				break
			}
		}
	case DenyOverride:
		// if no deny rules are matched, then allow
		result = Allow
		for i, eft := range effects {
			if matches[i] == 0 {
				continue
			}

			if eft == Deny {
				result = Deny
				hitPolicyIndex = i
				break
			}
		}
	case AllowAndDeny:
		for i, eft := range effects {
			if matches[i] == 0 {
				continue
			}

			if eft == Allow {
				// set hit rule to first matched allow rule, maybe overridden by the deny part
				if result == Indeterminate {
					hitPolicyIndex = i
				}
				result = Allow
			} else if eft == Deny {
				result = Deny
				// set hit rule to the (first) matched deny rule
				hitPolicyIndex = i
				break
			}
		}
	case Priority, SubjectPriority:
		for i, eft := range effects {
			if matches[i] == 0 {
				continue
			}

			if eft != Indeterminate {
				if eft == Allow {
					result = Allow
				} else {
					result = Deny
				}
				hitPolicyIndex = i
				break
			}
		}
	default:
		return Deny, -1, errors.New("unsupported effect")
	}
	return result, hitPolicyIndex, nil
}
