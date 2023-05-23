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

package errors

import "errors"

// Global errors for rbac defined here
var (
	ErrNameNotFound                = errors.New("error: name does not exist")
	ErrDomainParameter             = errors.New("error: domain should be 1 parameter")
	ErrLinkNotFound                = errors.New("error: link between name1 and name2 does not exist")
	ErrUseDomainParameter          = errors.New("error: useDomain should be 1 parameter")
	ErrInvalidFieldValuesParameter = errors.New("fieldValues requires at least one parameter")

	// GetAllowedObjectConditions errors
	ErrObjCondition   = errors.New("need to meet the prefix required by the object condition")
	ErrEmptyCondition = errors.New("GetAllowedObjectConditions have an empty condition")
)
