# TODOs

- [X] Lets start by using drivers
- [X] Read `heph.fn.call()` and so on
- [X] Heph: driver should be inserted at load time
- [X] Scope Variables
- [X] Scope Variables in function call

## Features

- [ ] Hover `heph.fn.call()` and so on — dotted-path builtins like `heph.pkg.addr` can't be resolved in hover; the hierarchy walker matches only exact `sym.Name` against the last segment. Fix hover.go to resolve dotted-path lookups through `FindSymbol` with the full dotted name.
- [ ] Use chain in capabilities — `SymbolHierarchy` builds a parent chain but capabilities (hover, completion) don't use it for scope-aware resolution. Should walk the chain to respect shadowing/scoping. See document.go TODO.
- [ ] Read structs — No `processClass()` in `query.go`. `ClassKind`/`StructKind`/`FieldKind` exist in `symbol.go` but are never produced by the query engine. Full plan in `STRUCT_TYPE_SYSTEM_PLAN.md`.
- [ ] Accept struct and see how to work with symbols there — Once `processClass()` extracts structs, need type-aware completion for `RunConfig(...)` constructor calls and `param: RunConfig` type annotations. Blocked by "Read structs".
- [ ] Complete completion based on new hierarchy and symbols — Completion should use `SymbolHierarchy` (scope chain) to offer contextually appropriate symbols instead of flat symbol list. Filter by scope depth.
- [ ] Permit load loads any public var — Currently `load()` only tracks imported names. Should allow loading any public var from another BUILD file into current scope.
- [ ] Import symbols (not only functions) from `load(...)` — Currently `load()` only imports function definitions. Should also import variables, structs, and other symbol types from loaded BUILD files.
- [ ] Read all files of current workdir — `manager.go`'s `loadDocumentsFromLoads` only loads files referenced by `load()`. Should scan workspace for all BUILD files to provide cross-file symbol resolution.
- [ ] Maybe have a package type to use heph.var.b — Would need a virtual `package` type symbol per document so users can reference `heph.var.b` style identifiers.
- [ ] Driver completion — When cursor inside `target(driver="")`, show available driver names as completion items. See TODO in completion.go.
- [ ] Member/attribute completion after dot — e.g. `heph.pkg.` should complete sub-attributes. Currently attribute access is only treated as a single call name.
- [ ] Signature help — All parameter info is available but no `textDocument/signatureHelp` handler exists to show parameter info as the user types.

## Config & Project Awareness

- [ ] Load hephconfig to really see stuff — Engine config would tell the LSP about actual driver names, available plugins, project settings. Need to expose config through `Engine` → `Manager` → capabilities.
- [ ] Read the yaml to know if it will always be BUILD — Need to read hephconfig/yaml to determine the BUILD file name (might not always be `BUILD`). Affects file discovery in `manager.go`.

## Robustness

- [ ] All uint conversions should be validated before, use some common function — Multiple places convert `uint32` from tree-sitter to `int`/`uint` without bounds checking. Create a common `safeUint()` or `treeSitterUint()` function.
- [ ] Add `textDocument/didClose` handler — Documents are never cleaned up on close, stays in `Manager.DocumentMap` indefinitely.
- [ ] Diagnostics — `diagnostic.go` is empty (stub). No syntax errors, type errors, or unknown-driver warnings are reported.
- [ ] `extractTargets2()` async without completion signaling — Runs in goroutine but queries may happen before driver schemas are ready, returning incomplete target info.
- [ ] Incremental parsing wasted — `didChange` calls `doc.Tree.Edit()` then immediately `SwapTree()` does full re-parse. The old tree edit is wasted work.
- [ ] UTF-16 handling — Comments indicate LSP spec requires UTF-16 but document stores byte slices as UTF-8. `didChange` has TODO about multi-byte character handling.

## Missing LSP Capabilities

- [ ] `textDocument/documentSymbol` — No outline view support.
- [ ] `textDocument/codeLens` — Commented out; would show full target addresses.
- [ ] `textDocument/rename` — No rename support despite `Manager.RenameDocument` existing for file-level renames.
- [ ] `textDocument/formatting` — No formatting support.
- [ ] `textDocument/semanticTokens` — No semantic highlighting.
- [ ] `textDocument/inlayHint` — No inline value hints.
- [ ] `textDocument/documentHighlight` — No highlight-on-cursor.
- [ ] `textDocument/foldingRange` — No folding support.
- [ ] CompletionItemKind mapping incomplete — `FieldKind`, `ValueKind`, `PrimitiveKind`, `FunctionCallKind`, `TargetCallKind`, `RootKind` all fall through to `Value`. Should map to proper kinds.


## Future

1. Change how we parse stuff. We should probably parse scope by scope, so we can have all data first:
   - Variables -> fun def -> fun call
   - This way we recursively go from scope from scope knowing current (and outer) scope variables and function definitions
2. Instead of fallback to first BUILD in directory, we can search for that function definition