// Copyright 2025 The casbin Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License").
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software.
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package casbin

import (
	"context"
	"errors"
	"sync"

	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/persist"
)

// Transaction represents a Casbin transaction.
// It provides methods to perform policy operations within a transaction.
// and commit or rollback all changes atomically.
type Transaction struct {
	enforcer   *TransactionalEnforcer     // Reference to the transactional enforcer.
	buffer     *TransactionBuffer         // Buffer for policy operations.
	txContext  persist.TransactionContext // Database transaction context.
	ctx        context.Context            // Context for the transaction.
	committed  bool                       // Whether the transaction has been committed.
	rolledBack bool                       // Whether the transaction has been rolled back.
	mutex      sync.RWMutex               // Protects transaction state.
}

// AddPolicy adds a policy within the transaction.
// The policy is buffered and will be applied when the transaction is committed.
func (tx *Transaction) AddPolicy(params ...interface{}) (bool, error) {
	return tx.AddNamedPolicy("p", params...)
}

// buildRuleFromParams converts parameters to a rule slice.
func (tx *Transaction) buildRuleFromParams(params ...interface{}) []string {
	if len(params) == 1 {
		if strSlice, ok := params[0].([]string); ok {
			rule := make([]string, 0, len(strSlice))
			rule = append(rule, strSlice...)
			return rule
		}
	}

	rule := make([]string, 0, len(params))
	for _, param := range params {
		rule = append(rule, param.(string))
	}
	return rule
}

// checkTransactionStatus checks if the transaction is active.
func (tx *Transaction) checkTransactionStatus() error {
	if tx.committed || tx.rolledBack {
		return errors.New("transaction is not active")
	}
	return nil
}

// AddNamedPolicy adds a named policy within the transaction.
// The policy is buffered and will be applied when the transaction is committed.
func (tx *Transaction) AddNamedPolicy(ptype string, params ...interface{}) (bool, error) {
	tx.mutex.Lock()
	defer tx.mutex.Unlock()

	if err := tx.checkTransactionStatus(); err != nil {
		return false, err
	}

	rule := tx.buildRuleFromParams(params...)

	// Check if policy already exists in the buffered model.
	bufferedModel, err := tx.buffer.ApplyOperationsToModel(tx.buffer.GetModelSnapshot())
	if err != nil {
		return false, err
	}

	hasPolicy, err := bufferedModel.HasPolicy("p", ptype, rule)
	if hasPolicy || err != nil {
		return false, err
	}

	// Add operation to buffer.
	op := persist.PolicyOperation{
		Type:       persist.OperationAdd,
		Section:    "p",
		PolicyType: ptype,
		Rules:      [][]string{rule},
	}
	tx.buffer.AddOperation(op)

	return true, nil
}

// AddPolicies adds multiple policies within the transaction.
func (tx *Transaction) AddPolicies(rules [][]string) (bool, error) {
	return tx.AddNamedPolicies("p", rules)
}

// AddNamedPolicies adds multiple named policies within the transaction.
func (tx *Transaction) AddNamedPolicies(ptype string, rules [][]string) (bool, error) {
	tx.mutex.Lock()
	defer tx.mutex.Unlock()

	if tx.committed || tx.rolledBack {
		return false, errors.New("transaction is not active")
	}

	if len(rules) == 0 {
		return false, nil
	}

	// Check if any policies already exist in the buffered model.
	bufferedModel, err := tx.buffer.ApplyOperationsToModel(tx.buffer.GetModelSnapshot())
	if err != nil {
		return false, err
	}

	var validRules [][]string
	for _, rule := range rules {
		hasPolicy, err := bufferedModel.HasPolicy("p", ptype, rule)
		if err != nil {
			return false, err
		}
		if !hasPolicy {
			validRules = append(validRules, rule)
		}
	}

	if len(validRules) == 0 {
		return false, nil
	}

	// Add operation to buffer.
	op := persist.PolicyOperation{
		Type:       persist.OperationAdd,
		Section:    "p",
		PolicyType: ptype,
		Rules:      validRules,
	}
	tx.buffer.AddOperation(op)

	return true, nil
}

// RemovePolicy removes a policy within the transaction.
func (tx *Transaction) RemovePolicy(params ...interface{}) (bool, error) {
	return tx.RemoveNamedPolicy("p", params...)
}

// RemoveNamedPolicy removes a named policy within the transaction.
func (tx *Transaction) RemoveNamedPolicy(ptype string, params ...interface{}) (bool, error) {
	tx.mutex.Lock()
	defer tx.mutex.Unlock()

	if err := tx.checkTransactionStatus(); err != nil {
		return false, err
	}

	rule := tx.buildRuleFromParams(params...)

	// Check if policy exists in the buffered model.
	bufferedModel, err := tx.buffer.ApplyOperationsToModel(tx.buffer.GetModelSnapshot())
	if err != nil {
		return false, err
	}

	hasPolicy, err := bufferedModel.HasPolicy("p", ptype, rule)
	if !hasPolicy || err != nil {
		return false, err
	}

	// Add operation to buffer.
	op := persist.PolicyOperation{
		Type:       persist.OperationRemove,
		Section:    "p",
		PolicyType: ptype,
		Rules:      [][]string{rule},
	}
	tx.buffer.AddOperation(op)

	return true, nil
}

// RemovePolicies removes multiple policies within the transaction.
func (tx *Transaction) RemovePolicies(rules [][]string) (bool, error) {
	return tx.RemoveNamedPolicies("p", rules)
}

// RemoveNamedPolicies removes multiple named policies within the transaction.
func (tx *Transaction) RemoveNamedPolicies(ptype string, rules [][]string) (bool, error) {
	tx.mutex.Lock()
	defer tx.mutex.Unlock()

	if tx.committed || tx.rolledBack {
		return false, errors.New("transaction is not active")
	}

	if len(rules) == 0 {
		return false, nil
	}

	// Check if policies exist in the buffered model.
	bufferedModel, err := tx.buffer.ApplyOperationsToModel(tx.buffer.GetModelSnapshot())
	if err != nil {
		return false, err
	}

	var validRules [][]string
	for _, rule := range rules {
		hasPolicy, err := bufferedModel.HasPolicy("p", ptype, rule)
		if err != nil {
			return false, err
		}
		if hasPolicy {
			validRules = append(validRules, rule)
		}
	}

	if len(validRules) == 0 {
		return false, nil
	}

	// Add operation to buffer.
	op := persist.PolicyOperation{
		Type:       persist.OperationRemove,
		Section:    "p",
		PolicyType: ptype,
		Rules:      validRules,
	}
	tx.buffer.AddOperation(op)

	return true, nil
}

// UpdatePolicy updates a policy within the transaction.
func (tx *Transaction) UpdatePolicy(oldPolicy []string, newPolicy []string) (bool, error) {
	return tx.UpdateNamedPolicy("p", oldPolicy, newPolicy)
}

// UpdateNamedPolicy updates a named policy within the transaction.
func (tx *Transaction) UpdateNamedPolicy(ptype string, oldPolicy []string, newPolicy []string) (bool, error) {
	tx.mutex.Lock()
	defer tx.mutex.Unlock()

	if tx.committed || tx.rolledBack {
		return false, errors.New("transaction is not active")
	}

	// Check if old policy exists and new policy doesn't exist.
	bufferedModel, err := tx.buffer.ApplyOperationsToModel(tx.buffer.GetModelSnapshot())
	if err != nil {
		return false, err
	}

	hasOldPolicy, err := bufferedModel.HasPolicy("p", ptype, oldPolicy)
	if err != nil {
		return false, err
	}
	if !hasOldPolicy {
		return false, nil
	}

	hasNewPolicy, errNew := bufferedModel.HasPolicy("p", ptype, newPolicy)
	if errNew != nil {
		return false, errNew
	}
	if hasNewPolicy {
		return false, nil
	}

	// Add operation to buffer.
	op := persist.PolicyOperation{
		Type:       persist.OperationUpdate,
		Section:    "p",
		PolicyType: ptype,
		Rules:      [][]string{newPolicy},
		OldRules:   [][]string{oldPolicy},
	}
	tx.buffer.AddOperation(op)

	return true, nil
}

// AddGroupingPolicy adds a grouping policy within the transaction.
func (tx *Transaction) AddGroupingPolicy(params ...interface{}) (bool, error) {
	return tx.AddNamedGroupingPolicy("g", params...)
}

// AddNamedGroupingPolicy adds a named grouping policy within the transaction.
func (tx *Transaction) AddNamedGroupingPolicy(ptype string, params ...interface{}) (bool, error) {
	tx.mutex.Lock()
	defer tx.mutex.Unlock()

	if err := tx.checkTransactionStatus(); err != nil {
		return false, err
	}

	rule := tx.buildRuleFromParams(params...)

	// Check if grouping policy already exists in the buffered model.
	bufferedModel, err := tx.buffer.ApplyOperationsToModel(tx.buffer.GetModelSnapshot())
	if err != nil {
		return false, err
	}

	hasPolicy, err := bufferedModel.HasPolicy("g", ptype, rule)
	if hasPolicy || err != nil {
		return false, err
	}

	// Add operation to buffer.
	op := persist.PolicyOperation{
		Type:       persist.OperationAdd,
		Section:    "g",
		PolicyType: ptype,
		Rules:      [][]string{rule},
	}
	tx.buffer.AddOperation(op)

	return true, nil
}

// RemoveGroupingPolicy removes a grouping policy within the transaction.
func (tx *Transaction) RemoveGroupingPolicy(params ...interface{}) (bool, error) {
	return tx.RemoveNamedGroupingPolicy("g", params...)
}

// RemoveNamedGroupingPolicy removes a named grouping policy within the transaction.
func (tx *Transaction) RemoveNamedGroupingPolicy(ptype string, params ...interface{}) (bool, error) {
	tx.mutex.Lock()
	defer tx.mutex.Unlock()

	if err := tx.checkTransactionStatus(); err != nil {
		return false, err
	}

	rule := tx.buildRuleFromParams(params...)

	// Check if grouping policy exists in the buffered model.
	bufferedModel, err := tx.buffer.ApplyOperationsToModel(tx.buffer.GetModelSnapshot())
	if err != nil {
		return false, err
	}

	hasPolicy, err := bufferedModel.HasPolicy("g", ptype, rule)
	if !hasPolicy || err != nil {
		return false, err
	}

	// Add operation to buffer.
	op := persist.PolicyOperation{
		Type:       persist.OperationRemove,
		Section:    "g",
		PolicyType: ptype,
		Rules:      [][]string{rule},
	}
	tx.buffer.AddOperation(op)

	return true, nil
}

// GetBufferedModel returns the model as it would look after applying all buffered operations.
// This is useful for preview or validation purposes within the transaction.
func (tx *Transaction) GetBufferedModel() (model.Model, error) {
	tx.mutex.RLock()
	defer tx.mutex.RUnlock()

	if tx.committed || tx.rolledBack {
		return nil, errors.New("transaction is not active")
	}

	return tx.buffer.ApplyOperationsToModel(tx.buffer.GetModelSnapshot())
}

// HasOperations returns true if the transaction has any buffered operations.
func (tx *Transaction) HasOperations() bool {
	tx.mutex.RLock()
	defer tx.mutex.RUnlock()
	return tx.buffer.HasOperations()
}

// OperationCount returns the number of buffered operations in the transaction.
func (tx *Transaction) OperationCount() int {
	tx.mutex.RLock()
	defer tx.mutex.RUnlock()
	return tx.buffer.OperationCount()
}
