# RBAC with Resource Scope

This document explains how to implement Azure RBAC-like functionality in Casbin, where the same role can be assigned to different users scoped to specific resources, preventing permission leakage.

## Problem Statement

In traditional RBAC implementations, when you assign a role to a user, that user gets access to all resources that the role has permissions for. This can lead to permission leakage when you want to reuse roles but scope them to specific resources.

For example, consider a scenario where:
- `user1` should have `reader` role for `resource1` only
- `user2` should have `reader` role for `resource2` only

In traditional RBAC, if you assign the `reader` role to both users, they would both get access to both resources if the role has permissions for both.

## Solution

Casbin provides a solution using 3-parameter grouping (`g = _, _, _`) to scope roles by resource. This allows you to assign the same role to different users with different resource scopes.

## Simple Resource Scope

### Model Configuration

```ini
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub, r.obj) && r.obj == p.obj && r.act == p.act
```

### Policy Configuration

```csv
p, reader, resource1, read
p, reader, resource2, read
p, writer, resource1, write
p, writer, resource2, write

g, user1, reader, resource1
g, user2, reader, resource2
g, user3, writer, resource1
```

### Usage

```go
e, _ := casbin.NewEnforcer("rbac_with_resource_scope_model.conf", "rbac_with_resource_scope_policy.csv")

// Check if user1 can read resource1 (returns true)
e.Enforce("user1", "resource1", "read")

// Check if user1 can read resource2 (returns false - different scope)
e.Enforce("user1", "resource2", "read")

// Check if user2 can read resource2 (returns true)
e.Enforce("user2", "resource2", "read")

// Get roles for user1 in resource1 scope
e.GetRolesForUser("user1", "resource1") // Returns ["reader"]

// Get roles for user1 in resource2 scope
e.GetRolesForUser("user1", "resource2") // Returns []

// Add a role for a user with resource scope
e.AddRoleForUser("user4", "reader", "resource1")

// Delete a role for a user with resource scope
e.DeleteRoleForUser("user4", "reader", "resource1")
```

## Multi-Tenant Resource Scope

For applications with multi-tenancy requirements, you can combine tenant and resource scoping by concatenating them in the grouping relationship.

### Model Configuration

```ini
[request_definition]
r = sub, tenant, obj, act

[policy_definition]
p = sub, tenant, obj, act

[role_definition]
g = _, _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub, r.tenant + "::" + r.obj) && r.tenant == p.tenant && r.obj == p.obj && r.act == p.act
```

### Policy Configuration

```csv
p, reader, tenant1, resource1, read
p, reader, tenant1, resource2, read
p, reader, tenant2, resource1, read
p, writer, tenant1, resource1, write

g, user1, reader, tenant1::resource1
g, user2, reader, tenant1::resource2
g, user3, reader, tenant2::resource1
g, user4, writer, tenant1::resource1
```

### Usage

```go
e, _ := casbin.NewEnforcer("rbac_with_resource_scope_tenant_model.conf", "rbac_with_resource_scope_tenant_policy.csv")

// Check if user1 can read resource1 in tenant1 (returns true)
e.Enforce("user1", "tenant1", "resource1", "read")

// Check if user1 can read resource2 in tenant1 (returns false - different resource scope)
e.Enforce("user1", "tenant1", "resource2", "read")

// Check if user1 can read resource1 in tenant2 (returns false - different tenant)
e.Enforce("user1", "tenant2", "resource1", "read")

// Get roles for user1 in tenant1::resource1 scope
e.GetRolesForUser("user1", "tenant1::resource1") // Returns ["reader"]

// Add a role for a user with tenant::resource scope
e.AddRoleForUser("user5", "reader", "tenant1::resource1")

// Delete a role for a user with tenant::resource scope
e.DeleteRoleForUser("user5", "reader", "tenant1::resource1")
```

## Comparison with Azure RBAC

This implementation provides functionality similar to Azure RBAC where:

1. **Role Definitions**: Define what actions can be performed (like Azure's built-in or custom roles)
2. **Role Assignments**: Assign roles to users with specific scopes (like Azure's role assignments at different scopes)
3. **Scope Hierarchy**: Support for multi-level scoping (tenant::resource is similar to Azure's subscription/resource group/resource hierarchy)

### Key Differences

- **Azure RBAC**: Uses a hierarchical scope model where permissions at a parent scope automatically apply to child scopes
- **Casbin Resource Scope**: Explicit scoping - permissions must be explicitly granted for each scope level

## API Compatibility

The standard Casbin RBAC APIs work seamlessly with resource-scoped roles by passing the scope as the domain parameter:

```go
// Get roles for a user in a specific scope
e.GetRolesForUser("user1", "resource1")

// Get users who have a role in a specific scope  
e.GetUsersForRole("reader", "resource1")

// Add a role for a user in a specific scope
e.AddRoleForUser("user3", "writer", "resource1")

// Delete a role for a user in a specific scope
e.DeleteRoleForUser("user3", "writer", "resource1")

// Check if a user has a role in a specific scope
e.HasRoleForUser("user1", "reader", "resource1")
```

For multi-tenant scenarios, use the concatenated scope:

```go
e.GetRolesForUser("user1", "tenant1::resource1")
e.AddRoleForUser("user5", "reader", "tenant1::resource1")
```

## Benefits

1. **Role Reusability**: Define roles once and reuse them across different resources
2. **Permission Isolation**: Users with the same role but different scopes cannot access each other's resources
3. **No Core Changes**: Uses existing Casbin capabilities (multi-domain role manager) without requiring library modifications
4. **Flexible Scoping**: Can be adapted for single-level (resource) or multi-level (tenant::resource) scoping
5. **Standard APIs**: Works with existing Casbin RBAC API methods

## Examples

See the following files for complete working examples:
- `examples/rbac_with_resource_scope_model.conf` - Simple resource scope model
- `examples/rbac_with_resource_scope_policy.csv` - Simple resource scope policy
- `examples/rbac_with_resource_scope_tenant_model.conf` - Multi-tenant resource scope model
- `examples/rbac_with_resource_scope_tenant_policy.csv` - Multi-tenant resource scope policy

## Tests

See `rbac_api_with_resource_scope_test.go` for comprehensive test cases covering:
- Simple resource scope
- Multi-tenant resource scope
- Multi-tenancy isolation
