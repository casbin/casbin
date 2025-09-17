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

	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/persist"
)

// MockTransactionalAdapter is a mock adapter that implements TransactionalAdapter for testing.
type MockTransactionalAdapter struct {
	policies map[string]map[string][][]string // section -> ptype -> rules
	tx       *MockTransactionContext
}

// MockTransactionContext is a mock transaction context for testing.
type MockTransactionContext struct {
	adapter    *MockTransactionalAdapter
	committed  bool
	rolledBack bool
}

// NewMockTransactionalAdapter creates a new mock transactional adapter.
func NewMockTransactionalAdapter() *MockTransactionalAdapter {
	return &MockTransactionalAdapter{
		policies: make(map[string]map[string][][]string),
	}
}

// LoadPolicy loads policy from the mock storage.
func (a *MockTransactionalAdapter) LoadPolicy(model model.Model) error {
	// Load policies from mock storage.
	for section, ptypes := range a.policies {
		for ptype, rules := range ptypes {
			for _, rule := range rules {
				if err := model.AddPolicy(section, ptype, rule); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// SavePolicy saves policy to the mock storage.
func (a *MockTransactionalAdapter) SavePolicy(model model.Model) error {
	a.policies = make(map[string]map[string][][]string)

	// Save p policies.
	if pSection, ok := model["p"]; ok {
		for ptype, ast := range pSection {
			if a.policies["p"] == nil {
				a.policies["p"] = make(map[string][][]string)
			}
			a.policies["p"][ptype] = ast.Policy
		}
	}

	// Save g policies.
	if gSection, ok := model["g"]; ok {
		for ptype, ast := range gSection {
			if a.policies["g"] == nil {
				a.policies["g"] = make(map[string][][]string)
			}
			a.policies["g"][ptype] = ast.Policy
		}
	}

	return nil
}

// AddPolicy adds a policy rule.
func (a *MockTransactionalAdapter) AddPolicy(sec string, ptype string, rule []string) error {
	if a.policies[sec] == nil {
		a.policies[sec] = make(map[string][][]string)
	}
	a.policies[sec][ptype] = append(a.policies[sec][ptype], rule)
	return nil
}

// RemovePolicy removes a policy rule.
func (a *MockTransactionalAdapter) RemovePolicy(sec string, ptype string, rule []string) error {
	if a.policies[sec] == nil || a.policies[sec][ptype] == nil {
		return nil
	}

	for i, existingRule := range a.policies[sec][ptype] {
		if len(existingRule) == len(rule) {
			match := true
			for j, v := range rule {
				if existingRule[j] != v {
					match = false
					break
				}
			}
			if match {
				a.policies[sec][ptype] = append(a.policies[sec][ptype][:i], a.policies[sec][ptype][i+1:]...)
				break
			}
		}
	}
	return nil
}

// RemoveFilteredPolicy removes policy rules that match the filter.
func (a *MockTransactionalAdapter) RemoveFilteredPolicy(sec string, ptype string, fieldIndex int, fieldValues ...string) error {
	// Simple implementation for testing.
	return nil
}

// BeginTransaction starts a transaction.
func (a *MockTransactionalAdapter) BeginTransaction(ctx context.Context) (persist.TransactionContext, error) {
	a.tx = &MockTransactionContext{
		adapter: a,
	}
	return a.tx, nil
}

// Commit commits the transaction.
func (tx *MockTransactionContext) Commit() error {
	tx.committed = true
	return nil
}

// Rollback rolls back the transaction.
func (tx *MockTransactionContext) Rollback() error {
	tx.rolledBack = true
	return nil
}

// GetAdapter returns the adapter within the transaction.
func (tx *MockTransactionContext) GetAdapter() persist.Adapter {
	return tx.adapter
}

// Test basic transaction functionality.
func TestTransactionBasicOperations(t *testing.T) {
	adapter := NewMockTransactionalAdapter()

	// Create transactional enforcer.
	e, err := NewTransactionalEnforcer("examples/rbac_model.conf", adapter)
	if err != nil {
		t.Fatalf("Failed to create transactional enforcer: %v", err)
	}

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

	// Check that policies are not yet in the enforcer.
	if has, _ := e.HasPolicy("alice", "data1", "read"); has {
		t.Fatal("Policy should not be in enforcer before commit")
	}

	// Commit transaction.
	if err := tx.Commit(); err != nil {
		t.Fatalf("Failed to commit transaction: %v", err)
	}

	// Check that policies are now in the enforcer.
	if has, _ := e.HasPolicy("alice", "data1", "read"); !has {
		t.Fatal("Policy should be in enforcer after commit")
	}

	if has, _ := e.HasPolicy("bob", "data2", "write"); !has {
		t.Fatal("Policy should be in enforcer after commit")
	}
}

// Test transaction rollback.
func TestTransactionRollback(t *testing.T) {
	adapter := NewMockTransactionalAdapter()

	e, err := NewTransactionalEnforcer("examples/rbac_model.conf", adapter)
	if err != nil {
		t.Fatalf("Failed to create transactional enforcer: %v", err)
	}

	// Add initial policy.
	e.AddPolicy("alice", "data1", "read")

	ctx := context.Background()

	// Begin transaction.
	tx, err := e.BeginTransaction(ctx)
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}

	// Add policy in transaction.
	ok, err := tx.AddPolicy("bob", "data2", "write")
	if !ok || err != nil {
		t.Fatalf("Failed to add policy in transaction: %v", err)
	}

	// Remove existing policy in transaction.
	ok, err = tx.RemovePolicy("alice", "data1", "read")
	if !ok || err != nil {
		t.Fatalf("Failed to remove policy in transaction: %v", err)
	}

	// Rollback transaction.
	if err := tx.Rollback(); err != nil {
		t.Fatalf("Failed to rollback transaction: %v", err)
	}

	// Check that original state is preserved.
	if has, _ := e.HasPolicy("alice", "data1", "read"); !has {
		t.Fatal("Original policy should still exist after rollback")
	}

	if has, _ := e.HasPolicy("bob", "data2", "write"); has {
		t.Fatal("New policy should not exist after rollback")
	}
}

// Test multiple transactions.
func TestMultipleTransactions(t *testing.T) {
	adapter := NewMockTransactionalAdapter()

	e, err := NewTransactionalEnforcer("examples/rbac_model.conf", adapter)
	if err != nil {
		t.Fatalf("Failed to create transactional enforcer: %v", err)
	}

	ctx := context.Background()

	// First transaction.
	tx1, err := e.BeginTransaction(ctx)
	if err != nil {
		t.Fatalf("Failed to begin first transaction: %v", err)
	}

	tx1.AddPolicy("alice", "data1", "read")

	if commitErr := tx1.Commit(); commitErr != nil {
		t.Fatalf("Failed to commit first transaction: %v", commitErr)
	}

	// Second transaction.
	tx2, err := e.BeginTransaction(ctx)
	if err != nil {
		t.Fatalf("Failed to begin second transaction: %v", err)
	}

	tx2.AddPolicy("bob", "data2", "write")

	if err := tx2.Commit(); err != nil {
		t.Fatalf("Failed to commit second transaction: %v", err)
	}

	// Check both policies exist.
	if has, _ := e.HasPolicy("alice", "data1", "read"); !has {
		t.Fatal("First policy should exist")
	}

	if has, _ := e.HasPolicy("bob", "data2", "write"); !has {
		t.Fatal("Second policy should exist")
	}
}

// Test WithTransaction helper method.
func TestWithTransaction(t *testing.T) {
	adapter := NewMockTransactionalAdapter()

	e, err := NewTransactionalEnforcer("examples/rbac_model.conf", adapter)
	if err != nil {
		t.Fatalf("Failed to create transactional enforcer: %v", err)
	}

	ctx := context.Background()

	// Test successful transaction.
	err = e.WithTransaction(ctx, func(tx *Transaction) error {
		tx.AddPolicy("alice", "data1", "read")
		tx.AddPolicy("bob", "data2", "write")
		return nil
	})

	if err != nil {
		t.Fatalf("WithTransaction failed: %v", err)
	}

	// Check policies exist.
	if has, _ := e.HasPolicy("alice", "data1", "read"); !has {
		t.Fatal("Policy should exist after successful WithTransaction")
	}

	// Test failed transaction (should rollback).
	err = e.WithTransaction(ctx, func(tx *Transaction) error {
		tx.AddPolicy("charlie", "data3", "read")
		return errors.New("simulated error")
	})

	if err == nil {
		t.Fatal("WithTransaction should have returned an error")
	}

	// Check that policy was not added due to rollback.
	if has, _ := e.HasPolicy("charlie", "data3", "read"); has {
		t.Fatal("Policy should not exist after failed WithTransaction")
	}
}

// Test transaction buffer operations.
func TestTransactionBuffer(t *testing.T) {
	adapter := NewMockTransactionalAdapter()

	e, err := NewTransactionalEnforcer("examples/rbac_model.conf", adapter)
	if err != nil {
		t.Fatalf("Failed to create transactional enforcer: %v", err)
	}

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
