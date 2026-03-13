# IAM Policy Optimization Guide

## Overview

This document explains the performance characteristics of different policy configurations for AWS IAM-like authorization systems in Casbin.

## Background

When designing IAM-like systems with explicit allow and deny permissions, there are two common approaches:

### Option 1: Separate Allow Policies (No Effect Field)

```ini
[policy_definition]
p = sub, obj, act

[policy_effect]
e = some(where (p.eft == allow))
```

**Policies:**
```
p, alice, data1, read
p, bob, data2, write
```

### Option 2: Combined Policies with Effect Field

```ini
[policy_definition]
p = sub, obj, act, eft

[policy_effect]
e = some(where (p.eft == allow)) && !some(where (p.eft == deny))
```

**Policies:**
```
p, alice, data1, read, allow
p, bob, data2, write, allow
p, alice, data2, write, deny
```

## Performance Analysis

### Benchmark Results

Based on comprehensive benchmarks with 1000-5000 policies:

| Configuration | Time (ns/op) | Memory (B/op) | Allocs/op | Relative Speed |
|---------------|--------------|---------------|-----------|----------------|
| **Option 1** (No eft field) | 259,631 | 102,785 | 3,023 | **1.0x** (baseline) |
| **Option 2** (With eft field) | 527,584 | 187,125 | 6,019 | **2.03x slower** |

Large dataset (5000 policies):

| Configuration | Time (ns/op) | Relative Speed |
|---------------|--------------|----------------|
| **Option 1** | 1,443,129 | **1.0x** |
| **Option 2** | 2,903,399 | **2.01x slower** |

### Why Option 2 is Slower

The performance difference is **inherent to the algorithm semantics**:

#### Option 1: AllowOverrideEffect
- Evaluation can **short-circuit** on the first matching allow policy
- Average case: evaluates **N/2 policies**
- Best case: evaluates **1 policy** (if match is first)
- Worst case: evaluates **N policies** (if no match)

#### Option 2: AllowAndDenyEffect
- Evaluation **MUST check ALL policies** to ensure no deny rule exists
- Always evaluates **N policies** regardless of match position
- Cannot short-circuit because deny can appear anywhere in the policy list
- This is required for correct AWS IAM-like semantics

## Recommendations

### Choose Option 1 if:
- ✅ You only need "allow" permissions (no explicit deny)
- ✅ Performance is critical
- ✅ Policy set is large (>1000 policies)
- ✅ You can use separate policy types for deny rules

### Choose Option 2 if:
- ✅ You need AWS IAM-like explicit allow/deny semantics
- ✅ Deny rules can override allow rules from different sources
- ✅ You need to minimize policy count (vs duplicating for allow/deny)
- ✅ 2x performance overhead is acceptable for your use case

### Alternative: Priority-Based Evaluation

For fine-grained control over policy precedence:

```ini
[policy_definition]
p = sub, obj, act, eft

[policy_effect]
e = priority(p.eft) || deny
```

This evaluates policies in order and returns the first match, providing both performance and flexibility.

## Running Benchmarks

To verify performance characteristics in your environment:

```bash
go test -bench="BenchmarkIAM" -benchtime=3s -benchmem
```

This will run the included benchmarks:
- `BenchmarkIAMWithoutEffectField` - Tests Option 1
- `BenchmarkIAMWithEffectField` - Tests Option 2
- Large dataset variants of both

## Optimization Tips

1. **Order policies strategically**: Place most commonly matched policies first (helps Option 1)
2. **Use caching**: Enable `NewCachedEnforcer()` for repeated enforcement checks
3. **Minimize policy count**: Remove redundant or overlapping policies
4. **Consider policy granularity**: Fewer, broader policies are faster than many specific ones

## Conclusion

The ~2x performance difference between Option 1 and Option 2 is **not a bug** but a fundamental consequence of the different evaluation semantics. Choose the approach that best matches your security requirements and performance constraints.
