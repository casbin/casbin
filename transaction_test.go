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
	"context"
	"errors"
	"testing"

	"github.com/casbin/casbin/v3/model"
	"github.com/casbin/casbin/v3/persist"
)

// MockTransactionalAdapter implements TransactionalAdapter interface for testing.
type MockTransactionalAdapter struct {
	Enforcer *Enforcer
}

// MockTransactionContext implements TransactionContext interface for testing.
type MockTransactionContext struct {
	adapter    *MockTransactionalAdapter
	committed  bool
	rolledBack bool
}

// NewMockTransactionalAdapter creates a new mock adapter.
func NewMockTransactionalAdapter() *MockTransactionalAdapter {
	return &MockTransactionalAdapter{}
}

// LoadPolicy implements Adapter interface.
func (a *MockTransactionalAdapter) LoadPolicy(model model.Model) error {
	return nil
}

// SavePolicy implements Adapter interface.
func (a *MockTransactionalAdapter) SavePolicy(model model.Model) error {
	return nil
}

// AddPolicy implements Adapter interface.
func (a *MockTransactionalAdapter) AddPolicy(sec string, ptype string, rule []string) error {
	return nil
}

// RemovePolicy implements Adapter interface.
func (a *MockTransactionalAdapter) RemovePolicy(sec string, ptype string, rule []string) error {
	return nil
}

// RemoveFilteredPolicy implements Adapter interface.
func (a *MockTransactionalAdapter) RemoveFilteredPolicy(sec string, ptype string, fieldIndex int, fieldValues ...string) error {
	return nil
}

// BeginTransaction implements TransactionalAdapter interface.
func (a *MockTransactionalAdapter) BeginTransaction(ctx context.Context) (persist.TransactionContext, error) {
	return &MockTransactionContext{adapter: a}, nil
}

// Commit implements TransactionContext interface.
func (tx *MockTransactionContext) Commit() error {
	if tx.committed || tx.rolledBack {
		return errors.New("transaction already finished")
	}
	tx.committed = true
	return nil
}

// Rollback implements TransactionContext interface.
func (tx *MockTransactionContext) Rollback() error {
	if tx.committed || tx.rolledBack {
		return errors.New("transaction already finished")
	}
	tx.rolledBack = true
	return nil
}

// GetAdapter implements TransactionContext interface.
func (tx *MockTransactionContext) GetAdapter() persist.Adapter {
	return tx.adapter
}

// Test basic transaction functionality.
func TestTransactionBasicOperations(t *testing.T) {
	adapter := NewMockTransactionalAdapter()
	e, err := NewTransactionalEnforcer("examples/rbac_model.conf", adapter)
	if err != nil {
		t.Fatalf("Failed to create transactional enforcer: %v", err)
	}
	adapter.Enforcer = e.Enforcer

	ctx := context.Background()

	// Begin transaction.
	tx, err := e.BeginTransaction(ctx)
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}

	// Add policies in transaction.
	ok, err := tx.AddPolicy("alice", "data1", "read")
	if !ok || err != nil {
		t.Fatalf("Failed to add policy in transaction: %v", err)
	}

	ok, err = tx.AddPolicy("bob", "data2", "write")
	if !ok || err != nil {
		t.Fatalf("Failed to add policy in transaction: %v", err)
	}

	// Commit transaction.
	if err := tx.Commit(); err != nil {
		t.Fatalf("Failed to commit transaction: %v", err)
	}

	// Verify transaction was committed.
	if !tx.IsCommitted() {
		t.Error("Transaction should be committed")
	}
}

// Test transaction rollback.
func TestTransactionRollback(t *testing.T) {
	adapter := NewMockTransactionalAdapter()
	e, err := NewTransactionalEnforcer("examples/rbac_model.conf", adapter)
	if err != nil {
		t.Fatalf("Failed to create transactional enforcer: %v", err)
	}
	adapter.Enforcer = e.Enforcer

	ctx := context.Background()

	// Begin transaction.
	tx, err := e.BeginTransaction(ctx)
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}

	// Add policy in transaction.
	ok, err := tx.AddPolicy("alice", "data1", "read")
	if !ok || err != nil {
		t.Fatalf("Failed to add policy in transaction: %v", err)
	}

	// Rollback transaction.
	if err := tx.Rollback(); err != nil {
		t.Fatalf("Failed to rollback transaction: %v", err)
	}

	// Verify transaction was rolled back.
	if !tx.IsRolledBack() {
		t.Error("Transaction should be rolled back")
	}
}

// Test concurrent transactions.
func TestConcurrentTransactions(t *testing.T) {
	adapter := NewMockTransactionalAdapter()
	e, err := NewTransactionalEnforcer("examples/rbac_model.conf", adapter)
	if err != nil {
		t.Fatalf("Failed to create transactional enforcer: %v", err)
	}
	adapter.Enforcer = e.Enforcer

	ctx := context.Background()

	// Start first transaction
	tx1, err := e.BeginTransaction(ctx)
	if err != nil {
		t.Fatalf("Failed to begin transaction 1: %v", err)
	}

	// Add policy in first transaction
	ok, err := tx1.AddPolicy("alice", "data1", "read")
	if !ok || err != nil {
		t.Fatalf("Failed to add policy in transaction 1: %v", err)
	}

	// Start second transaction
	tx2, err := e.BeginTransaction(ctx)
	if err != nil {
		t.Fatalf("Failed to begin transaction 2: %v", err)
	}

	// Add different policy in second transaction
	ok, err = tx2.AddPolicy("bob", "data2", "write")
	if !ok || err != nil {
		t.Fatalf("Failed to add policy in transaction 2: %v", err)
	}

	// Commit first transaction
	if err := tx1.Commit(); err != nil {
		t.Fatalf("Failed to commit transaction 1: %v", err)
	}

	// Commit second transaction
	if err := tx2.Commit(); err != nil {
		t.Fatalf("Failed to commit transaction 2: %v", err)
	}

	// Verify transactions were committed
	if !tx1.IsCommitted() {
		t.Error("Transaction 1 should be committed")
	}
	if !tx2.IsCommitted() {
		t.Error("Transaction 2 should be committed")
	}
}

// Test transaction conflicts.
func TestTransactionConflicts(t *testing.T) {
	adapter := NewMockTransactionalAdapter()
	e, err := NewTransactionalEnforcer("examples/rbac_model.conf", adapter)
	if err != nil {
		t.Fatalf("Failed to create transactional enforcer: %v", err)
	}
	adapter.Enforcer = e.Enforcer

	ctx := context.Background()

	// Test Case 1: Two transactions commit
	t.Run("TwoTransactionsCommit", func(t *testing.T) {
		tx1, _ := e.BeginTransaction(ctx)
		tx2, _ := e.BeginTransaction(ctx)

		// Commit both transactions
		if err := tx1.Commit(); err != nil {
			t.Fatalf("Failed to commit tx1: %v", err)
		}
		if err := tx2.Commit(); err != nil {
			t.Fatalf("Failed to commit tx2: %v", err)
		}

		// Verify both transactions were committed
		if !tx1.IsCommitted() {
			t.Error("Transaction 1 should be committed")
		}
		if !tx2.IsCommitted() {
			t.Error("Transaction 2 should be committed")
		}
	})

	// Test Case 2: Transaction rollback
	t.Run("TransactionRollback", func(t *testing.T) {
		tx, _ := e.BeginTransaction(ctx)

		// Rollback transaction
		if err := tx.Rollback(); err != nil {
			t.Fatalf("Failed to rollback transaction: %v", err)
		}

		// Verify transaction was rolled back
		if !tx.IsRolledBack() {
			t.Error("Transaction should be rolled back")
		}
	})

	// Test Case 3: Cannot commit after rollback
	t.Run("NoCommitAfterRollback", func(t *testing.T) {
		tx, _ := e.BeginTransaction(ctx)

		// Rollback transaction
		if err := tx.Rollback(); err != nil {
			t.Fatalf("Failed to rollback transaction: %v", err)
		}

		// Try to commit
		if err := tx.Commit(); err == nil {
			t.Error("Should not be able to commit after rollback")
		}
	})
}

// Test transaction buffer operations.
func TestTransactionBuffer(t *testing.T) {
	adapter := NewMockTransactionalAdapter()
	e, err := NewTransactionalEnforcer("examples/rbac_model.conf", adapter)
	if err != nil {
		t.Fatalf("Failed to create transactional enforcer: %v", err)
	}
	adapter.Enforcer = e.Enforcer

	ctx := context.Background()

	tx, err := e.BeginTransaction(ctx)
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}

	// Initially no operations.
	if tx.HasOperations() {
		t.Fatal("Transaction should have no operations initially")
	}

	if tx.OperationCount() != 0 {
		t.Fatal("Operation count should be 0 initially")
	}

	// Add some operations.
	tx.AddPolicy("alice", "data1", "read")
	tx.AddPolicy("bob", "data2", "write")

	if !tx.HasOperations() {
		t.Fatal("Transaction should have operations")
	}

	if tx.OperationCount() != 2 {
		t.Fatalf("Expected 2 operations, got %d", tx.OperationCount())
	}

	// Get buffered model.
	bufferedModel, err := tx.GetBufferedModel()
	if err != nil {
		t.Fatalf("Failed to get buffered model: %v", err)
	}

	// Check that buffered model contains the policies.
	hasPolicy, _ := bufferedModel.HasPolicy("p", "p", []string{"alice", "data1", "read"})
	if !hasPolicy {
		t.Fatal("Buffered model should contain the added policy")
	}

	tx.Rollback()
}

// externalTxContext wraps a MockTransactionalAdapter to simulate an externally-managed
// DB transaction (e.g. GORM). Commit and Rollback are intentional no-ops because the
// external system owns the transaction lifecycle.
type externalTxContext struct {
	adapter    *MockTransactionalAdapter
	committed  bool
	rolledBack bool
}

func (e *externalTxContext) Commit() error {
	e.committed = true
	return nil
}

func (e *externalTxContext) Rollback() error {
	e.rolledBack = true
	return nil
}

func (e *externalTxContext) GetAdapter() persist.Adapter {
	return e.adapter
}

// TestBeginTransactionWithContext verifies that Casbin operations are applied to the
// database adapter but the external transaction's Commit/Rollback are never called.
func TestBeginTransactionWithContext(t *testing.T) {
	adapter := NewMockTransactionalAdapter()
	e, err := NewTransactionalEnforcer("examples/rbac_model.conf", adapter)
	if err != nil {
		t.Fatalf("Failed to create transactional enforcer: %v", err)
	}
	adapter.Enforcer = e.Enforcer

	ctx := context.Background()
	extCtx := &externalTxContext{adapter: adapter}

	tx, err := e.BeginTransactionWithContext(ctx, extCtx)
	if err != nil {
		t.Fatalf("Failed to begin transaction with context: %v", err)
	}

	ok, err := tx.AddPolicy("alice", "data1", "read")
	if !ok || err != nil {
		t.Fatalf("Failed to add policy in external transaction: %v", err)
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("Failed to commit external transaction: %v", err)
	}

	// Casbin should NOT have called Commit on the external context.
	if extCtx.committed {
		t.Error("Casbin must not commit the external transaction context")
	}
	if extCtx.rolledBack {
		t.Error("Casbin must not rollback the external transaction context")
	}

	// The in-memory model should reflect the added policy.
	bufferedModel := e.GetModel()
	hasPolicy, _ := bufferedModel.HasPolicy("p", "p", []string{"alice", "data1", "read"})
	if !hasPolicy {
		t.Fatal("In-memory model should contain the added policy after commit")
	}
}

// TestBeginTransactionWithContextRollback verifies that rolling back an external
// transaction does not touch the external DB transaction.
func TestBeginTransactionWithContextRollback(t *testing.T) {
	adapter := NewMockTransactionalAdapter()
	e, err := NewTransactionalEnforcer("examples/rbac_model.conf", adapter)
	if err != nil {
		t.Fatalf("Failed to create transactional enforcer: %v", err)
	}
	adapter.Enforcer = e.Enforcer

	ctx := context.Background()
	extCtx := &externalTxContext{adapter: adapter}

	tx, err := e.BeginTransactionWithContext(ctx, extCtx)
	if err != nil {
		t.Fatalf("Failed to begin transaction with context: %v", err)
	}

	if _, err := tx.AddPolicy("alice", "data1", "read"); err != nil {
		t.Fatalf("Failed to add policy in external transaction: %v", err)
	}

	if err := tx.Rollback(); err != nil {
		t.Fatalf("Failed to rollback external transaction: %v", err)
	}

	// Casbin should NOT have called Rollback on the external context.
	if extCtx.rolledBack {
		t.Error("Casbin must not rollback the external transaction context")
	}
	if extCtx.committed {
		t.Error("Casbin must not commit the external transaction context")
	}
}

// TestWithExternalTransaction verifies the convenience wrapper applies operations
// and does not commit/rollback the external context.
func TestWithExternalTransaction(t *testing.T) {
	adapter := NewMockTransactionalAdapter()
	e, err := NewTransactionalEnforcer("examples/rbac_model.conf", adapter)
	if err != nil {
		t.Fatalf("Failed to create transactional enforcer: %v", err)
	}
	adapter.Enforcer = e.Enforcer

	ctx := context.Background()
	extCtx := &externalTxContext{adapter: adapter}

	err = e.WithExternalTransaction(ctx, extCtx, func(tx *Transaction) error {
		_, err := tx.AddPolicy("bob", "data2", "write")
		return err
	})
	if err != nil {
		t.Fatalf("WithExternalTransaction failed: %v", err)
	}

	// External context must not be committed/rolled back by Casbin.
	if extCtx.committed {
		t.Error("Casbin must not commit the external transaction context")
	}
	if extCtx.rolledBack {
		t.Error("Casbin must not rollback the external transaction context")
	}

	// In-memory model should reflect the change.
	hasPolicy, _ := e.GetModel().HasPolicy("p", "p", []string{"bob", "data2", "write"})
	if !hasPolicy {
		t.Fatal("In-memory model should contain the added policy after WithExternalTransaction")
	}
}

// TestWithExternalTransactionRollbackOnError verifies that when fn returns an error,
// the external context is not rolled back by Casbin.
func TestWithExternalTransactionRollbackOnError(t *testing.T) {
	adapter := NewMockTransactionalAdapter()
	e, err := NewTransactionalEnforcer("examples/rbac_model.conf", adapter)
	if err != nil {
		t.Fatalf("Failed to create transactional enforcer: %v", err)
	}
	adapter.Enforcer = e.Enforcer

	ctx := context.Background()
	extCtx := &externalTxContext{adapter: adapter}

	fnErr := errors.New("business logic failure")
	err = e.WithExternalTransaction(ctx, extCtx, func(tx *Transaction) error {
		if _, addErr := tx.AddPolicy("charlie", "data3", "read"); addErr != nil {
			return addErr
		}
		return fnErr
	})
	if !errors.Is(err, fnErr) {
		t.Fatalf("Expected fnErr, got %v", err)
	}

	// Casbin must not touch the external transaction.
	if extCtx.rolledBack {
		t.Error("Casbin must not rollback the external transaction context on error")
	}
	if extCtx.committed {
		t.Error("Casbin must not commit the external transaction context on error")
	}
}
