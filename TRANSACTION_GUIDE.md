# Transaction Consistency Guide

## Overview

Casbin provides built-in support for transaction consistency between policy operations and business data through the `TransactionalEnforcer`. This guide explains how to ensure atomic updates when modifying both Casbin authorization policies and business database records.

## The Problem

In applications that update both business data and Casbin policies, there's a risk of data inconsistency if operations happen in separate transactions:

```go
// ❌ NOT ATOMIC - Risk of inconsistency
db.UpdateUserRole(userId, "admin")        // Business transaction
enforcer.AddGroupingPolicy(userId, "admin") // Separate Casbin operation
// If second operation fails, data is inconsistent!
```

## The Solution: TransactionalEnforcer

Casbin's `TransactionalEnforcer` coordinates policy operations with database transactions to ensure atomicity.

### Key Features

- **Atomic Operations**: Policy changes and database updates happen atomically
- **Two-Phase Commit**: Ensures consistency between database and in-memory model
- **Conflict Detection**: Detects and prevents concurrent modification conflicts
- **Optimistic Locking**: Uses version numbers to detect concurrent changes
- **Rollback Support**: Automatically rolls back on failure

## Quick Start

### 1. Use a TransactionalAdapter

Your adapter must implement the `persist.TransactionalAdapter` interface:

```go
type TransactionalAdapter interface {
    Adapter
    BeginTransaction(ctx context.Context) (TransactionContext, error)
}
```

### 2. Create a TransactionalEnforcer

```go
import "github.com/casbin/casbin/v3"

// Create enforcer with a transactional adapter
enforcer, err := casbin.NewTransactionalEnforcer("model.conf", adapter)
if err != nil {
    log.Fatal(err)
}
```

### 3. Use Transactions

#### Option A: Using WithTransaction (Recommended)

```go
err := enforcer.WithTransaction(ctx, func(tx *casbin.Transaction) error {
    // Update business data
    if err := db.UpdateUserRole(userId, "admin"); err != nil {
        return err // Transaction will be rolled back
    }
    
    // Update Casbin policy
    if _, err := tx.AddGroupingPolicy(userId, "admin"); err != nil {
        return err // Transaction will be rolled back
    }
    
    return nil // Transaction will be committed
})
```

#### Option B: Manual Transaction Management

```go
// Begin transaction
tx, err := enforcer.BeginTransaction(ctx)
if err != nil {
    return err
}

// Add policy operations
ok, err := tx.AddGroupingPolicy("alice", "admin")
if err != nil {
    tx.Rollback()
    return err
}

ok, err = tx.AddPolicy("admin", "data1", "write")
if err != nil {
    tx.Rollback()
    return err
}

// Commit transaction
if err := tx.Commit(); err != nil {
    return err
}
```

## Complete Example: User Role Management

This example shows how to atomically update user roles in both the business database and Casbin:

```go
package main

import (
    "context"
    "database/sql"
    "fmt"
    "log"

    "github.com/casbin/casbin/v3"
    _ "github.com/lib/pq"
)

type UserService struct {
    db       *sql.DB
    enforcer *casbin.TransactionalEnforcer
}

// UpdateUserRole atomically updates user role in database and Casbin
func (s *UserService) UpdateUserRole(ctx context.Context, userId, oldRole, newRole string) error {
    return s.enforcer.WithTransaction(ctx, func(tx *casbin.Transaction) error {
        // Get database transaction from adapter
        // (This requires your adapter to provide access to the DB transaction)
        
        // Update user role in business database
        _, err := s.db.ExecContext(ctx, 
            "UPDATE users SET role = $1 WHERE id = $2", 
            newRole, userId)
        if err != nil {
            return fmt.Errorf("failed to update user role: %w", err)
        }
        
        // Remove old role mapping in Casbin
        if oldRole != "" {
            if _, err := tx.RemoveGroupingPolicy(userId, oldRole); err != nil {
                return fmt.Errorf("failed to remove old role: %w", err)
            }
        }
        
        // Add new role mapping in Casbin
        if _, err := tx.AddGroupingPolicy(userId, newRole); err != nil {
            return fmt.Errorf("failed to add new role: %w", err)
        }
        
        return nil
    })
}

// CreateUser atomically creates a user with initial permissions
func (s *UserService) CreateUser(ctx context.Context, userId, role string, permissions [][]string) error {
    return s.enforcer.WithTransaction(ctx, func(tx *casbin.Transaction) error {
        // Insert user into database
        _, err := s.db.ExecContext(ctx,
            "INSERT INTO users (id, role) VALUES ($1, $2)",
            userId, role)
        if err != nil {
            return fmt.Errorf("failed to create user: %w", err)
        }
        
        // Assign role in Casbin
        if _, err := tx.AddGroupingPolicy(userId, role); err != nil {
            return fmt.Errorf("failed to assign role: %w", err)
        }
        
        // Add initial permissions
        for _, perm := range permissions {
            if _, err := tx.AddPolicy(perm...); err != nil {
                return fmt.Errorf("failed to add permission: %w", err)
            }
        }
        
        return nil
    })
}
```

## Transaction Operations

The `Transaction` type supports all standard policy operations:

### Policy Operations
- `AddPolicy(params ...interface{}) (bool, error)`
- `AddPolicies(rules [][]string) (bool, error)`
- `RemovePolicy(params ...interface{}) (bool, error)`
- `RemovePolicies(rules [][]string) (bool, error)`
- `UpdatePolicy(oldPolicy, newPolicy []string) (bool, error)`

### Grouping Policy Operations
- `AddGroupingPolicy(params ...interface{}) (bool, error)`
- `RemoveGroupingPolicy(params ...interface{}) (bool, error)`

### Named Operations
- `AddNamedPolicy(ptype string, params ...interface{}) (bool, error)`
- `AddNamedPolicies(ptype string, rules [][]string) (bool, error)`
- `RemoveNamedPolicy(ptype string, params ...interface{}) (bool, error)`
- `RemoveNamedPolicies(ptype string, rules [][]string) (bool, error)`
- `UpdateNamedPolicy(ptype string, oldPolicy, newPolicy []string) (bool, error)`
- `AddNamedGroupingPolicy(ptype string, params ...interface{}) (bool, error)`
- `RemoveNamedGroupingPolicy(ptype string, params ...interface{}) (bool, error)`

### Transaction State
- `IsActive() bool` - Check if transaction is still active
- `IsCommitted() bool` - Check if transaction was committed
- `IsRolledBack() bool` - Check if transaction was rolled back
- `HasOperations() bool` - Check if transaction has buffered operations
- `OperationCount() int` - Get number of buffered operations
- `GetBufferedModel() (model.Model, error)` - Preview model state after operations

## Implementing a TransactionalAdapter

To use transactions, your adapter must implement the `persist.TransactionalAdapter` interface:

```go
package myadapter

import (
    "context"
    "database/sql"
    
    "github.com/casbin/casbin/v3/persist"
)

type MyAdapter struct {
    db *sql.DB
}

// BeginTransaction starts a database transaction
func (a *MyAdapter) BeginTransaction(ctx context.Context) (persist.TransactionContext, error) {
    tx, err := a.db.BeginTx(ctx, nil)
    if err != nil {
        return nil, err
    }
    
    return &MyTransactionContext{
        tx:      tx,
        adapter: a,
    }, nil
}

// MyTransactionContext wraps a database transaction
type MyTransactionContext struct {
    tx      *sql.Tx
    adapter *MyAdapter
}

func (tc *MyTransactionContext) Commit() error {
    return tc.tx.Commit()
}

func (tc *MyTransactionContext) Rollback() error {
    return tc.tx.Rollback()
}

func (tc *MyTransactionContext) GetAdapter() persist.Adapter {
    // Return an adapter that uses this transaction
    return &MyTransactionalAdapter{
        tx:      tc.tx,
        adapter: tc.adapter,
    }
}

// MyTransactionalAdapter is an adapter that operates within a transaction
type MyTransactionalAdapter struct {
    tx      *sql.Tx
    adapter *MyAdapter
}

func (a *MyTransactionalAdapter) AddPolicy(sec string, ptype string, rule []string) error {
    // Use a.tx instead of a.adapter.db for queries
    _, err := a.tx.Exec("INSERT INTO casbin_rule (...) VALUES (...)")
    return err
}

// Implement other Adapter methods using a.tx...
```

## Using with GORM Adapter

If you're using the GORM adapter, here's how to ensure transaction support:

```go
import (
    "github.com/casbin/gorm-adapter/v3"
    "gorm.io/gorm"
)

// Custom GORM adapter with transaction support
type GormTransactionalAdapter struct {
    *gormadapter.Adapter
    db *gorm.DB
}

func NewGormTransactionalAdapter(db *gorm.DB) (*GormTransactionalAdapter, error) {
    adapter, err := gormadapter.NewAdapterByDB(db)
    if err != nil {
        return nil, err
    }
    
    return &GormTransactionalAdapter{
        Adapter: adapter,
        db:      db,
    }, nil
}

func (a *GormTransactionalAdapter) BeginTransaction(ctx context.Context) (persist.TransactionContext, error) {
    tx := a.db.Begin()
    if tx.Error != nil {
        return nil, tx.Error
    }
    
    // Create adapter for this transaction
    txAdapter, err := gormadapter.NewAdapterByDB(tx)
    if err != nil {
        tx.Rollback()
        return nil, err
    }
    
    return &GormTransactionContext{
        tx:      tx,
        adapter: txAdapter,
    }, nil
}

type GormTransactionContext struct {
    tx      *gorm.DB
    adapter *gormadapter.Adapter
}

func (tc *GormTransactionContext) Commit() error {
    return tc.tx.Commit().Error
}

func (tc *GormTransactionContext) Rollback() error {
    return tc.tx.Rollback().Error
}

func (tc *GormTransactionContext) GetAdapter() persist.Adapter {
    return tc.adapter
}

// Usage
func main() {
    db, _ := gorm.Open(...)
    adapter, _ := NewGormTransactionalAdapter(db)
    enforcer, _ := casbin.NewTransactionalEnforcer("model.conf", adapter)
    
    enforcer.WithTransaction(ctx, func(tx *casbin.Transaction) error {
        // Your transactional operations
        return nil
    })
}
```

## Best Practices

### 1. Always Use WithTransaction for Automatic Cleanup

```go
// ✅ Good - Automatic rollback on error or panic
err := enforcer.WithTransaction(ctx, func(tx *casbin.Transaction) error {
    // Operations
    return nil
})

// ❌ Avoid - Manual management is error-prone
tx, _ := enforcer.BeginTransaction(ctx)
// Easy to forget rollback on error
tx.Commit()
```

### 2. Keep Transactions Short

```go
// ✅ Good - Quick transaction
err := enforcer.WithTransaction(ctx, func(tx *casbin.Transaction) error {
    tx.AddPolicy("alice", "data1", "read")
    tx.AddGroupingPolicy("alice", "admin")
    return nil
})

// ❌ Avoid - Long-running operations in transaction
err := enforcer.WithTransaction(ctx, func(tx *casbin.Transaction) error {
    // External API call - can be slow!
    result := callExternalAPI()
    tx.AddPolicy(result...)
    return nil
})
```

### 3. Handle Context Cancellation

```go
err := enforcer.WithTransaction(ctx, func(tx *casbin.Transaction) error {
    // Check context regularly in long operations
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
    }
    
    // Do work
    return nil
})
```

### 4. Use Buffered Model for Validation

```go
err := enforcer.WithTransaction(ctx, func(tx *casbin.Transaction) error {
    tx.AddPolicy("alice", "data1", "read")
    tx.AddGroupingPolicy("alice", "admin")
    
    // Preview what the model will look like after commit
    bufferedModel, err := tx.GetBufferedModel()
    if err != nil {
        return err
    }
    
    // Validate before committing
    hasPolicy, _ := bufferedModel.HasPolicy("p", "p", []string{"alice", "data1", "read"})
    if !hasPolicy {
        return errors.New("policy not added correctly")
    }
    
    return nil
})
```

## Error Handling

### Transaction Errors

```go
err := enforcer.WithTransaction(ctx, func(tx *casbin.Transaction) error {
    if _, err := tx.AddPolicy("alice", "data1", "read"); err != nil {
        // Error is automatically rolled back
        return fmt.Errorf("failed to add policy: %w", err)
    }
    return nil
})

if err != nil {
    // Handle transaction failure
    log.Printf("Transaction failed: %v", err)
}
```

### Conflict Detection

```go
// Transaction detects concurrent modifications
tx1, _ := enforcer.BeginTransaction(ctx)
tx2, _ := enforcer.BeginTransaction(ctx)

tx1.AddPolicy("alice", "data1", "read")
tx1.Commit() // Success

tx2.AddPolicy("bob", "data2", "write")
err := tx2.Commit() // May fail with conflict error if policies overlap
```

## Performance Considerations

1. **Batch Operations**: Use `AddPolicies` instead of multiple `AddPolicy` calls
2. **Transaction Duration**: Keep transactions as short as possible
3. **Database Locks**: Be aware of database locking behavior
4. **In-Memory Model**: Model updates happen after database commit

## Limitations

1. The adapter must implement `persist.TransactionalAdapter`
2. Transactions are not distributed - they only cover one database
3. In-memory model is updated after database commit (brief inconsistency window)

## Frequently Asked Questions

### Q: Can I use regular Enforcer methods during a transaction?

No, you must use the Transaction methods. Regular enforcer operations are not part of the transaction.

### Q: What happens if the adapter doesn't support transactions?

`BeginTransaction()` will return an error: "adapter does not support transactions"

### Q: Can I have multiple active transactions?

Yes, but be aware of potential conflicts. The last transaction to commit may fail if it conflicts with earlier commits.

### Q: How do I share the database transaction with my business code?

Your adapter's `TransactionContext` should provide access to the underlying database transaction so business code can use the same transaction.

## Related Resources

- [Casbin Documentation](https://casbin.org/docs)
- [Adapter List](https://casbin.org/docs/adapters)
- [Policy Management API](https://casbin.org/docs/management-api)
- [GORM Adapter](https://github.com/casbin/gorm-adapter)

## Support

For questions or issues:
- GitHub Issues: https://github.com/casbin/casbin/issues
- Discord: https://discord.gg/S5UjpzGZjN
