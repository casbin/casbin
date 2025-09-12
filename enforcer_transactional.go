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
	"sync"

	"github.com/casbin/casbin/v2/persist"
)

// TransactionalEnforcer extends Enforcer with transaction support.
// It provides atomic policy operations through transactions.
type TransactionalEnforcer struct {
	*Enforcer              // Embedded enforcer for all standard functionality
	currentTx *Transaction // Current active transaction (nil if none)
	txMutex   sync.RWMutex // Protects transaction state
}

// NewTransactionalEnforcer creates a new TransactionalEnforcer.
// It accepts the same parameters as NewEnforcer.
func NewTransactionalEnforcer(params ...interface{}) (*TransactionalEnforcer, error) {
	enforcer, err := NewEnforcer(params...)
	if err != nil {
		return nil, err
	}

	return &TransactionalEnforcer{
		Enforcer: enforcer,
	}, nil
}

// BeginTransaction starts a new transaction.
// Returns an error if a transaction is already in progress or if the adapter doesn't support transactions.
func (te *TransactionalEnforcer) BeginTransaction(ctx context.Context) (*Transaction, error) {
	te.txMutex.Lock()
	defer te.txMutex.Unlock()

	if te.currentTx != nil {
		return nil, errors.New("transaction already in progress")
	}

	// Check if adapter supports transactions.
	txAdapter, ok := te.adapter.(persist.TransactionalAdapter)
	if !ok {
		return nil, errors.New("adapter does not support transactions")
	}

	// Start database transaction.
	txContext, err := txAdapter.BeginTransaction(ctx)
	if err != nil {
		return nil, err
	}

	// Create transaction buffer with current model snapshot.
	buffer := NewTransactionBuffer(te.model)

	tx := &Transaction{
		enforcer:  te,
		buffer:    buffer,
		txContext: txContext,
		ctx:       ctx,
	}

	te.currentTx = tx
	return tx, nil
}

// GetCurrentTransaction returns the current active transaction, or nil if none.
func (te *TransactionalEnforcer) GetCurrentTransaction() *Transaction {
	te.txMutex.RLock()
	defer te.txMutex.RUnlock()
	return te.currentTx
}

// IsInTransaction returns true if there is an active transaction.
func (te *TransactionalEnforcer) IsInTransaction() bool {
	te.txMutex.RLock()
	defer te.txMutex.RUnlock()
	return te.currentTx != nil
}

// clearTransaction clears the current transaction (called internally).
func (te *TransactionalEnforcer) clearTransaction() {
	te.txMutex.Lock()
	defer te.txMutex.Unlock()
	te.currentTx = nil
}

// WithTransaction executes a function within a transaction.
// If the function returns an error, the transaction is rolled back.
// Otherwise, it's committed automatically.
func (te *TransactionalEnforcer) WithTransaction(ctx context.Context, fn func(*Transaction) error) error {
	tx, err := te.BeginTransaction(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r)
		}
	}()

	err = fn(tx)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}
