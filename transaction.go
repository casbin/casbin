// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package casbin

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/casbin/casbin/v3/model"
	"github.com/casbin/casbin/v3/persist"
)

const (
	// Default timeout duration for lock acquisition.
	defaultLockTimeout = 30 * time.Second
)

// Transaction represents a Casbin transaction.
// It provides methods to perform policy operations within a transaction.
// and commit or rollback all changes atomically.
type Transaction struct {
	id          string                     // Unique transaction identifier.
	enforcer    *TransactionalEnforcer     // Reference to the transactional enforcer.
	buffer      *TransactionBuffer         // Buffer for policy operations.
	txContext   persist.TransactionContext // Database transaction context.
	ctx         context.Context            // Context for the transaction.
	baseVersion int64                      // Model version at transaction start.
	committed   bool                       // Whether the transaction has been committed.
	rolledBack  bool                       // Whether the transaction has been rolled back.
	startTime   time.Time                  // Transaction start timestamp.
	mutex       sync.RWMutex               // Protects transaction state.
}

// GetAdapter returns the adapter that operates within this transaction.
// Writes done through it join the same database transaction as the buffered
// policy changes, so non-policy operations can be made atomic together with the
// policy changes, e.g. with the GORM adapter:
//
//	err := e.WithTransaction(ctx, func(tx *casbin.Transaction) error {
//		db := tx.GetAdapter().(*gormadapter.Adapter).GetDb().
//			Session(&gorm.Session{NewDB: true})
//		if err := db.Where("name = ?", "admin").Delete(&Role{}).Error; err != nil {
//			return err
//		}
//		_, err := tx.RemoveGroupingPolicy("alice", "admin")
//		return err
//	})
//
// Mind the Session(&gorm.Session{NewDB: true}): the GORM adapter pins its
// *gorm.DB to the policy table, and without resetting the statement every query
// would silently run against casbin_rule instead of the caller's own table.
func (tx *Transaction) GetAdapter() persist.Adapter {
	return tx.txContext.GetAdapter()
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

// getBufferedModel returns the model with all buffered operations applied.
// The caller must hold tx.mutex.
func (tx *Transaction) getBufferedModel() (model.Model, error) {
	if err := tx.checkTransactionStatus(); err != nil {
		return nil, err
	}

	return tx.buffer.ApplyOperationsToModel(tx.buffer.GetModelSnapshot())
}

// bufferAdd buffers an add operation for the rules that are not present yet.
// Rules that already exist in the buffered model are skipped, and false is
// returned when nothing is left to add.
func (tx *Transaction) bufferAdd(sec string, ptype string, rules [][]string) (bool, error) {
	tx.mutex.Lock()
	defer tx.mutex.Unlock()

	if len(rules) == 0 {
		if err := tx.checkTransactionStatus(); err != nil {
			return false, err
		}
		return false, nil
	}

	bufferedModel, err := tx.getBufferedModel()
	if err != nil {
		return false, err
	}

	var validRules [][]string
	for _, rule := range rules {
		hasPolicy, err := bufferedModel.HasPolicy(sec, ptype, rule)
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

	tx.buffer.AddOperation(persist.PolicyOperation{
		Type:       persist.OperationAdd,
		Section:    sec,
		PolicyType: ptype,
		Rules:      validRules,
	})

	return true, nil
}

// bufferRemove buffers a remove operation for the rules that are present.
// Rules that are missing from the buffered model are skipped, and false is
// returned when nothing is left to remove.
func (tx *Transaction) bufferRemove(sec string, ptype string, rules [][]string) (bool, error) {
	tx.mutex.Lock()
	defer tx.mutex.Unlock()

	if len(rules) == 0 {
		if err := tx.checkTransactionStatus(); err != nil {
			return false, err
		}
		return false, nil
	}

	bufferedModel, err := tx.getBufferedModel()
	if err != nil {
		return false, err
	}

	var validRules [][]string
	for _, rule := range rules {
		hasPolicy, err := bufferedModel.HasPolicy(sec, ptype, rule)
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

	tx.buffer.AddOperation(persist.PolicyOperation{
		Type:       persist.OperationRemove,
		Section:    sec,
		PolicyType: ptype,
		Rules:      validRules,
	})

	return true, nil
}

// bufferRemoveFiltered resolves the filter against the buffered model and
// buffers the matched rules for removal. Resolving the filter here keeps the
// operation a plain remove, so it works with every adapter and takes part in
// conflict detection like any other remove.
func (tx *Transaction) bufferRemoveFiltered(sec string, ptype string, fieldIndex int, fieldValues []string) (bool, error) {
	tx.mutex.Lock()
	defer tx.mutex.Unlock()

	bufferedModel, err := tx.getBufferedModel()
	if err != nil {
		return false, err
	}

	matched, err := bufferedModel.GetFilteredPolicy(sec, ptype, fieldIndex, fieldValues...)
	if err != nil {
		return false, err
	}
	if len(matched) == 0 {
		return false, nil
	}

	// The matched rules belong to a throwaway model copy, so keep our own copy.
	rules := make([][]string, 0, len(matched))
	for _, rule := range matched {
		rules = append(rules, append([]string(nil), rule...))
	}

	tx.buffer.AddOperation(persist.PolicyOperation{
		Type:       persist.OperationRemove,
		Section:    sec,
		PolicyType: ptype,
		Rules:      rules,
	})

	return true, nil
}

// bufferUpdate buffers an update operation. The update is all-or-nothing: it is
// buffered only if every old rule exists and no new rule exists yet.
func (tx *Transaction) bufferUpdate(sec string, ptype string, oldRules [][]string, newRules [][]string) (bool, error) {
	tx.mutex.Lock()
	defer tx.mutex.Unlock()

	if len(oldRules) != len(newRules) {
		return false, fmt.Errorf("the length of oldRules should be equal to the length of newRules, but got the length of oldRules is %d, the length of newRules is %d", len(oldRules), len(newRules))
	}

	if len(oldRules) == 0 {
		if err := tx.checkTransactionStatus(); err != nil {
			return false, err
		}
		return false, nil
	}

	bufferedModel, err := tx.getBufferedModel()
	if err != nil {
		return false, err
	}

	for i, oldRule := range oldRules {
		hasOldPolicy, err := bufferedModel.HasPolicy(sec, ptype, oldRule)
		if err != nil {
			return false, err
		}
		if !hasOldPolicy {
			return false, nil
		}

		hasNewPolicy, err := bufferedModel.HasPolicy(sec, ptype, newRules[i])
		if err != nil {
			return false, err
		}
		if hasNewPolicy {
			return false, nil
		}
	}

	tx.buffer.AddOperation(persist.PolicyOperation{
		Type:       persist.OperationUpdate,
		Section:    sec,
		PolicyType: ptype,
		Rules:      newRules,
		OldRules:   oldRules,
	})

	return true, nil
}

// AddPolicy adds a policy within the transaction.
// The policy is buffered and will be applied when the transaction is committed.
func (tx *Transaction) AddPolicy(params ...interface{}) (bool, error) {
	return tx.AddNamedPolicy("p", params...)
}

// AddNamedPolicy adds a named policy within the transaction.
// The policy is buffered and will be applied when the transaction is committed.
func (tx *Transaction) AddNamedPolicy(ptype string, params ...interface{}) (bool, error) {
	return tx.bufferAdd("p", ptype, [][]string{tx.buildRuleFromParams(params...)})
}

// AddPolicies adds multiple policies within the transaction.
func (tx *Transaction) AddPolicies(rules [][]string) (bool, error) {
	return tx.AddNamedPolicies("p", rules)
}

// AddNamedPolicies adds multiple named policies within the transaction.
func (tx *Transaction) AddNamedPolicies(ptype string, rules [][]string) (bool, error) {
	return tx.bufferAdd("p", ptype, rules)
}

// RemovePolicy removes a policy within the transaction.
func (tx *Transaction) RemovePolicy(params ...interface{}) (bool, error) {
	return tx.RemoveNamedPolicy("p", params...)
}

// RemoveNamedPolicy removes a named policy within the transaction.
func (tx *Transaction) RemoveNamedPolicy(ptype string, params ...interface{}) (bool, error) {
	return tx.bufferRemove("p", ptype, [][]string{tx.buildRuleFromParams(params...)})
}

// RemovePolicies removes multiple policies within the transaction.
func (tx *Transaction) RemovePolicies(rules [][]string) (bool, error) {
	return tx.RemoveNamedPolicies("p", rules)
}

// RemoveNamedPolicies removes multiple named policies within the transaction.
func (tx *Transaction) RemoveNamedPolicies(ptype string, rules [][]string) (bool, error) {
	return tx.bufferRemove("p", ptype, rules)
}

// RemoveFilteredPolicy removes the policies that match the filter within the transaction.
// An empty string in fieldValues means "match any value" for that field.
func (tx *Transaction) RemoveFilteredPolicy(fieldIndex int, fieldValues ...string) (bool, error) {
	return tx.RemoveFilteredNamedPolicy("p", fieldIndex, fieldValues...)
}

// RemoveFilteredNamedPolicy removes the named policies that match the filter within the transaction.
// An empty string in fieldValues means "match any value" for that field.
func (tx *Transaction) RemoveFilteredNamedPolicy(ptype string, fieldIndex int, fieldValues ...string) (bool, error) {
	return tx.bufferRemoveFiltered("p", ptype, fieldIndex, fieldValues)
}

// UpdatePolicy updates a policy within the transaction.
func (tx *Transaction) UpdatePolicy(oldPolicy []string, newPolicy []string) (bool, error) {
	return tx.UpdateNamedPolicy("p", oldPolicy, newPolicy)
}

// UpdateNamedPolicy updates a named policy within the transaction.
func (tx *Transaction) UpdateNamedPolicy(ptype string, oldPolicy []string, newPolicy []string) (bool, error) {
	return tx.bufferUpdate("p", ptype, [][]string{oldPolicy}, [][]string{newPolicy})
}

// UpdatePolicies updates multiple policies within the transaction.
// Nothing is buffered unless every old policy exists and no new policy exists yet.
func (tx *Transaction) UpdatePolicies(oldPolicies [][]string, newPolicies [][]string) (bool, error) {
	return tx.UpdateNamedPolicies("p", oldPolicies, newPolicies)
}

// UpdateNamedPolicies updates multiple named policies within the transaction.
// Nothing is buffered unless every old policy exists and no new policy exists yet.
func (tx *Transaction) UpdateNamedPolicies(ptype string, oldPolicies [][]string, newPolicies [][]string) (bool, error) {
	return tx.bufferUpdate("p", ptype, oldPolicies, newPolicies)
}

// AddGroupingPolicy adds a grouping policy within the transaction.
func (tx *Transaction) AddGroupingPolicy(params ...interface{}) (bool, error) {
	return tx.AddNamedGroupingPolicy("g", params...)
}

// AddNamedGroupingPolicy adds a named grouping policy within the transaction.
func (tx *Transaction) AddNamedGroupingPolicy(ptype string, params ...interface{}) (bool, error) {
	return tx.bufferAdd("g", ptype, [][]string{tx.buildRuleFromParams(params...)})
}

// AddGroupingPolicies adds multiple grouping policies within the transaction.
func (tx *Transaction) AddGroupingPolicies(rules [][]string) (bool, error) {
	return tx.AddNamedGroupingPolicies("g", rules)
}

// AddNamedGroupingPolicies adds multiple named grouping policies within the transaction.
func (tx *Transaction) AddNamedGroupingPolicies(ptype string, rules [][]string) (bool, error) {
	return tx.bufferAdd("g", ptype, rules)
}

// RemoveGroupingPolicy removes a grouping policy within the transaction.
func (tx *Transaction) RemoveGroupingPolicy(params ...interface{}) (bool, error) {
	return tx.RemoveNamedGroupingPolicy("g", params...)
}

// RemoveNamedGroupingPolicy removes a named grouping policy within the transaction.
func (tx *Transaction) RemoveNamedGroupingPolicy(ptype string, params ...interface{}) (bool, error) {
	return tx.bufferRemove("g", ptype, [][]string{tx.buildRuleFromParams(params...)})
}

// RemoveGroupingPolicies removes multiple grouping policies within the transaction.
func (tx *Transaction) RemoveGroupingPolicies(rules [][]string) (bool, error) {
	return tx.RemoveNamedGroupingPolicies("g", rules)
}

// RemoveNamedGroupingPolicies removes multiple named grouping policies within the transaction.
func (tx *Transaction) RemoveNamedGroupingPolicies(ptype string, rules [][]string) (bool, error) {
	return tx.bufferRemove("g", ptype, rules)
}

// RemoveFilteredGroupingPolicy removes the grouping policies that match the filter within the transaction.
// An empty string in fieldValues means "match any value" for that field.
func (tx *Transaction) RemoveFilteredGroupingPolicy(fieldIndex int, fieldValues ...string) (bool, error) {
	return tx.RemoveFilteredNamedGroupingPolicy("g", fieldIndex, fieldValues...)
}

// RemoveFilteredNamedGroupingPolicy removes the named grouping policies that match
// the filter within the transaction.
// An empty string in fieldValues means "match any value" for that field.
func (tx *Transaction) RemoveFilteredNamedGroupingPolicy(ptype string, fieldIndex int, fieldValues ...string) (bool, error) {
	return tx.bufferRemoveFiltered("g", ptype, fieldIndex, fieldValues)
}

// UpdateGroupingPolicy updates a grouping policy within the transaction.
func (tx *Transaction) UpdateGroupingPolicy(oldRule []string, newRule []string) (bool, error) {
	return tx.UpdateNamedGroupingPolicy("g", oldRule, newRule)
}

// UpdateNamedGroupingPolicy updates a named grouping policy within the transaction.
func (tx *Transaction) UpdateNamedGroupingPolicy(ptype string, oldRule []string, newRule []string) (bool, error) {
	return tx.bufferUpdate("g", ptype, [][]string{oldRule}, [][]string{newRule})
}

// UpdateGroupingPolicies updates multiple grouping policies within the transaction.
// Nothing is buffered unless every old rule exists and no new rule exists yet.
func (tx *Transaction) UpdateGroupingPolicies(oldRules [][]string, newRules [][]string) (bool, error) {
	return tx.UpdateNamedGroupingPolicies("g", oldRules, newRules)
}

// UpdateNamedGroupingPolicies updates multiple named grouping policies within the transaction.
// Nothing is buffered unless every old rule exists and no new rule exists yet.
func (tx *Transaction) UpdateNamedGroupingPolicies(ptype string, oldRules [][]string, newRules [][]string) (bool, error) {
	return tx.bufferUpdate("g", ptype, oldRules, newRules)
}

// GetBufferedModel returns the model as it would look after applying all buffered operations.
// This is useful for preview or validation purposes within the transaction.
func (tx *Transaction) GetBufferedModel() (model.Model, error) {
	tx.mutex.RLock()
	defer tx.mutex.RUnlock()

	return tx.getBufferedModel()
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

// tryLockWithTimeout attempts to acquire the lock within the specified timeout period.
func tryLockWithTimeout(lock *sync.Mutex, startTime time.Time, maxWait time.Duration) bool {
	// Calculate remaining wait time based on transaction start time.
	remainingTime := maxWait - time.Since(startTime)
	if remainingTime <= 0 {
		return false
	}

	// Create a context with timeout for lock acquisition.
	ctx, cancel := context.WithTimeout(context.Background(), remainingTime)
	defer cancel()

	// Use channel for timeout control.
	done := make(chan bool, 1)
	go func() {
		lock.Lock()
		done <- true
	}()

	// Wait for either lock acquisition or timeout.
	select {
	case <-done:
		return true
	case <-ctx.Done():
		return false
	}
}
