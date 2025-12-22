# Multi-line Matcher Support

Casbin now supports multi-line matchers with block-style syntax, allowing you to write more complex and readable matcher expressions.

## Features

- **Let statements**: Define intermediate variables to break down complex expressions
- **Early returns**: Use `if` statements with `return` for conditional logic
- **Block syntax**: Write matchers within `{}` braces with multiple lines

## Syntax

### Basic Block Syntax

```ini
[matchers]
m = { \
    return r.sub == p.sub && r.obj == p.obj && r.act == p.act \
}
```

### With Let Statements

```ini
[matchers]
m = { \
    let role_match = g(r.sub, p.sub) \
    let obj_match = r.obj == p.obj \
    let act_match = r.act == p.act \
    return role_match && obj_match && act_match \
}
```

### With Nested Variables

```ini
[matchers]
m = { \
    let role_match = g(r.sub, p.sub) \
    let obj_direct_match = r.obj == p.obj \
    let obj_inherit_match = g2(r.obj, p.obj) \
    let obj_match = obj_direct_match || obj_inherit_match \
    let act_match = r.act == p.act \
    return role_match && obj_match && act_match \
}
```

### With Early Returns

```ini
[matchers]
m = { \
    let role_match = g(r.sub, p.sub) \
    if !role_match { \
        return false \
    } \
    if r.act != p.act { \
        return false \
    } \
    if r.obj == p.obj { \
        return true \
    } \
    if g2(r.obj, p.obj) { \
        return true \
    } \
    return false \
}
```

## How It Works

The multi-line matcher syntax is automatically transformed into a single-line expression that can be evaluated by the underlying govaluate engine. This transformation:

1. **Extracts let statements**: Variable definitions are identified and their expressions are stored
2. **Substitutes variables**: All variable references are replaced with their actual expressions
3. **Converts early returns**: `if` statements with returns are transformed into conditional logic using boolean operators

For example, the matcher:
```
{
    let role_match = g(r.sub, p.sub)
    let obj_match = r.obj == p.obj
    return role_match && obj_match
}
```

Is transformed into:
```
(g(r.sub, p.sub)) && (r.obj == p.obj)
```

## Important Notes

1. **Semicolons**: Do NOT use semicolons (`;`) at the end of statements. The config parser treats semicolons as comment markers and will strip them out.

2. **Line continuation**: Use backslash (`\`) at the end of each line to continue the matcher across multiple lines in the config file.

3. **Backward compatibility**: Traditional single-line matchers continue to work without any changes.

4. **In-memory models**: You can use multi-line matchers in code when creating models programmatically:
   ```go
   m := model.NewModel()
   m.AddDef("m", "m", `{
       let role_match = g(r.sub, p.sub)
       let obj_match = r.obj == p.obj
       return role_match && obj_match
   }`)
   ```

## Examples

See the `examples/` directory for complete working examples:
- `rbac_with_hierarchy_multiline_model.conf` - Multi-line matcher with let statements
- `rbac_with_early_return_model.conf` - Multi-line matcher with early returns

## Testing

Run the multi-line matcher tests:
```bash
go test -v -run TestMultiLineMatcher
```
