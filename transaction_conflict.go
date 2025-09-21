// Copyright 2025 The casbin Authors. All Rights Reserved.
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
	"fmt"

	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/persist"
)

// ConflictError represents a transaction conflict error.
type ConflictError struct {
	Operation persist.PolicyOperation
	Reason    string
}

func (e *ConflictError) Error() string {
	return fmt.Sprintf("transaction conflict: %s for operation %v", e.Reason, e.Operation)
}

// ConflictDetector detects conflicts between transaction operations and current model state.
type ConflictDetector struct {
	baseModel    model.Model               // Model snapshot when transaction started
	currentModel model.Model               // Current model state
	operations   []persist.PolicyOperation // Operations to be applied
}

// NewConflictDetector creates a new conflict detector instance.
func NewConflictDetector(baseModel, currentModel model.Model, operations []persist.PolicyOperation) *ConflictDetector {
	return &ConflictDetector{
		baseModel:    baseModel,
		currentModel: currentModel,
		operations:   operations,
	}
}

// DetectConflicts checks for conflicts between the transaction operations and current model state.
// Returns nil if no conflicts are found, otherwise returns a ConflictError describing the conflict.
func (cd *ConflictDetector) DetectConflicts() error {
	for _, op := range cd.operations {
		var err error
		switch op.Type {
		case persist.OperationAdd:
			// Add operations never conflict
			continue

		case persist.OperationRemove:
			err = cd.detectRemoveConflict(op)

		case persist.OperationUpdate:
			err = cd.detectUpdateConflict(op)
		}

		if err != nil {
			return err
		}
	}
	return nil
}

// detectRemoveConflict checks for conflicts in remove operations.
func (cd *ConflictDetector) detectRemoveConflict(op persist.PolicyOperation) error {
	for _, rule := range op.Rules {
		// Check if policy existed in base model
		baseHasPolicy, err := cd.baseModel.HasPolicy(op.Section, op.PolicyType, rule)
		if err != nil {
			return err
		}
		if !baseHasPolicy {
			continue // Policy didn't exist when transaction started
		}

		// Check if policy still exists in current model
		currentHasPolicy, err := cd.currentModel.HasPolicy(op.Section, op.PolicyType, rule)
		if err != nil {
			return err
		}
		if !currentHasPolicy {
			return &ConflictError{
				Operation: op,
				Reason:    "policy has been removed by another transaction",
			}
		}
	}
	return nil
}

// detectUpdateConflict checks for conflicts in update operations.
func (cd *ConflictDetector) detectUpdateConflict(op persist.PolicyOperation) error {
	for i, oldRule := range op.OldRules {
		if i >= len(op.Rules) {
			break
		}
		newRule := op.Rules[i]

		// Check if old policy still exists
		oldExists, err := cd.currentModel.HasPolicy(op.Section, op.PolicyType, oldRule)
		if err != nil {
			return err
		}
		if !oldExists {
			return &ConflictError{
				Operation: op,
				Reason:    "policy to be updated no longer exists",
			}
		}

		// Check if new policy already exists
		newExists, err := cd.currentModel.HasPolicy(op.Section, op.PolicyType, newRule)
		if err != nil {
			return err
		}
		if newExists {
			return &ConflictError{
				Operation: op,
				Reason:    "target policy already exists",
			}
		}
	}
	return nil
}
