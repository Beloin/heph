# Struct Type System for BUILD Files - Implementation Plan

## Overview

Enable users to define custom struct types in BUILD files using Python/Starlark `class` syntax, with type-aware completion for struct fields.

## Phase 1: Define Struct Syntax

### Example BUILD file usage:

```python
class RunConfig:
    """Configuration for target execution"""
    env: dict
    sandbox: bool = False  
    timeout: int
    cmd: str

target(
    name = "app",
    driver = "bash",
    config = RunConfig(  # Struct instance
        env = {"FOO": "bar"},
        sandbox = True,
        timeout = 30,
        cmd = "echo hello"
    )
)
```

### Alternative inline syntax (already works via dict):

```python
target(
    name = "app", 
    driver = "bash",
    config = {  # No type info, but works
        "env": {"FOO": "bar"},
        "sandbox": True
    }
)
```

## Phase 2: Parse Struct Definitions

### 2.1 Add Tree-Sitter Query for Classes (query/query.go)

**New function: `processClass()`**

```go
// Find class definitions in tree
func processClass(node *tree_sitter.Node, text []byte, parentSym *symbol.Symbol)
```

**Query pattern (similar to functions):**

```scheme
(class_definition
  name: (identifier) @class.name
  body: (block) @class.body)
```

**Extract:**
- Class name → `Symbol.Name`
- Docstring (first string in body) → `Symbol.DocString`
- Fields (assignments with type annotations) → `Symbol.Symbols` (each as `FieldKind`)
- Inheritance (optional) → Could support `class Config(Dict):`

### 2.2 Field Parsing

For each assignment in class body:

```python
env: dict              # Field with type, no default
sandbox: bool = False  # Field with type and default value
```

**Create Field Symbols:**

```go
&symbol.Symbol{
    Name: "sandbox",
    Kind: symbol.FieldKind,
    Type: symbol.PrimitiveBool,  // Resolved type annotation
    Value: "False",              // Default value text
    DocString: "..."             // Optional inline comment
}
```

## Phase 3: Struct Instance Resolution

### 3.1 Type Annotations for Parameters

When parsing functions/calls, resolve type annotations:

```python
def target(name: str, config: RunConfig):
```

**Store in Parameter.Type:**

```go
Parameter{
    Name: "config",
    Type: RunConfigSymbol,  // Points to struct definition
    DocString: "...",
}
```

### 3.2 Detect Struct Instances in Calls

When parsing call arguments:

```python
target(
    config = RunConfig(...)  # Struct instance
)
```

**Or (inline dict with type hint):**

```python
target(
    config = {  # Dict literal
        "env": {...},
        "sandbox": True
    }
)
```

**Resolution logic:**
1. If `RunConfig(...)` - look up `RunConfig` in scope, create instance with `StructKind`
2. If dict literal - check if parameter expects a struct type, validate if known

## Phase 4: Type-Aware Completion

### 4.1 Complete Struct Type Names

**Scenario:** User types:

```python
target(
    config = Ru|  # Cursor here
)
```

**Completion:** Show all struct types in scope starting with "Ru"

### 4.2 Complete Struct Fields

**Scenario:** User types:

```python
target(
    config = RunConfig(
        en|  # Cursor here
    )
)
```

**Completion:** Show `RunConfig` fields starting with "en":
- `env: dict` 
- `env: dict = <required>` (indicate required vs optional)

**Implementation:**

```go
// In completion.go - detect struct instance
if callingConstructor(structType) {
    // Get struct definition
    structDef := getStructDefinition(structType.Name)
    // Show required fields first, then optional
    for _, field := range structDef.Symbols {
        // Add completion item for field
    }
}
```

### 4.3 Hover Information

**Hover over struct instance:**

```python
config = RunConfig(env=..., sandbox=False)
```

**Show:**

```
RunConfig {
  env: dict
  sandbox: bool = False
  timeout: int
  cmd: str
}
```

## Phase 5: Implementation Details

### Files to Modify:

1. **`internal/hlsp/runtime/query/query.go`**
   - Add `processClass()` function
   - Add class query to `QueryAll()`
   - Extract class fields with type annotations

2. **`internal/hlsp/runtime/symbol/symbol.go`**
   - Already has `StructKind` and `FieldKind` ✓
   - Consider adding `Required` field to `Symbol` (for required fields)

3. **`internal/hlsp/runtime/document/document.go`**
   - Ensure `Query()` finds struct definitions
   - Ensure `QueryType()` returns struct symbols

4. **`internal/hlsp/capabilities/lang/completion.go`**
   - Detect struct constructor calls
   - Complete struct field names
   - Show field types and defaults

5. **`internal/hlsp/capabilities/lang/hover.go`**
   - Show struct definition when hovering over struct type
   - Show field info when hovering over field usage

### New Files (optional):

6. **`internal/hlsp/runtime/query/class.go`**
   - Dedicated class/struct parsing logic
   - Field extraction with type resolution

## Phase 6: Edge Cases & Decisions

### Questions to Consider:

1. **Inheritance?** Support `class Child(Parent)` and inherit fields?
   - *Recommendation:* Start simple, no inheritance in v1

2. **Nested structs?** Support struct fields that are themselves structs?
   ```python
   class Outer:
       inner: InnerStruct
   ```
   - *Recommendation:* Yes, naturally supported via `Type` pointer

3. **Generic types?** Support `list[RunConfig]`?
   - *Recommendation:* Not in v1, stick to simple types

4. **Forward references?** Use struct before it's defined?
   - *Recommendation:* Support it via two-pass resolution

5. **Validate required fields?** Show error if required field missing?
   - *Recommendation:* No validation, just completion (as per user preference)

## Phase 7: Testing Strategy

### Test Cases:

1. **Parse struct definition:**
   ```python
   class Config:
       env: dict
       sandbox: bool = False
   ```

2. **Resolve struct type:**
   ```python
   def target(config: Config):  # Type resolves to Config struct
   ```

3. **Complete struct fields:**
   ```python
   target(config = Config(
       sa|  # Should complete 'sandbox'
   ))
   ```

4. **Hover over struct:**
   - Show all fields with types and defaults

## Implementation Priority:

### High Priority (MVP):
1. Parse `class` definitions → StructKind symbols
2. Parse field assignments with type annotations
3. Store struct definitions in scope
4. Basic completion for struct type names

### Medium Priority:
5. Complete struct fields in constructor calls
6. Show struct definition on hover
7. Required vs optional field completion

### Low Priority (Future):
8. Nested struct types
9. Forward references
10. Inheritance

## Estimated Effort:

- **Phase 1-3 (Parsing)**: 2-3 hours
- **Phase 4 (Completion)**: 1-2 hours  
- **Phase 5 (Hover)**: 1 hour
- **Testing**: 1-2 hours

**Total**: ~6-8 hours of focused development

## Current State (Pre-Implementation)

- **Type system exists**: `Symbol.Type` and `Parameter.Type` point to type symbols
- **Primitives work**: int, str, bool, list, dict all have sentinels
- **Type annotations parsed**: `def foo(a: str, b: int)` works
- **Driver schemas**: Protobuf schemas map to types (including nested messages → `ClassKind`)
- **StructKind defined but unused**: No way to define or use struct types in BUILD files
- **FieldKind defined**: Ready for use in struct fields

## User Decisions (from planning session):

- **Use case**: Type definitions in BUILD files (user-defined structs)
- **Type checking**: Just completion, no validation
- **Struct representation**: Reuse `StructKind` for both definitions and instances