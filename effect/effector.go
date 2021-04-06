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

package effect

import "errors"

// Effect is the result for a policy rule.
type Effect int

// Values for policy effect.
const (
	Allow Effect = iota
	Indeterminate
	Deny
)

// Effector is the interface for Casbin effectors.
type Effector interface {
	// IntermediateEffect returns a intermediate effect based on the matched effects of the enforcer
	IntermediateEffect(effects [3]int) Effect

	//FinalEffect returns the final effect based on the matched effects of the enforcer """
	FinalEffect(effects [3]int) Effect
}

// NewDefaultEffector is the constructor for DefaultEffector.
func NewEffector(expr string) (Effector, error) {
	if expr == "some(where (p_eft == allow))" {
		return AllowOverrideEffector{}, nil
	} else if expr == "!some(where (p_eft == deny))" {
		return DenyOverrideEffector{}, nil
	} else if expr == "some(where (p_eft == allow)) && !some(where (p_eft == deny))" {
		return AllowAndDenyEffector{}, nil
	} else if expr == "some(where (p_eft == allow)) || !some(where (p_eft == deny))" {
		return AllowOrDenyEffector{}, nil
	} else if expr == "priority(p_eft) || deny" {
		return PriorityEffector{}, nil
	} else {
		return nil, errors.New("unsupported effect")
	}
}

func EffectToBool(effect Effect) (bool, error) {
	if effect == Allow {
		return true, nil
	}
	if effect == Deny {
		return false, nil
	}
	return false, errors.New("effect can't be converted to boolean")
}
