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

package main

import (
	"context"
	"errors"

	"github.com/casbin/casbin/v3"
	"github.com/casbin/casbin/v3/model"
	"github.com/casbin/casbin/v3/persist"
)

// MockTransactionalAdapter implements TransactionalAdapter interface for examples.
type MockTransactionalAdapter struct {
	Enforcer *casbin.Enforcer
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

// MockTransactionContext implements TransactionContext interface for examples.
type MockTransactionContext struct {
	adapter    *MockTransactionalAdapter
	committed  bool
	rolledBack bool
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
