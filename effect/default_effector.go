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

// DefaultEffector is default effector for Casbin.
type DefaultEffector struct {
}

// NewDefaultEffector is the constructor for DefaultEffector.
func NewDefaultEffector() *DefaultEffector {
	e := DefaultEffector{}
	return &e
}

func (e *DefaultEffector) NewStream(expr string, cap int) (DefaultEffectorStream) {
	if !(cap>0) {
		panic("cap should be greater than 0")
	}

	var res bool
	if expr=="some(where (p_eft == allow))" || expr=="some(where (p_eft == allow)) && !some(where (p_eft == deny))" || expr=="priority(p_eft) || deny" {
		res = false
	} else if expr=="!some(where (p_eft == deny))" {
		res = true
	} else {
		panic("unsupported effect: " + expr)
	}

	des := DefaultEffectorStream{}
	des.done = false
	des.res = res
	des.expr = expr
	des.cap = cap
	des.idx = 0
	des.expl = make([]int, 0)

	return des
}