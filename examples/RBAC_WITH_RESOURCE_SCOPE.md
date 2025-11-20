# RBAC with Resource Scope

This document explains how to implement Azure RBAC-like resource-scoped roles in Casbin, where the same role can be assigned to different users with different resource scopes.

## Overview

Resource-scoped RBAC allows you to assign the same role to different users, but limit the scope of that role to specific resources. This prevents one user's role permissions from affecting another user's permissions, even when they have the same role name.

## Problem Statement

In traditional RBAC, if you have:
```
p, reader, resource1, read
p, reader, resource2, read

g, user1, reader
g, user2, reader
```

Both `user1` and `user2` would have access to both `resource1` and `resource2` because they both have the `reader` role.

## Solution: Resource-Scoped Roles

With resource-scoped RBAC, you can scope roles to specific resources:

### Example 1: Simple Resource Scope

**Model** (`rbac_with_resource_scope_model.conf`):
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

**Policy** (`rbac_with_resource_scope_policy.csv`):
```csv
p, reader, resource1, read
p, reader, resource2, read

g, user1, reader, resource1
g, user2, reader, resource2
```

In this model:
- `user1` has the `reader` role scoped to `resource1`
- `user2` has the `reader` role scoped to `resource2`
- `user1` cannot access `resource2` and vice versa

### Example 2: Multi-Tenant with Resource Scope

For multi-tenant scenarios where you need to scope by both tenant and resource:

**Model** (`rbac_with_resource_scope_multitenancy_model.conf`):
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

**Policy** (`rbac_with_resource_scope_multitenancy_policy.csv`):
```csv
p, reader, tenant1, resource1, read
p, reader, tenant1, resource2, read
p, reader, tenant2, resource1, read

g, user1, reader, tenant1::resource1
g, user2, reader, tenant1::resource2
g, user3, reader, tenant2::resource1
```

In this model:
- `user1` has the `reader` role for `resource1` in `tenant1`
- `user2` has the `reader` role for `resource2` in `tenant1`
- `user3` has the `reader` role for `resource1` in `tenant2`
- Each user's permissions are isolated to their specific tenant and resource combination

## Usage

### Go Code Example

```go
package main

import (
    "fmt"
    "github.com/casbin/casbin/v2"
)

func main() {
    // Load the model and policy
    e, _ := casbin.NewEnforcer(
        "examples/rbac_with_resource_scope_model.conf",
        "examples/rbac_with_resource_scope_policy.csv",
    )

    // Check permissions
    ok, _ := e.Enforce("user1", "resource1", "read")
    fmt.Println(ok) // true

    ok, _ = e.Enforce("user1", "resource2", "read")
    fmt.Println(ok) // false

    ok, _ = e.Enforce("user2", "resource2", "read")
    fmt.Println(ok) // true

    // Get roles for a user with resource scope
    roles, _ := e.GetRolesForUser("user1", "resource1")
    fmt.Println(roles) // [reader]

    roles, _ = e.GetRolesForUser("user1", "resource2")
    fmt.Println(roles) // []

    // Add a new role assignment with resource scope
    e.AddRoleForUser("user3", "writer", "resource1")

    ok, _ = e.Enforce("user3", "resource1", "write")
    fmt.Println(ok) // true
}
```

### Multi-Tenant Example

```go
package main

import (
    "fmt"
    "github.com/casbin/casbin/v2"
)

func main() {
    // Load the multi-tenant model and policy
    e, _ := casbin.NewEnforcer(
        "examples/rbac_with_resource_scope_multitenancy_model.conf",
        "examples/rbac_with_resource_scope_multitenancy_policy.csv",
    )

    // Check permissions
    ok, _ := e.Enforce("user1", "tenant1", "resource1", "read")
    fmt.Println(ok) // true

    ok, _ = e.Enforce("user1", "tenant1", "resource2", "read")
    fmt.Println(ok) // false - user1 doesn't have access to resource2

    ok, _ = e.Enforce("user1", "tenant2", "resource1", "read")
    fmt.Println(ok) // false - user1 doesn't have access to tenant2

    // Get roles for a user with tenant and resource scope
    roles, _ := e.GetRolesForUser("user1", "tenant1::resource1")
    fmt.Println(roles) // [reader]

    // Add a new role assignment with tenant and resource scope
    e.AddRoleForUser("user4", "writer", "tenant1::resource1")
}
```

## Key Points

1. **Third Parameter in Grouping**: The third parameter in the `g` definition specifies the scope (resource or tenant::resource combination).

2. **Matcher Logic**: The matcher checks that the role assignment matches the requested resource scope using `g(r.sub, p.sub, r.obj)` or `g(r.sub, p.sub, r.tenant + "::" + r.obj)`.

3. **Isolation**: Users with the same role name but different scopes are completely isolated from each other.

4. **API Compatibility**: The standard Casbin RBAC APIs work with resource-scoped roles by passing the scope as the domain parameter:
   - `GetRolesForUser(user, scope)`
   - `GetUsersForRole(role, scope)`
   - `AddRoleForUser(user, role, scope)`

## Comparison with Azure RBAC

This implementation provides similar functionality to Azure RBAC:

| Azure RBAC | Casbin Resource-Scoped RBAC |
|------------|----------------------------|
| User assigned "Reader" role for Resource1 | `g, user1, reader, resource1` |
| User assigned "Reader" role for Resource2 | `g, user2, reader, resource2` |
| Roles are reusable across resources | Same role name with different scopes |
| Permissions isolated per assignment | Isolated through scope in grouping |

## See Also

- [rbac_with_domains_model.conf](rbac_with_domains_model.conf) - Traditional RBAC with domain support
- [rbac_with_resource_roles_model.conf](rbac_with_resource_roles_model.conf) - RBAC with resource hierarchies
