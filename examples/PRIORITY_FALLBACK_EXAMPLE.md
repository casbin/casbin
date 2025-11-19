# Priority-Based Fallback Policy Example

## Overview

This example demonstrates how to use Casbin's priority-based enforcer to implement a fallback policy pattern. This is useful when you want to:

1. Check specific policies first (normal policies)
2. Fall back to more general policies if no specific policy matches (fallback policies)
3. Have different actions/effects for each level

## How Priority Works in Casbin

In Casbin, when using the `priority(p.eft)` effect, policies are evaluated in order of their priority values:

- **Lower priority numbers = Higher priority** (evaluated first)
- When a policy matches, its effect is determined
- Higher priority effects override lower priority effects
- If no policy matches at all, the default effect is applied (usually deny)

## Example Configuration

### Model Configuration (`priority_fallback_model.conf`)

```ini
[request_definition]
r = sub, obj, act

[policy_definition]
p = priority, sub, obj, act, eft

[role_definition]
g = _, _

[policy_effect]
e = priority(p.eft) || deny

[matchers]
m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act
```

Key points:
- `priority` is the first field in the policy definition
- `priority(p.eft) || deny` means evaluate policies by priority, default to deny if none match
- The matcher checks for role membership using `g()`, object, and action

### Policy Configuration (`priority_fallback_policy.csv`)

```csv
# Normal policies (priority 10) - evaluated first
p, 10, alice, data1, read, allow
p, 10, alice, data1, write, allow
p, 10, bob, data2, read, allow
p, 10, bob, data2, write, deny

# Fallback policies (priority 100) - evaluated when no higher priority matches
p, 100, fallback_admin, data1, read, allow
p, 100, fallback_admin, data1, write, allow
p, 100, fallback_admin, data2, read, allow
p, 100, fallback_admin, data2, write, allow

# Role assignments
g, alice, fallback_admin
g, bob, fallback_admin
```

## Use Cases and Behavior

### Case 1: Direct Policy Match (Highest Priority)

```go
e.Enforce("alice", "data1", "read")  // Returns: true
e.Enforce("alice", "data1", "write") // Returns: true
```

- Alice has explicit policies at priority 10 for `data1`
- These policies match directly, so the priority 10 effect is used
- Priority 100 fallback policies are not needed

### Case 2: Fallback Policy Match

```go
e.Enforce("alice", "data2", "read")  // Returns: true
e.Enforce("alice", "data2", "write") // Returns: true
```

- Alice has NO explicit policy for `data2` at priority 10
- Alice is a member of `fallback_admin` role
- The matcher checks role membership: `g("alice", "fallback_admin")` returns true
- Falls back to priority 100 policies, which allow both read and write

### Case 3: Priority Override

```go
e.Enforce("bob", "data2", "write") // Returns: false
```

- Bob has an explicit DENY policy at priority 10 for `data2` write
- Bob is also a member of `fallback_admin` which allows `data2` write at priority 100
- Priority 10 (higher priority) takes precedence over priority 100
- Result: **denied** (priority 10 deny overrides priority 100 allow)

### Case 4: No Matching Policy

```go
e.Enforce("charlie", "data3", "read") // Returns: false
```

- Charlie has no explicit policies
- Charlie is not a member of any role
- No policy matches at any priority level
- Default effect from policy_effect (`deny`) is applied
- Result: **denied**

## Implementation in Go

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/casbin/casbin/v2"
)

func main() {
    // Initialize enforcer with priority fallback model and policies
    e, err := casbin.NewEnforcer(
        "examples/priority_fallback_model.conf",
        "examples/priority_fallback_policy.csv",
    )
    if err != nil {
        log.Fatal(err)
    }
    
    // Test normal policy matching
    allowed, _ := e.Enforce("alice", "data1", "read")
    fmt.Printf("alice can read data1: %v (matched priority 10 policy)\n", allowed)
    
    // Test fallback policy matching
    allowed, _ = e.Enforce("alice", "data2", "read")
    fmt.Printf("alice can read data2: %v (matched priority 100 fallback policy)\n", allowed)
    
    // Test priority override
    allowed, _ = e.Enforce("bob", "data2", "write")
    fmt.Printf("bob can write data2: %v (priority 10 deny overrides priority 100 allow)\n", allowed)
}
```

## Dynamic Policy Management

You can dynamically add policies with different priorities:

```go
// Add a high priority policy (overrides existing lower priority policies)
e.AddPolicy("1", "charlie", "data1", "read", "allow")

// Add a low priority fallback policy
e.AddPolicy("200", "guest_role", "public_data", "read", "allow")

// Add role assignment
e.AddRoleForUser("charlie", "guest_role")
```

## Best Practices

1. **Use meaningful priority ranges:**
   - 1-99: Critical/override policies
   - 100-199: Normal policies
   - 200-299: Fallback policies
   - 300+: Default/catch-all policies

2. **Keep priority numbers spread out:**
   - Allows inserting new priority levels between existing ones
   - Example: Use 10, 100, 1000 instead of 1, 2, 3

3. **Document your priority scheme:**
   - Make it clear what each priority range means
   - Document which priorities override which

4. **Test priority conflicts:**
   - Always test cases where multiple priorities could match
   - Verify that higher priority (lower number) takes precedence

## Advanced: Multiple Fallback Levels

You can have multiple fallback levels:

```csv
# Level 1: Direct user permissions (priority 10)
p, 10, alice, data1, read, allow

# Level 2: Department-level permissions (priority 50)
p, 50, engineering_dept, data1, read, allow
p, 50, engineering_dept, data1, write, deny

# Level 3: Organization-level permissions (priority 100)
p, 100, org_member, data1, read, allow

# Role assignments
g, alice, engineering_dept
g, alice, org_member
```

This creates a three-tier hierarchy:
1. Direct user permissions (highest priority)
2. Department-level permissions
3. Organization-level fallback permissions (lowest priority)

## Comparison with Other Patterns

### vs. Multiple Policy Definitions

Instead of using priorities, you could use multiple policy definitions (p, p2, p3). However:
- ✅ Priorities: Single policy type, clear ordering, easier to manage
- ❌ Multiple definitions: More complex model, harder to understand precedence

### vs. Conditional Matchers

Instead of priorities, you could use complex conditional matchers. However:
- ✅ Priorities: Clean separation, easy to understand, maintainable
- ❌ Conditionals: Can become very complex, harder to debug

## Troubleshooting

### Priority Not Working as Expected

1. **Check model configuration:**
   - Ensure `priority` field is first in policy definition
   - Ensure `priority(p.eft)` is used in policy_effect

2. **Verify priority values:**
   - Lower numbers = higher priority
   - Check that policies have the priority values you expect

3. **Check policy sorting:**
   - Casbin automatically sorts policies by priority
   - Use `e.GetPolicy()` to verify order

### Fallback Not Triggering

1. **Check matcher:**
   - Ensure matcher properly checks role membership with `g()`
   - Verify role assignments are correct

2. **Verify no higher priority matches:**
   - Higher priority matches will prevent fallback
   - Use `e.EnforceEx()` to see which policy matched

## References

- [Casbin Priority Model](https://casbin.org/docs/syntax-for-models#priority)
- [Casbin RBAC](https://casbin.org/docs/rbac)
- [Policy Effect](https://casbin.org/docs/syntax-for-models#policy-effect)
