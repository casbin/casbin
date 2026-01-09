# Rate Limiting with Casbin

This example demonstrates how to use Casbin's rate limiting feature through the `RateLimitEffector`.

## Overview

Rate limiting in Casbin allows you to control the rate of requests based on various criteria such as subject, object, or action. This is useful for:

- Preventing abuse and brute-force attacks
- Ensuring fair resource allocation among users
- Protecting backend services from overload

## Configuration

### Model Configuration

To enable rate limiting, use the `rate_limit()` function in your policy effect definition:

```ini
[policy_effect]
e = rate_limit(max, unit, count_type, bucket)
```

**Parameters:**

- `max`: Maximum number of requests allowed within the time window (integer)
- `unit`: Time window unit - can be `second`, `minute`, `hour`, or `day`
- `count_type`: What to count:
  - `allow`: Only count allowed requests (useful for API quotas)
  - `all`: Count all requests including denied ones (useful for preventing brute-force attacks)
- `bucket`: How to group requests:
  - `sub`: Separate bucket per subject (user-based rate limiting)
  - `obj`: Separate bucket per object (resource-based rate limiting)
  - `act`: Separate bucket per action (operation-based rate limiting)
  - `all`: Single bucket for all requests (global rate limiting)

### Example Models

**User-based rate limiting (3 requests per second per user):**
```ini
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[policy_effect]
e = rate_limit(3, second, allow, sub)

[matchers]
m = r.sub == p.sub && r.obj == p.obj && r.act == p.act
```

**Global rate limiting (100 requests per minute for all users):**
```ini
[policy_effect]
e = rate_limit(100, minute, allow, all)
```

**Resource-based rate limiting (10 requests per hour per resource):**
```ini
[policy_effect]
e = rate_limit(10, hour, allow, obj)
```

**Brute-force protection (count all attempts, not just allowed ones):**
```ini
[policy_effect]
e = rate_limit(5, minute, all, sub)
```

## Usage

To use rate limiting in your code:

```go
package main

import (
    "github.com/casbin/casbin/v3"
    "github.com/casbin/casbin/v3/effector"
)

func main() {
    // Create enforcer with rate limit model
    e, err := casbin.NewEnforcer("model.conf", "policy.csv")
    if err != nil {
        panic(err)
    }

    // Set the rate limit effector (required!)
    rateLimitEft := effector.NewRateLimitEffector()
    e.SetEffector(rateLimitEft)

    // Now enforce with rate limiting
    ok, err := e.Enforce("alice", "data1", "read")
    if err != nil {
        // Handle error
    }
    if !ok {
        // Request denied (either by policy or rate limit)
    }
}
```

## Running the Example

```bash
cd examples/rate_limit_example
go run main.go
```

## How It Works

1. The `RateLimitEffector` maintains internal state for each bucket (counter and window expiration time)
2. When a request is enforced, the effector:
   - First checks if the request matches any policy rules
   - If it should be counted (based on `count_type`), updates the appropriate bucket counter
   - If the counter exceeds the limit, denies the request
   - When the time window expires, the counter is reset

3. Buckets are isolated based on the `bucket` parameter:
   - `sub`: Each subject (user) has its own bucket
   - `obj`: Each object (resource) has its own bucket
   - `act`: Each action has its own bucket
   - `all`: All requests share a single bucket

## Important Notes

- **You must call `SetEffector()` to enable rate limiting.** Without it, the default effector will be used.
- Rate limiting is stateful and maintained in memory. If your application restarts, counters are reset.
- The time windows are sliding windows - they start from the first request in the window.
- Bucket keys are automatically generated from the request context based on the bucket type.

## Advanced Use Cases

### Combining with RBAC

You can use rate limiting together with RBAC models:

```ini
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = rate_limit(10, minute, allow, sub)

[matchers]
m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act
```

### Different Limits for Different Users

To implement different rate limits for different users or roles, you would need to use multiple enforcers or implement a custom effector that reads limit values from policies.

## See Also

- [Main Casbin Documentation](https://casbin.org/docs/overview)
- [Policy Effects](https://casbin.org/docs/syntax-for-models#policy-effect)
- [Effector Interface](../../effector/effector.go)
