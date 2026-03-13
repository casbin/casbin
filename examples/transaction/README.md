# Casbin Transaction Example

This example demonstrates how to use Casbin's `TransactionalEnforcer` to ensure atomic updates between business data and authorization policies.

## Overview

The TransactionalEnforcer provides transaction support for Casbin policy operations, allowing you to:

- Update business data and Casbin policies atomically
- Ensure consistency between your database and authorization model
- Automatically rollback changes on errors
- Prevent concurrent modification conflicts

## Running the Example

```bash
cd examples/transaction
go run .
```

## What This Example Demonstrates

1. **Basic Transaction Usage**: Using `WithTransaction` for automatic transaction management
2. **Automatic Rollback**: How transactions automatically rollback on errors
3. **User Role Management**: Real-world scenario of updating user roles atomically
4. **Batch Operations**: Efficiently adding multiple policies in one transaction

## Code Structure

- `main.go` - Example scenarios demonstrating transaction usage
- `adapter.go` - Mock adapter implementation for the examples

## Key Concepts

### Using WithTransaction (Recommended)

```go
err := enforcer.WithTransaction(ctx, func(tx *casbin.Transaction) error {
    // Add policies
    tx.AddPolicy("alice", "data1", "read")
    tx.AddGroupingPolicy("alice", "admin")
    
    // Any error will cause automatic rollback
    if someCondition {
        return errors.New("validation failed")
    }
    
    return nil // Commits transaction
})
```

### Manual Transaction Management

```go
tx, err := enforcer.BeginTransaction(ctx)
if err != nil {
    return err
}

// Add operations
tx.AddPolicy("alice", "data1", "read")

// Commit or rollback
if err := tx.Commit(); err != nil {
    return err
}
```

## Using with Real Databases

To use transactions with a real database, your adapter must implement the `persist.TransactionalAdapter` interface:

```go
type TransactionalAdapter interface {
    Adapter
    BeginTransaction(ctx context.Context) (TransactionContext, error)
}
```

See the main [TRANSACTION_GUIDE.md](../../TRANSACTION_GUIDE.md) for complete implementation examples with GORM and other adapters.

## Next Steps

- Read the [Transaction Guide](../../TRANSACTION_GUIDE.md) for comprehensive documentation
- Implement `TransactionalAdapter` in your custom adapter
- Integrate transaction support into your application
- Check the [official Casbin documentation](https://casbin.org/docs)

## Related Resources

- [Casbin Documentation](https://casbin.org/docs)
- [Management API](https://casbin.org/docs/management-api)
- [RBAC API](https://casbin.org/docs/rbac-api)
- [Adapters](https://casbin.org/docs/adapters)
