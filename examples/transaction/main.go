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

// Package main demonstrates how to use Casbin's TransactionalEnforcer
// to ensure atomic updates between business data and authorization policies.
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/casbin/casbin/v3"
)

// This example demonstrates how to use TransactionalEnforcer to ensure
// that business data updates and Casbin policy updates happen atomically.
func main() {
	fmt.Println("Casbin Transaction Example")
	fmt.Println("===========================")
	fmt.Println()

	// Example 1: Basic transaction usage
	basicTransactionExample()

	// Example 2: Transaction with rollback
	transactionRollbackExample()

	// Example 3: Real-world scenario - User role management
	userRoleManagementExample()

	// Example 4: Batch operations in transaction
	batchOperationsExample()
}

// basicTransactionExample shows the basic usage of transactions
func basicTransactionExample() {
	fmt.Println("Example 1: Basic Transaction Usage")
	fmt.Println("-----------------------------------")

	// Create a transactional enforcer with mock adapter
	adapter := NewMockTransactionalAdapter()
	enforcer, err := casbin.NewTransactionalEnforcer("../rbac_model.conf", adapter)
	if err != nil {
		log.Fatalf("Failed to create enforcer: %v", err)
	}
	adapter.Enforcer = enforcer.Enforcer

	ctx := context.Background()

	// Use WithTransaction for automatic transaction management
	err = enforcer.WithTransaction(ctx, func(tx *casbin.Transaction) error {
		fmt.Println("  Adding policies in transaction...")

		// Add multiple policies atomically
		if _, err := tx.AddPolicy("alice", "data1", "read"); err != nil {
			return err
		}

		if _, err := tx.AddPolicy("bob", "data2", "write"); err != nil {
			return err
		}

		if _, err := tx.AddGroupingPolicy("alice", "admin"); err != nil {
			return err
		}

		fmt.Println("  Policies added successfully")
		return nil
	})

	if err != nil {
		log.Printf("Transaction failed: %v", err)
	} else {
		fmt.Println("  Transaction committed successfully")
	}

	fmt.Println()
}

// transactionRollbackExample demonstrates automatic rollback on error
func transactionRollbackExample() {
	fmt.Println("Example 2: Transaction Rollback on Error")
	fmt.Println("-----------------------------------------")

	adapter := NewMockTransactionalAdapter()
	enforcer, err := casbin.NewTransactionalEnforcer("../rbac_model.conf", adapter)
	if err != nil {
		log.Fatalf("Failed to create enforcer: %v", err)
	}
	adapter.Enforcer = enforcer.Enforcer

	ctx := context.Background()

	err = enforcer.WithTransaction(ctx, func(tx *casbin.Transaction) error {
		fmt.Println("  Adding first policy...")
		if _, err := tx.AddPolicy("charlie", "data1", "read"); err != nil {
			return err
		}

		fmt.Println("  Simulating an error...")
		// Simulate an error (e.g., business logic validation failure)
		return fmt.Errorf("business validation failed")
	})

	if err != nil {
		fmt.Printf("  Transaction rolled back: %v\n", err)
		fmt.Println("  All changes were reverted")
	}

	fmt.Println()
}

// userRoleManagementExample shows a real-world scenario
func userRoleManagementExample() {
	fmt.Println("Example 3: User Role Management")
	fmt.Println("--------------------------------")

	adapter := NewMockTransactionalAdapter()
	enforcer, err := casbin.NewTransactionalEnforcer("../rbac_model.conf", adapter)
	if err != nil {
		log.Fatalf("Failed to create enforcer: %v", err)
	}
	adapter.Enforcer = enforcer.Enforcer

	ctx := context.Background()

	// Simulate updating a user's role
	// In a real application, this would also update the database
	updateUserRole := func(userId, oldRole, newRole string) error {
		return enforcer.WithTransaction(ctx, func(tx *casbin.Transaction) error {
			fmt.Printf("  Updating %s from %s to %s...\n", userId, oldRole, newRole)

			// In real code, you would update the database here:
			// _, err := db.ExecContext(ctx, "UPDATE users SET role = $1 WHERE id = $2", newRole, userId)

			// Remove old role
			if oldRole != "" {
				if _, err := tx.RemoveGroupingPolicy(userId, oldRole); err != nil {
					return fmt.Errorf("failed to remove old role: %w", err)
				}
			}

			// Add new role
			if _, err := tx.AddGroupingPolicy(userId, newRole); err != nil {
				return fmt.Errorf("failed to add new role: %w", err)
			}

			// Add role-specific permissions
			// Note: In a real application, you would check these errors.
			// For this example, we're showing the pattern and ignoring errors
			// since the policies might already exist.
			switch newRole {
			case "admin":
				tx.AddPolicy("admin", "data1", "write")
				tx.AddPolicy("admin", "data2", "write")
			case "user":
				tx.AddPolicy(userId, "data1", "read")
			}

			fmt.Printf("  Successfully updated %s to %s\n", userId, newRole)
			return nil
		})
	}

	// Update a user's role
	if err := updateUserRole("dave", "", "admin"); err != nil {
		log.Printf("Failed to update user role: %v", err)
	}

	fmt.Println()
}

// batchOperationsExample demonstrates batch operations in transactions
func batchOperationsExample() {
	fmt.Println("Example 4: Batch Operations in Transaction")
	fmt.Println("-------------------------------------------")

	adapter := NewMockTransactionalAdapter()
	enforcer, err := casbin.NewTransactionalEnforcer("../rbac_model.conf", adapter)
	if err != nil {
		log.Fatalf("Failed to create enforcer: %v", err)
	}
	adapter.Enforcer = enforcer.Enforcer

	ctx := context.Background()

	err = enforcer.WithTransaction(ctx, func(tx *casbin.Transaction) error {
		fmt.Println("  Adding batch policies...")

		// Add multiple policies in one operation (more efficient)
		policies := [][]string{
			{"eve", "data1", "read"},
			{"eve", "data2", "read"},
			{"eve", "data3", "write"},
		}

		if _, err := tx.AddPolicies(policies); err != nil {
			return fmt.Errorf("failed to add policies: %w", err)
		}

		// Check buffered model state before commit
		bufferedModel, err := tx.GetBufferedModel()
		if err != nil {
			return err
		}

		// Validate that policies were added correctly
		hasPolicy, _ := bufferedModel.HasPolicy("p", "p", []string{"eve", "data1", "read"})
		if !hasPolicy {
			return fmt.Errorf("policy validation failed")
		}

		fmt.Printf("  Successfully added %d policies\n", len(policies))
		fmt.Printf("  Transaction has %d operations\n", tx.OperationCount())

		return nil
	})

	if err != nil {
		log.Printf("Batch operation failed: %v", err)
	} else {
		fmt.Println("  Batch operation committed successfully")
	}

	fmt.Println()
}
