// Copyright 2017 The casbin Authors. All Rights Reserved.
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

package casbin

import (
	"errors"
	"fmt"
)

// NewEnforcerSafe calls NewEnforcer in a safe way, returns error instead of causing panic.
func NewEnforcerSafe(params ...interface{}) (e *Enforcer, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprintf("%v", r))
			e = nil
		}
	}()

	e = NewEnforcer(params...)
	err = nil
	return
}

// LoadModelSafe calls LoadModel in a safe way, returns error instead of causing panic.
func (e *Enforcer) LoadModelSafe() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprintf("%v", r))
		}
	}()

	e.LoadModel()
	err = nil
	return
}

// EnforceSafe calls Enforce in a safe way, returns error instead of causing panic.
func (e *Enforcer) EnforceSafe(rvals ...interface{}) (result bool, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprintf("%v", r))
			result = false
		}
	}()

	result = e.Enforce(rvals...)
	err = nil
	return
}

func (e *Enforcer) AddPolicySafe(params ...interface{}) (result bool, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprintf("%v", r))
			result = false
		}
	}()

	result = e.AddNamedPolicy("p", params...)
	err = nil
	return
}

func (e *Enforcer) RemovePolicySafe(params ...interface{}) (result bool, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprintf("%v", r))
			result = false
		}
	}()

	result = e.RemoveNamedPolicy("p", params...)
	err = nil
	return
}

func (e *Enforcer) RemoveFilteredPolicySafe(fieldIndex int, fieldValues ...string) (result bool, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprintf("%v", r))
			result = false
		}
	}()

	result = e.RemoveFilteredNamedPolicy("p", fieldIndex, fieldValues...)
	err = nil
	return
}