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

	"github.com/casbin/casbin/v2/constant/policyEffect"
)

// DefaultEffector is default effector for Casbin.
type DefaultEffector struct {
}

// NewDefaultEffector is the constructor for DefaultEffector.
func NewDefaultEffector() *DefaultEffector {
	e := DefaultEffector{}
	return &e
}

func (e *DefaultEffector) TryEvaluate(expr string, effect Effect, match bool) (result Effect, isOver bool, isHit bool, err error) {
	result = Indeterminate
	isOver = false
	isHit = false

	switch expr {
	case policyEffect.AllowOverride:
		result = Deny
		if match && effect == Allow {
			result = Allow
			isOver = true
			isHit = true
		}
	case policyEffect.DenyOverride:
		result = Allow
		if match && effect == Deny {
			result = Deny
			isOver = true
			isHit = true
		}
	case policyEffect.AllowAndDeny:
		if !match {
			break
		}
		if effect == Allow {
			result = Allow
			isHit = true
		} else if effect == Deny {
			result = Deny
			isOver = true
			isHit = true
		}
	case policyEffect.Priority, policyEffect.SubjectPriority:
		if !match {
			break
		}
		if effect == Allow || effect == Deny {
			result = effect
			isOver = true
			isHit = true
		}
	case policyEffect.PriorityDenyOverride, policyEffect.SubjectPriorityDenyOverride:
		if !match {
			break
		}
		if effect == Allow {
			result = Allow
			isHit = true
		} else if effect == Deny {
			result = Deny
			isOver = true
			isHit = true
		}
	case policyEffect.PriorityAllowOverride, policyEffect.SubjectPriorityAllowOverride:
		if !match {
			break
		}
		if effect == Allow {
			result = Allow
			isOver = true
			isHit = true
		} else if effect == Deny {
			result = Deny
			isHit = true
		}
	default:
		return Deny, false, false, errors.New("unsupported effect")
	}

	return result, isOver, isHit, nil
}

// MergeEffects merges all matching results collected by the enforcer into a single decision.
func (e *DefaultEffector) MergeEffects(expr string, effects []Effect, matches []float64, policyIndex int, policyLength int) (Effect, int, error) {
	result := Indeterminate
	explainIndex := -1

	switch expr {
	case policyEffect.AllowOverride:
		if matches[policyIndex] == 0 {
			break
		}
		// only check the current policyIndex
		if effects[policyIndex] == Allow {
			result = Allow
			explainIndex = policyIndex
			break
		}
	case policyEffect.DenyOverride:
		// only check the current policyIndex
		if matches[policyIndex] != 0 && effects[policyIndex] == Deny {
			result = Deny
			explainIndex = policyIndex
			break
		}
		// if no deny rules are matched  at last, then allow
		if policyIndex == policyLength-1 {
			result = Allow
		}
	case policyEffect.AllowAndDeny:
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
		// merge all effects at last
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
		}
	case policyEffect.Priority, policyEffect.SubjectPriority:
		// reverse merge, short-circuit may be earlier
		for i := len(effects) - 1; i >= 0; i-- {
			if matches[i] == 0 {
				continue
			}

			if effects[i] != Indeterminate {
				if effects[i] == Allow {
					result = Allow
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
