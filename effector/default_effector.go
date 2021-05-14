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

// NewDefaultEffector is the constructor for DefaultEffector.
func NewDefaultEffector() *DefaultEffector {
	e := DefaultEffector{}
	return &e
}

// MergeEffects merges all matching results collected by the enforcer into a single decision.
func (e *DefaultEffector) MergeEffects(expr string, effects []Effect, matches []float64, policyIndex int, policyLength int) (Effect, int, error) {
	result := Indeterminate
	explainIndex := -1

	// short-circuit some effects in the middle
	if expr != "priority(p_eft) || deny" {
		if policyIndex < policyLength-1 {
			// choose not to short-circuit
			return result, explainIndex, nil
		}
	}

	// merge all effects at last
	if expr == "some(where (p_eft == allow))" {
		result = Indeterminate
		for i, eft := range effects {
			if matches[i] == 0 {
				continue
			}

			if eft == Allow {
				result = Allow
				explainIndex = i
				break
			}
		}
	} else if expr == "!some(where (p_eft == deny))" {
		// if no deny rules are matched, then allow
		result = Allow
		for i, eft := range effects {
			if matches[i] == 0 {
				continue
			}

			if eft == Deny {
				result = Deny
				explainIndex = i
				break
			}
		}
	} else if expr == "some(where (p_eft == allow)) && !some(where (p_eft == deny))" {
		result = Indeterminate
		for i, eft := range effects {
			if matches[i] == 0 {
				continue
			}

			if eft == Allow {
				// set hit rule to first matched allow rule, maybe overridden by the deny part
				if result == Indeterminate {
					explainIndex = i
				}
				result = Allow
			} else if eft == Deny {
				result = Deny
				// set hit rule to the (first) matched deny rule
				explainIndex = i
				break
			}
		}
	} else if expr == "priority(p_eft) || deny" {
		result = Indeterminate
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
				explainIndex = i
				break
			}
		}
	} else {
		return Deny, -1, errors.New("unsupported effect")
	}

	return result, explainIndex, nil
}
