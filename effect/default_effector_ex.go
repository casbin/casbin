// Copyright 2020 The casbin Authors. All Rights Reserved.
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

// DefaultEffectorEx is default effectorEx for Casbin.
type DefaultEffectorEx struct {
	DefaultEffector
}

// NewDefaultEffectorEx is the constructor for DefaultEffectorEx.
func NewDefaultEffectorEx() *DefaultEffectorEx {
	e := DefaultEffectorEx{}
	return &e
}

// IsHit return whether it should be returned by exforceEx()
func (e *DefaultEffectorEx) IsHit(expr string, eft Effect) (bool, error) {
	if expr == "some(where (p_eft == allow))" {
		if eft == Allow {
			return true, nil
		}
	} else if expr == "!some(where (p_eft == deny))" || expr == "some(where (p_eft == allow)) && !some(where (p_eft == deny))" {
		if eft == Deny {
			return true, nil
		}
	} else if expr == "priority(p_eft) || deny" {
		if eft != Indeterminate {
			return true, nil
		}
	} else {
		return false, errors.New("unsupported effect")
	}
	return false, nil
}