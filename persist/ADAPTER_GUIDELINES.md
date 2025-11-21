# Adapter Development Guidelines

## Handling Empty Strings in Database Adapters

### Background

Policy rules in Casbin can contain empty strings (`""`) as field values. Empty strings have semantic meaning - they represent omitted or irrelevant fields, which is different from wildcards or NULL values. It's crucial that adapters preserve empty strings correctly to maintain policy integrity.

### The Problem

When storing policy rules in databases, empty strings can be incorrectly converted to NULL values. This happens when adapters naively check if a string is empty and mark it as invalid/NULL:

```go
// ❌ WRONG: This converts empty strings to NULL
var dbField sql.NullString
dbField.String = fieldValue
dbField.Valid = fieldValue != "" // Empty string becomes NULL!
```

### The Solution

Casbin provides helper functions in the `persist` package to correctly handle empty strings:

#### 1. StringToNullable(s string) NullableString

Use this function when **saving** policy rules to the database:

```go
// ✅ CORRECT: Empty strings are preserved
nullable := persist.StringToNullable(fieldValue)
// nullable.Valid is always true, even for empty strings
```

#### 2. NullableToString(value string, valid bool) string

Use this function when **loading** policy rules from the database:

```go
// ✅ CORRECT: Properly handles both NULL and empty strings
fieldValue := persist.NullableToString(dbValue.String, dbValue.Valid)
```

### Example: Database Adapter Implementation

Here's how to implement a database adapter that correctly handles empty strings:

```go
package myadapter

import (
    "database/sql"
    "github.com/casbin/casbin/v2/model"
    "github.com/casbin/casbin/v2/persist"
)

type Adapter struct {
    db *sql.DB
}

// SavePolicy saves all policy rules to the database
func (a *Adapter) SavePolicy(model model.Model) error {
    tx, err := a.db.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()

    for ptype, ast := range model["p"] {
        for _, rule := range ast.Policy {
            // Convert each field using StringToNullable
            v0 := persist.StringToNullable(rule[0])
            v1 := persist.StringToNullable(rule[1])
            v2 := persist.StringToNullable(rule[2])
            
            _, err = tx.Exec(
                "INSERT INTO casbin_rule (ptype, v0, v1, v2) VALUES ($1, $2, $3, $4)",
                ptype, v0.Value, v1.Value, v2.Value,
            )
            if err != nil {
                return err
            }
        }
    }

    return tx.Commit()
}

// LoadPolicy loads all policy rules from the database
func (a *Adapter) LoadPolicy(model model.Model) error {
    rows, err := a.db.Query("SELECT ptype, v0, v1, v2, v3, v4, v5 FROM casbin_rule")
    if err != nil {
        return err
    }
    defer rows.Close()

    for rows.Next() {
        var ptype string
        var v0, v1, v2, v3, v4, v5 sql.NullString

        err = rows.Scan(&ptype, &v0, &v1, &v2, &v3, &v4, &v5)
        if err != nil {
            return err
        }

        // Convert each field using NullableToString
        rule := []string{
            persist.NullableToString(v0.String, v0.Valid),
            persist.NullableToString(v1.String, v1.Valid),
            persist.NullableToString(v2.String, v2.Valid),
            persist.NullableToString(v3.String, v3.Valid),
            persist.NullableToString(v4.String, v4.Valid),
            persist.NullableToString(v5.String, v5.Valid),
        }

        // Remove trailing empty strings if they're all empty
        for len(rule) > 0 && rule[len(rule)-1] == "" {
            rule = rule[:len(rule)-1]
        }

        if len(rule) == 0 {
            continue
        }

        err = persist.LoadPolicyArray(append([]string{ptype}, rule...), model)
        if err != nil {
            return err
        }
    }

    return rows.Err()
}
```

### Database Schema Considerations

When creating database tables for policy storage, ensure that columns can store empty strings:

```sql
-- ✅ CORRECT: Use VARCHAR/TEXT columns
CREATE TABLE casbin_rule (
    id SERIAL PRIMARY KEY,
    ptype VARCHAR(100) NOT NULL,
    v0 VARCHAR(100) NOT NULL DEFAULT '',  -- Can store empty strings
    v1 VARCHAR(100) NOT NULL DEFAULT '',
    v2 VARCHAR(100) NOT NULL DEFAULT '',
    v3 VARCHAR(100) NOT NULL DEFAULT '',
    v4 VARCHAR(100) NOT NULL DEFAULT '',
    v5 VARCHAR(100) NOT NULL DEFAULT ''
);
```

Note: Using `NOT NULL DEFAULT ''` ensures that fields always have a value (either a string or empty string), never NULL.

### Key Points

1. **Always use `persist.StringToNullable()` when saving to database** - This ensures empty strings are marked as valid and stored correctly.

2. **Always use `persist.NullableToString()` when loading from database** - This ensures proper conversion whether the database contains NULL or empty strings.

3. **Empty strings are meaningful** - They represent omitted/irrelevant fields, which is semantically different from wildcards or NULL values.

4. **Backward compatibility** - If your adapter previously stored empty strings as NULL, `NullableToString()` will convert them to empty strings, maintaining compatibility.

### Testing

Test your adapter with policies containing empty strings:

```go
func TestAdapterEmptyStrings(t *testing.T) {
    adapter := NewAdapter(/* ... */)
    e, _ := casbin.NewEnforcer("model.conf", adapter)
    
    // Add policy with empty string
    e.AddPolicy("alice", "", "read")
    
    // Reload
    e.LoadPolicy()
    
    // Verify empty string is preserved
    if !e.HasPolicy("alice", "", "read") {
        t.Error("Empty string was not preserved")
    }
}
```

### References

- Related issue: [#316](https://github.com/casbin/casbin/issues/316)
- This standard applies to all database adapters (PostgreSQL, MySQL, SQLite, etc.)
