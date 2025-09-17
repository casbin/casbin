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
	"errors"

	"github.com/casbin/casbin/v2/persist"
)

// Commit commits the transaction using a two-phase commit protocol.
// Phase 1: Apply all operations to the database
// Phase 2: Apply changes to the in-memory model and rebuild role links.
func (tx *Transaction) Commit() error {
	tx.mutex.Lock()
	defer tx.mutex.Unlock()

	if tx.committed {
		return errors.New("transaction already committed")
	}
	if tx.rolledBack {
		return errors.New("transaction already rolled back")
	}

	// If no operations, just commit the database transaction and clear state.
	if !tx.buffer.HasOperations() {
		if err := tx.txContext.Commit(); err != nil {
			return err
		}
		tx.committed = true
		tx.enforcer.clearTransaction()
		return nil
	}

	// Phase 1: Apply all buffered operations to the database
	if err := tx.applyOperationsToDatabase(); err != nil {
		// Rollback database transaction on failure.
		_ = tx.txContext.Rollback()
		return err
	}

	// Commit database transaction.
	if err := tx.txContext.Commit(); err != nil {
		return err
	}

	// Phase 2: Apply changes to the in-memory model
	if err := tx.applyOperationsToModel(); err != nil {
		// At this point, database is committed but model update failed.
		// This is a critical error that should not happen in normal circumstances.
		return errors.New("critical error: database committed but model update failed: " + err.Error())
	}

	tx.committed = true
	tx.enforcer.clearTransaction()

	return nil
}

// Rollback rolls back the transaction.
// This will rollback the database transaction and clear the transaction state.
func (tx *Transaction) Rollback() error {
	tx.mutex.Lock()
	defer tx.mutex.Unlock()

	if tx.committed {
		return errors.New("transaction already committed")
	}
	if tx.rolledBack {
		return errors.New("transaction already rolled back")
	}

	// Rollback database transaction.
	if err := tx.txContext.Rollback(); err != nil {
		return err
	}

	tx.rolledBack = true
	tx.enforcer.clearTransaction()

	return nil
}

// applyOperationsToDatabase applies all buffered operations to the database.
func (tx *Transaction) applyOperationsToDatabase() error {
	operations := tx.buffer.GetOperations()
	txAdapter := tx.txContext.GetAdapter()

	for _, op := range operations {
		switch op.Type {
		case persist.OperationAdd:
			if err := tx.applyAddOperationToDatabase(txAdapter, op); err != nil {
				return err
			}
		case persist.OperationRemove:
			if err := tx.applyRemoveOperationToDatabase(txAdapter, op); err != nil {
				return err
			}
		case persist.OperationUpdate:
			if err := tx.applyUpdateOperationToDatabase(txAdapter, op); err != nil {
				return err
			}
		}
	}

	return nil
}

// applyAddOperationToDatabase applies an add operation to the database.
func (tx *Transaction) applyAddOperationToDatabase(adapter persist.Adapter, op persist.PolicyOperation) error {
	if batchAdapter, ok := adapter.(persist.BatchAdapter); ok {
		// Use batch operation if available.
		return batchAdapter.AddPolicies(op.Section, op.PolicyType, op.Rules)
	} else {
		// Fall back to individual operations.
		for _, rule := range op.Rules {
			if err := adapter.AddPolicy(op.Section, op.PolicyType, rule); err != nil {
				return err
			}
		}
	}
	return nil
}

// applyRemoveOperationToDatabase applies a remove operation to the database.
func (tx *Transaction) applyRemoveOperationToDatabase(adapter persist.Adapter, op persist.PolicyOperation) error {
	if batchAdapter, ok := adapter.(persist.BatchAdapter); ok {
		// Use batch operation if available.
		return batchAdapter.RemovePolicies(op.Section, op.PolicyType, op.Rules)
	} else {
		// Fall back to individual operations.
		for _, rule := range op.Rules {
			if err := adapter.RemovePolicy(op.Section, op.PolicyType, rule); err != nil {
				return err
			}
		}
	}
	return nil
}

// applyUpdateOperationToDatabase applies an update operation to the database.
func (tx *Transaction) applyUpdateOperationToDatabase(adapter persist.Adapter, op persist.PolicyOperation) error {
	if updateAdapter, ok := adapter.(persist.UpdatableAdapter); ok {
		// Use update operation if available.
		return updateAdapter.UpdatePolicies(op.Section, op.PolicyType, op.OldRules, op.Rules)
	}

	// Fall back to remove + add.
	for i, oldRule := range op.OldRules {
		if i >= len(op.Rules) {
			continue
		}

		if err := adapter.RemovePolicy(op.Section, op.PolicyType, oldRule); err != nil {
			return err
		}
		if err := adapter.AddPolicy(op.Section, op.PolicyType, op.Rules[i]); err != nil {
			return err
		}
	}
	return nil
}

// applyOperationsToModel applies all buffered operations to the in-memory model.
func (tx *Transaction) applyOperationsToModel() error {
	// Create new model with all operations applied.
	newModel, err := tx.buffer.ApplyOperationsToModel(tx.buffer.GetModelSnapshot())
	if err != nil {
		return err
	}

	// Replace the enforcer's model.
	tx.enforcer.model = newModel
	tx.enforcer.invalidateMatcherMap()

	// Rebuild role links if necessary.
	if tx.enforcer.autoBuildRoleLinks {
		// Check if any operations involved grouping policies.
		operations := tx.buffer.GetOperations()
		needRoleRebuild := false

		for _, op := range operations {
			if op.Section == "g" {
				needRoleRebuild = true
				break
			}
		}

		if needRoleRebuild {
			if err := tx.enforcer.BuildRoleLinks(); err != nil {
				return err
			}
		}
	}

	return nil
}

// IsCommitted returns true if the transaction has been committed.
func (tx *Transaction) IsCommitted() bool {
	tx.mutex.RLock()
	defer tx.mutex.RUnlock()
	return tx.committed
}

// IsRolledBack returns true if the transaction has been rolled back.
func (tx *Transaction) IsRolledBack() bool {
	tx.mutex.RLock()
	defer tx.mutex.RUnlock()
	return tx.rolledBack
}

// IsActive returns true if the transaction is still active (not committed or rolled back).
func (tx *Transaction) IsActive() bool {
	tx.mutex.RLock()
	defer tx.mutex.RUnlock()
	return !tx.committed && !tx.rolledBack
}
