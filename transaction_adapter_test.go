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
	"testing"
)

// GetAdapter hands out the transaction scoped adapter, so the caller can run
// its own database operations in the same transaction as the policy changes.
// See https://github.com/casbin/casbin/issues/1706.
func TestTransactionGetAdapter(t *testing.T) {
	adapter := NewMockTransactionalAdapter()
	e, err := NewTransactionalEnforcer("examples/rbac_model.conf", adapter)
	if err != nil {
		t.Fatalf("Failed to create transactional enforcer: %v", err)
	}

	err = e.WithTransaction(context.Background(), func(tx *Transaction) error {
		// The adapter of the transaction, not the one of the enforcer: for a
		// real adapter this is the one bound to the open database transaction.
		if tx.GetAdapter() == nil {
			t.Error("The transaction scoped adapter should be reachable")
		}
		_, addErr := tx.AddPolicy("admin", "data1", "read")
		return addErr
	})
	if err != nil {
		t.Fatalf("Failed to run transaction: %v", err)
	}

	if ok, err := e.HasPolicy("admin", "data1", "read"); err != nil || !ok {
		t.Fatalf("The policy added in the transaction should be in effect, got %v, %v", ok, err)
	}
}
