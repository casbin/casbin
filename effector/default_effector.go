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

import (
	"errors"

	"github.com/casbin/casbin/v3/constant"
)

// DefaultEffector is default effector for Casbin.
type DefaultEffector struct {
}

// NewDefaultEffector is the constructor for DefaultEffector.
func NewDefaultEffector() *DefaultEffector {
	e := DefaultEffector{}
	return &e
}

// MergeEffects merges all matching results collected by the enforcer into a single decision.
func (e *DefaultEffector) MergeEffects(expr string, effects []Effect, matches []float64, policyIndex int, policyLength int) (Effect, int, error) {
	result := Indeterminate
	explainIndex := -1

	switch expr {
	case constant.AllowOverrideEffect:
		if matches[policyIndex] == 0 {
			break
		}
		// only check the current policyIndex (priority order: Allow > RateLimit)
		if effects[policyIndex] == Allow {
			result = Allow
			explainIndex = policyIndex
			break
		}
		if effects[policyIndex] == RateLimit {
			result = RateLimit
			explainIndex = policyIndex
			break
		}
	case constant.DenyOverrideEffect:
		// only check the current policyIndex
		if matches[policyIndex] != 0 && effects[policyIndex] == Deny {
			result = Deny
			explainIndex = policyIndex
			break
		}
		// if no deny rules are matched at last, check for allow or rate_limit
		if policyIndex == policyLength-1 {
			// Check all matched policies for allow first, then rate_limit (priority order: Allow > RateLimit)
			for i := range effects {
				if matches[i] == 0 {
					continue
				}
				if effects[i] == Allow {
					result = Allow
					explainIndex = i
					break
				}
			}
			// If no allow found, check for rate_limit
			if result == Indeterminate {
				for i := range effects {
					if matches[i] == 0 {
						continue
					}
					if effects[i] == RateLimit {
						result = RateLimit
						explainIndex = i
						break
					}
				}
			}
			// DenyOverride defaults to Allow if no deny rules matched (matches original behavior)
			if result == Indeterminate {
				result = Allow
			}
		}
	case constant.AllowAndDenyEffect:
		// short-circuit if matched deny rule
		if matches[policyIndex] != 0 && effects[policyIndex] == Deny {
			result = Deny
			// set hit rule to the (first) matched deny rule
			explainIndex = policyIndex
			break
		}

		// short-circuit some effects in the middle
		if policyIndex < policyLength-1 {
			// choose not to short-circuit
			return result, explainIndex, nil
		}
		// merge all effects at last (priority order: Allow > RateLimit)
		for i, eft := range effects {
			if matches[i] == 0 {
				continue
			}

			if eft == Allow {
				result = Allow
				// set hit rule to first matched allow rule
				explainIndex = i
				break
			}
			if eft == RateLimit {
				result = RateLimit
				// set hit rule to first matched rate_limit rule
				explainIndex = i
				break
			}
		}
	case constant.PriorityEffect, constant.SubjectPriorityEffect:
		// reverse merge, short-circuit may be earlier (priority order: Allow > RateLimit > Deny)
		for i := len(effects) - 1; i >= 0; i-- {
			if matches[i] == 0 {
				continue
			}

			if effects[i] != Indeterminate {
				if effects[i] == Allow {
					result = Allow
				} else if effects[i] == RateLimit {
					result = RateLimit
				} else {
					result = Deny
				}
				explainIndex = i
				break
			}
		}
	default:
		return Deny, -1, errors.New("unsupported effect")
	}

	return result, explainIndex, nil
}
