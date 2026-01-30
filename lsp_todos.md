# LSP TODO Analysis

## Summary
The LSP implementation for Heph has numerous TODOs indicating areas for improvement, ranging from parser upgrades and symbol indexing to capability implementations and proper shutdown handling. Key themes include transitioning to Starlark grammar, improving symbol management with trees/graphs, implementing missing LSP features, and enhancing initialization and query mechanisms.

## Major TODO Categories

### Parser and Grammar
- **Switch to Starlark tree-sitter**: Current Python grammar is inaccurate for BUILD files; Starlark grammar would provide better parsing.
- **Parser reuse**: Investigate reusing parsers instead of creating new ones.
- **Global parser usage**: See if a global parser instance can be used.
- **Remove class extraction**: Since Heph/Starlark has no classes, remove unnecessary class parsing.

### Symbol Extraction and Indexing
- **Symbol indexing**: Implement efficient indexing for symbols (currently linear search).
- **Tree structure**: Consider tree-based storage for symbols instead of flat arrays.
- **Prefix tree**: Use a trie for symbol lookup by FullyQualifiedName.
- **Target extraction**: Parse target names from BUILD files for better navigation.
- **Builtin loading**: Load builtins from hbuiltin package instead of embedded Python files.
- **FullyQualifiedName testing**: Verify extraction of qualified names like `heph.path.cwd`.

### Document and Load Management
- **Graph-based loads**: Use a graph to track load relationships instead of flat lists.
- **DAG integration**: Integrate Heph's DAG for understanding imports and dependencies.
- **Load argument handling**: Support multiple arguments in load statements.
- **Inline extraction**: Combine symbol and load extraction into single pass.

### LSP Capabilities
- **Missing features**: Implement code lens, workspace operations (file rename), diagnostics.
- **Completion improvements**: Add CompletionItemResolve, better context-aware completion.
- **Navigation scoping**: Limit declarations/definitions/references to loaded files.
- **Signature help**: Understand when signature help is triggered.

### Initialization and Sync
- **Full workspace indexing**: Load all BUILD files at startup for complete symbol index.
- **Sync robustness**: Fix panics on invalid syntax, correct edit calculations, add tests.

### Query and Navigation
- **Graph-based navigation**: Need parsed Heph tree/graph for resolving complex references.
- **Bidirectional relationships**: Track which files load others for scoped searches.

### Infrastructure
- **Interface definitions**: Better define LSP server interfaces.
- **Protocol abstraction**: Remove direct protocol usage, use raw values.
- **Concurrent safety**: Use sync.Map for document storage.
- **Proper shutdown**: Implement clean LSP server shutdown.
- **Logging**: Integrate with Heph's default logger.

## Implementation Priorities
1. **High**: Switch to Starlark parser, implement symbol indexing/tree.
2. **Medium**: Add missing capabilities (diagnostics, code lens), improve navigation scoping.
3. **Low**: Interface cleanup, logging integration, parser optimizations.

## Dependencies
Many TODOs are interconnected - e.g., graph/tree structures are needed for proper navigation, which requires DAG integration and target extraction.</content>
<parameter name="filepath">lsp_todos.md