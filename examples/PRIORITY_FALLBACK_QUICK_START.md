# Quick Start: Priority-Based Fallback Policies

This is a quick reference guide for implementing priority-based fallback policies in Casbin Go.

## Problem

You want to:
1. Check specific user policies first (e.g., alice can read data1)
2. If no specific policy exists, fall back to role-based policies (e.g., alice is in admin group)
3. Have different actions/effects for each level

## Solution

Use Casbin's priority-based enforcer where **lower priority numbers = higher priority**.

## Minimal Example

### 1. Create Model (`model.conf`)

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

### 2. Create Policies (`policy.csv`)

```csv
# High priority (specific user policies)
p, 10, alice, data1, read, allow
p, 10, bob, data2, write, deny

# Low priority (fallback via roles)
p, 100, admin_role, data1, read, allow
p, 100, admin_role, data2, write, allow

# Role assignments
g, alice, admin_role
g, bob, admin_role
```

### 3. Use in Code

```go
package main

import (
    "fmt"
    "github.com/casbin/casbin/v2"
)

func main() {
    e, _ := casbin.NewEnforcer("model.conf", "policy.csv")
    
    // alice has explicit rule at priority 10 for data1
    allowed, _ := e.Enforce("alice", "data1", "read")
    fmt.Printf("alice read data1: %v (matched priority 10)\n", allowed)  // true
    
    // alice has no explicit rule for data2, falls back to admin_role at priority 100
    allowed, _ = e.Enforce("alice", "data2", "write")
    fmt.Printf("alice write data2: %v (matched priority 100)\n", allowed)  // true
    
    // bob has explicit DENY at priority 10 (overrides admin_role ALLOW at priority 100)
    allowed, _ = e.Enforce("bob", "data2", "write")
    fmt.Printf("bob write data2: %v (priority 10 overrides 100)\n", allowed)  // false
}
```

## How It Works

1. **Request comes in**: `Enforce("alice", "data2", "write")`
2. **Check priority 10 policies**: No match for alice + data2 + write
3. **Check priority 100 policies**: Matches via admin_role membership → ALLOW
4. **Return result**: true

## Priority Rules

- **Lower number = Higher priority**: 10 beats 100
- **First match wins at each priority**: Within same priority, first matching rule wins
- **Evaluation stops**: Once a policy matches, lower priorities are ignored

## Common Use Cases

### User > Department > Organization
```csv
p, 10, alice, data1, read, allow          # User-specific (highest)
p, 50, engineering, data1, read, allow    # Department-level
p, 100, all_employees, data1, read, allow # Organization-wide (lowest)
```

### Explicit Deny > Default Allow
```csv
p, 10, bob, sensitive_data, read, deny    # Explicit deny (highest)
p, 100, default_role, sensitive_data, read, allow  # Default allow (lowest)
```

### Dynamic Priority
```go
// Add high priority override
e.AddPolicy("1", "charlie", "data1", "write", "deny")

// Add low priority default
e.AddPolicy("999", "guest_role", "public_data", "read", "allow")
```

## Debugging

Use `EnforceEx` to see which policy matched:

```go
allowed, explain, _ := e.EnforceEx("alice", "data1", "read")
fmt.Printf("Result: %v\n", allowed)
fmt.Printf("Matched: %v\n", explain)  // Shows [priority, sub, obj, act, eft]
```

## Next Steps

- See [PRIORITY_FALLBACK_EXAMPLE.md](./PRIORITY_FALLBACK_EXAMPLE.md) for detailed documentation
- Run the example: `go run priority_fallback_example.go`
- Check test: `go test -v -run TestPriorityFallback`

## Key Takeaways

✅ Priority-based enforcement allows clean fallback patterns  
✅ Lower numbers = higher priority (10 > 100)  
✅ Great for user-specific rules with organization-wide defaults  
✅ Explicit denies can override default allows  
✅ Easy to understand and maintain
