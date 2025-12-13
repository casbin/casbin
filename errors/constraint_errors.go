// Copyright 2024 The casbin Authors. All Rights Reserved.
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

import (
	"errors"
	"fmt"
)

// Global errors for constraints defined here.
var (
	ErrConstraintViolation         = errors.New("constraint violation")
	ErrConstraintParsingError      = errors.New("constraint parsing error")
	ErrConstraintRequiresRBAC      = errors.New("constraints require RBAC to be enabled (role_definition section must exist)")
	ErrInvalidConstraintDefinition = errors.New("invalid constraint definition")
)

// ConstraintViolationError represents a specific constraint violation.
type ConstraintViolationError struct {
	ConstraintName string
	Message        string
}

func (e *ConstraintViolationError) Error() string {
	return fmt.Sprintf("constraint violation [%s]: %s", e.ConstraintName, e.Message)
}

// NewConstraintViolationError creates a new constraint violation error.
func NewConstraintViolationError(constraintName, message string) error {
	return &ConstraintViolationError{
		ConstraintName: constraintName,
		Message:        message,
	}
}
