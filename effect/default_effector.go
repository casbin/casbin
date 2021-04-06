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

type AllowOverrideEffector struct {
}

//IntermediateEffect returns a intermediate effect based on the matched effects of the enforcer """
func (e AllowOverrideEffector) IntermediateEffect(effects [3]int) Effect {
	if effects[Allow] != 0 {
		return Allow
	}
	return Indeterminate
}

//FinalEffect returns the final effect based on the matched effects of the enforcer """
func (e AllowOverrideEffector) FinalEffect(effects [3]int) Effect {
	if effects[Allow] != 0 {
		return Allow
	}
	return Deny
}

type DenyOverrideEffector struct {
}

//IntermediateEffect returns a intermediate effect based on the matched effects of the enforcer """
func (e DenyOverrideEffector) IntermediateEffect(effects [3]int) Effect {
	if effects[Deny] != 0 {
		return Deny
	}
	return Indeterminate
}

//FinalEffect returns the final effect based on the matched effects of the enforcer """
func (e DenyOverrideEffector) FinalEffect(effects [3]int) Effect {
	if effects[Deny] != 0 {
		return Deny
	}
	return Allow
}

type AllowAndDenyEffector struct {
}

//IntermediateEffect returns a intermediate effect based on the matched effects of the enforcer """
func (e AllowAndDenyEffector) IntermediateEffect(effects [3]int) Effect {
	if effects[Deny] != 0 {
		return Deny
	}
	return Indeterminate
}

//FinalEffect returns the final effect based on the matched effects of the enforcer """
func (e AllowAndDenyEffector) FinalEffect(effects [3]int) Effect {
	if effects[Deny] != 0 || effects[Allow] == 0 {
		return Deny
	}
	return Allow
}

type AllowOrDenyEffector struct {
}

//IntermediateEffect returns a intermediate effect based on the matched effects of the enforcer """
func (e AllowOrDenyEffector) IntermediateEffect(effects [3]int) Effect {
	return Indeterminate
}

//FinalEffect returns the final effect based on the matched effects of the enforcer """
func (e AllowOrDenyEffector) FinalEffect(effects [3]int) Effect {
	if effects[Allow] != 0 || effects[Deny] == 0 {
		return Allow
	}
	return Deny
}

type PriorityEffector struct {
}

//IntermediateEffect returns a intermediate effect based on the matched effects of the enforcer """
func (e PriorityEffector) IntermediateEffect(effects [3]int) Effect {
	if effects[Allow] != 0 {
		return Allow
	} else if effects[Deny] != 0 {
		return Deny
	}
	return Indeterminate
}

//FinalEffect returns the final effect based on the matched effects of the enforcer """
func (e PriorityEffector) FinalEffect(effects [3]int) Effect {
	if effects[Allow] != 0 {
		return Allow
	} else if effects[Deny] != 0 {
		return Deny
	}
	return Deny
}
